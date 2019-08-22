package plugin

import (
	"encoding/json"
	"log"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"secure-docker-plugin/integrity"
	"secure-docker-plugin/keyfetch"
	"secure-docker-plugin/util"

	pinfo "intel/isecl/lib/platform-info/platforminfo"
	"intel/isecl/lib/vml"

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

	if flavor.ImageFlavor.Meta.ID == "" {
		log.Println("Flavor does not exist for the image.", imageUUID)
		return authorization.Response{Allow: true}
	}

	integritych := make(chan bool)
	keyfetchch := make(chan bool)

	integrityVerified := true
	keyFetched := true

	integrityRequired := flavor.ImageFlavor.IntegrityEnforced
	keyfetchRequired := flavor.ImageFlavor.EncryptionRequired

	if integrityRequired {
		notaryURL := flavor.ImageFlavor.Integrity.NotaryURL
		go integrity.VerifyIntegrity(notaryURL, imageRef, integritych)
	}

	if keyfetchRequired {
		keyURL := flavor.ImageFlavor.Encryption.KeyURL
		go keyfetch.FetchKey(keyURL, imageUUID, keyfetchch)
	}

	if integrityRequired {
		integrityVerified = <-integritych
	}

	if keyfetchRequired {
		keyFetched = <-keyfetchch
	}

	if integrityVerified && keyFetched {
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
	var encrypted bool
	var integrityEnforced bool
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
	if securityMetaData != nil && securityMetaData.KeyHandle == "" {
		encrypted = false
	} else {
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
	_, err = exec.Command("wlagent", "create-instance-trust-report", string(manifestByte)).CombinedOutput()
	if err != nil {
		log.Println("Failed to create image trust report: ", err.Error())
	}
}
