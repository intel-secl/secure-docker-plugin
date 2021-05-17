/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package plugin

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gopkg.in/retry.v1"
	"log"
	"net"
	"net/rpc"
	"net/url"
	"regexp"
	"secure-docker-plugin/v3/integrity"
	"secure-docker-plugin/v3/util"
	"strings"
	"sync"
	"time"

	pinfo "intel/isecl/lib/platform-info/v3/platforminfo"
	"intel/isecl/lib/vml/v3"

	"context"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
)

const (
	containerCreateURI                  = `/containers/create`
	regexForContainerIDValidation       = `^[a-fA-F0-9]{64}$`
	regexForcontainerStartURIValidation = `(.*?\bcontainers\b)(.*?\bstart\b).*`
)

// SecureDockerPlugin struct definition
type SecureDockerPlugin struct {
	// Docker client
	dockerClient *dockerclient.Client
	dockerHost   string
	// Wlagent client
	wlaClient         *rpc.Client
	wlaSocketFilePath string
	dcmtx             sync.Mutex
	wlamtx            sync.Mutex
}

type ManifestString struct {
	Manifest string
}

func (plugin *SecureDockerPlugin) closeDockerClient() {
	plugin.dcmtx.Lock()
	defer plugin.dcmtx.Unlock()
	if plugin.dockerClient != nil {
		plugin.dockerClient.Close()
	}
	plugin.dockerClient = nil
}

func (plugin *SecureDockerPlugin) closeWlaClient() {
	plugin.wlamtx.Lock()
	defer plugin.wlamtx.Unlock()
	if plugin.wlaClient != nil {
		plugin.wlaClient.Close()
	}
	plugin.wlaClient = nil
}

func (plugin *SecureDockerPlugin) getDockerClient() (*dockerclient.Client, error) {
	var err error
	plugin.dcmtx.Lock()
	defer plugin.dcmtx.Unlock()
	if plugin.dockerClient != nil {
		return plugin.dockerClient, nil
	}

	attempts := retry.Regular{
		Total: 30 * time.Second,
		Delay: 500 * time.Millisecond,
	}
	for attempt := attempts.Start(nil); attempt.Next(); {
		log.Printf("Attempt %v to initialize docker client", attempt.Count())
		plugin.dockerClient, err = dockerclient.NewClientWithOpts(dockerclient.WithHost(plugin.dockerHost),
			dockerclient.WithAPIVersionNegotiation())
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.Wrap(err, "SDP: Failed to initialize Docker engine client")
	}

	return plugin.dockerClient, nil
}

func (plugin *SecureDockerPlugin) getWlaClient() (*rpc.Client, error) {
	var err error
	plugin.wlamtx.Lock()
	defer plugin.wlamtx.Unlock()

	if plugin.wlaClient != nil {
		return plugin.wlaClient, nil
	}

	timeOut, _ := time.ParseDuration(util.WlagentSocketDialTimeout)
	attempts := retry.Regular{
		Total: 30 * time.Second,
		Delay: 500 * time.Millisecond,
	}
	var conn net.Conn
	for attempt := attempts.Start(nil); attempt.Next(); {
		log.Printf("Attempt %v to initialize WLA client", attempt.Count())
		conn, err = net.DialTimeout("unix", plugin.wlaSocketFilePath, timeOut)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "SDP: Failed to initialize WLA client")
	}

	plugin.wlaClient = rpc.NewClient(conn)
	return plugin.wlaClient, nil
}

// NewPlugin creates a new instance of the secure docker plugin
func NewPlugin(dockerHost, wlagentSocketFile string) (*SecureDockerPlugin, error) {
	sdp := &SecureDockerPlugin{
		dockerHost:        dockerHost,
		wlaSocketFilePath: wlagentSocketFile,
	}
	if _, err := sdp.getDockerClient(); err != nil {
		return nil, err
	}
	if _, err := sdp.getWlaClient(); err != nil {
		return nil, err
	}
	log.Println("SDP init OK")
	return sdp, nil
}

func (plugin *SecureDockerPlugin) Cleanup() error {
	dockerCleanupErr := plugin.dockerClient.Close()
	wlaCleanupErr := plugin.wlaClient.Close()
	return errors.Errorf("Plugin client cleanup err: Docker %v | WLA - %v",
		dockerCleanupErr, wlaCleanupErr)
}

// AuthZReq acts only on image run (containers/create) requests.
// Remaining requests are passed through by default.
func (plugin *SecureDockerPlugin) AuthZReq(req authorization.Request) authorization.Response {
	//Parse request and the request body
	reqURI, err := url.QueryUnescape(req.RequestURI)
	if err != nil {
		log.Println("Error retrieving the request URI", err)
		return authorization.Response{Allow: false}
	}
	reqURL, err := url.ParseRequestURI(reqURI)
	if err != nil {
		log.Println("Error retrieving the request URL", err)
		return authorization.Response{Allow: false}
	}

	// Checking reqURL Path for the request type
	// If request type is not /containers/create, then passthrough the request
	if !strings.HasSuffix(reqURL.Path, containerCreateURI) {
		// Passthrough request - not a run request
		return authorization.Response{Allow: true}
	}

	// Request path contains /containers/create request so request body will be parsed
	// Extract image reference from request
	imageRef := util.GetImageRef(req)

	dc, err := plugin.getDockerClient()
	if err != nil {
		log.Println("Error retrieving the image id.", err)
		plugin.closeDockerClient()
		return authorization.Response{Allow: false}
	}

	// Image ID is needed to fetch image flavor
	imageID, err := util.GetImageID(dc, imageRef)
	if err != nil {
		log.Println("Error retrieving the image id.", err)
		return authorization.Response{Allow: false}
	}

	// Convert image id into uuid format
	imageUUID := util.GetUUIDFromImageID(imageID)

	wlac, err := plugin.getWlaClient()
	if err != nil {
		if strings.Contains(err.Error(), "connection is shut down") {
			plugin.closeWlaClient()
		}
		log.Println("Error retrieving the image id.", err)
		return authorization.Response{Allow: false}
	}

	// Get Image flavor
	flavor, err := util.GetImageFlavor(wlac, imageUUID)
	if err != nil {
		if strings.Contains(err.Error(), "connection is shut down") {
			plugin.closeWlaClient()
		}
		log.Println("Error retrieving the image flavor.", err)
		return authorization.Response{Allow: false}
	}

	if flavor.Meta.ID == "" {
		log.Println("Flavor does not exist for the image.", imageUUID)
		return authorization.Response{Allow: true}
	}

	integrityVerified := true

	integrityRequired := flavor.IntegrityEnforced

	if integrityRequired {
		notaryURL := flavor.Integrity.NotaryURL
		if strings.HasSuffix(notaryURL, "/") {
			notaryURL = notaryURL[:len(notaryURL)-1]
		}

		integrityVerified = integrity.VerifyIntegrity(dc, notaryURL, imageRef)
	}

	if integrityVerified {
		log.Println("[ALLOWED]")
		return authorization.Response{Allow: true}
	}

	log.Println("[DENIED]")
	return authorization.Response{Allow: false}
}

// AuthZRes authorizes the docker client response.
// All responses are allowed by default.
func (plugin *SecureDockerPlugin) AuthZRes(req authorization.Request) authorization.Response {
	//Parse request and the request body
	reqURI, err := url.QueryUnescape(req.RequestURI)
	if err != nil {
		log.Println("AuthZRes: Error parsing req")
		return authorization.Response{Allow: false}
	}
	reqURL, err := url.ParseRequestURI(reqURI)
	if err != nil {
		log.Println("AuthZRes: Error parsing request URI")
		return authorization.Response{Allow: false}
	}

	dc, err := plugin.getDockerClient()
	if err != nil {
		plugin.closeDockerClient()
		log.Println("Error retrieving the image id.", err)
		return authorization.Response{Allow: false}
	}

	wlac, err := plugin.getWlaClient()
	if err != nil {
		if strings.Contains(err.Error(), "connection is shut down") {
			plugin.closeWlaClient()
		}
		log.Println("Error retrieving the image id.", err)
		return authorization.Response{Allow: false}
	}

	// Checking reqURL Path for the request type
	// If request type is not /containers/(id or name)/start, then passthrough the request
	r := regexp.MustCompile(regexForcontainerStartURIValidation)
	if !r.MatchString(reqURL.Path) {
		return authorization.Response{Allow: true}
	}
	if req.ResponseStatusCode == 204 {
		reqURLSplit := strings.Split(reqURL.Path, "/")
		containerID := reqURLSplit[len(reqURLSplit)-2]
		if isValidContainerID(containerID) {
			err = createTrustReport(dc, wlac, containerID)
			if err != nil && strings.Contains(err.Error(), "connection is shut down") {
				plugin.closeWlaClient()
			}
		}
	}

	// Allowed by default.
	return authorization.Response{Allow: true}
}

func isValidContainerID(containerID string) bool {
	r := regexp.MustCompile(regexForContainerIDValidation)
	return r.MatchString(containerID)
}

func createTrustReport(dc *dockerclient.Client, wlac *rpc.Client, containerID string) error {
	encrypted := false
	integrityEnforced := false
	instanceInfo, err := dc.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.Println("Failed to get instance info: ", err.Error())
	} else {
		imageID := strings.TrimPrefix(instanceInfo.Image, "sha256:")
		securityMetaData, err := util.GetSecurityMetaData(dc, imageID)
		if err != nil {
			log.Println("Error getting security meta data: ", err)
			return nil
		}
		if securityMetaData != nil {
			encrypted = securityMetaData.RequiresConfidentiality
			integrityEnforced = securityMetaData.RequiresIntegrity
		} else {
			return nil
		}
		// get host hardware UUID
		log.Println("Retrieving host hardware UUID...")
		hardwareUUID, err := pinfo.HardwareUUID()
		if err != nil {
			log.Println("Unable to get the host hardware UUID")
			return err
		}
		containerUUID := util.GetUUIDFromImageID(containerID)
		imageUUID := util.GetUUIDFromImageID(imageID)
		log.Println("The host hardware UUID is :", hardwareUUID)
		log.Println("Container id : ", containerID)
		manifest, err := vml.CreateContainerManifest(containerUUID, hardwareUUID, imageUUID, encrypted, integrityEnforced)
		if err != nil {
			log.Println("Unable to create manifest for container ", containerID)
			return err
		}
		manifestByte, err := json.Marshal(manifest)
		if err != nil {
			log.Printf("Error marshalling Manifest for container %v %v", containerID, err.Error())
			return err
		}
		log.Println("Manifest: ", string(manifestByte))

		var args = ManifestString{
			Manifest: string(manifestByte),
		}
		var status bool
		err = wlac.Call("VirtualMachine.CreateInstanceTrustReport", &args, &status)
		if err != nil {
			log.Printf("Failed to create trust report for container %v %v: ", containerID, err.Error())
			return err
		}
	}
	return nil
}
