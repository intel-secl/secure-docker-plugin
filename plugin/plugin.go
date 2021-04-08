/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package plugin

import (
	"encoding/json"
	"log"
	"net"
	"net/rpc"
	"net/url"
	"regexp"
	"secure-docker-plugin/v3/integrity"
	"secure-docker-plugin/v3/util"
	"strings"
	"time"

	pinfo "intel/isecl/lib/platform-info/v3/platforminfo"
	"intel/isecl/lib/vml/v3"

	"context"
	dockerapi "github.com/docker/docker/api"
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
	client *dockerclient.Client
}

type ManifestString struct {
	Manifest string
}

// NewPlugin creates a new instance of the secure docker plugin
func NewPlugin(dockerHost string) (*SecureDockerPlugin, error) {
	client, err := dockerclient.NewClient(dockerHost, dockerapi.DefaultVersion, nil, nil)
	if err != nil {
		return nil, err
	}

	return &SecureDockerPlugin{client: client}, nil
}

// AuthZReq acts only on image run (containers/create) requests.
// Remaining requests are passed through by default.
func (plugin *SecureDockerPlugin) AuthZReq(req authorization.Request) authorization.Response {
	//Parse request and the request body
	reqURI, _ := url.QueryUnescape(req.RequestURI)
	reqURL, _ := url.ParseRequestURI(reqURI)

	// Checking reqURL Path for the request type
	// If request type is not /containers/create, then passthrough the request
	if !strings.HasSuffix(reqURL.Path, containerCreateURI) {
		// Passthrough request - not a run request
		return authorization.Response{Allow: true}
	}

	// Request path contains /containers/create request so request body will be parsed
	// Extract image reference from request
	imageRef := util.GetImageRef(req)

	// Image ID is needed to fetch image flavor
	imageID, err := util.GetImageID(imageRef)
	if err != nil {
		log.Println("Error retrieving the image id.", err)
		return authorization.Response{Allow: false}
	}

	// Convert image id into uuid format
	imageUUID := util.GetUUIDFromImageID(imageID)

	// Get Image flavor
	flavor, err := util.GetImageFlavor(imageUUID)
	if err != nil {
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

		integrityVerified = integrity.VerifyIntegrity(notaryURL, imageRef)
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
	reqURI, _ := url.QueryUnescape(req.RequestURI)
	reqURL, _ := url.ParseRequestURI(reqURI)

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
			createTrustReport(containerID)
		}
	}

	// Allowed by default.
	return authorization.Response{Allow: true}
}

func isValidContainerID(containerID string) bool {
	r := regexp.MustCompile(regexForContainerIDValidation)
	return r.MatchString(containerID)
}

func createTrustReport(containerID string) {
	encrypted := false
	integrityEnforced := false
	client, err := dockerclient.NewEnvClient()
	if err != nil {
		log.Println("Failed to get docker client: ", err.Error())
	}

	instanceInfo, err := client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.Println("Failed to get instance info: ", err.Error())
	}
	imageID := strings.TrimPrefix(instanceInfo.Image, "sha256:")
	securityMetaData, err := util.GetSecurityMetaData(imageID)
	if err != nil {
		log.Println("Error getting security meta data: ", err)
	}
	if securityMetaData != nil && securityMetaData.RequiresConfidentiality == true {
		encrypted = true
	}
	// get host hardware UUID
	log.Println("Retrieving host hardware UUID...")
	hardwareUUID, err := pinfo.HardwareUUID()
	if err != nil {
		log.Println("Unable to get the host hardware UUID")
	}
	containerUUID := util.GetUUIDFromImageID(containerID)
	imageUUID := util.GetUUIDFromImageID(imageID)
	log.Println("The host hardware UUID is :", hardwareUUID)
	log.Println("Container id : ", containerID)
	manifest, err := vml.CreateContainerManifest(containerUUID, hardwareUUID, imageUUID, encrypted, integrityEnforced)
	if err != nil {
		log.Println("Unable to create manifest")
	}
	manifestByte, err := json.Marshal(manifest)
	log.Println("Manifest: ", string(manifestByte))
	if err != nil {
		log.Printf("Failed to dial workload-agent wlagent.sock %s", err.Error())
		return
	}

	timeOut, err := time.ParseDuration(util.WlagentSocketDialTimeout)
	if err != nil {
		log.Printf("Error while parsing time %s", err.Error())
		return
	}

	conn, err := net.DialTimeout("unix", util.WlagentSocketFile, timeOut)
	if err != nil {
		log.Printf("Failed to dial workload-agent wlagent.sock %s", err.Error())
		return
	}
	rpcClient := rpc.NewClient(conn)
	defer rpcClient.Close()
	var args = ManifestString{
		Manifest: string(manifestByte),
	}
	var status bool
	err = rpcClient.Call("VirtualMachine.CreateInstanceTrustReport", &args, &status)
	if err != nil {
		log.Println("Failed to create image trust report: ", err.Error())
	}
}
