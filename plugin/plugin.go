package plugin

import (
	"log"
	"net/url"
	"strings"

	"secure-docker-plugin/integrity"
	"secure-docker-plugin/keyfetch"
	"secure-docker-plugin/util"

	dockerapi "github.com/docker/docker/api"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
)

const (
	containerCreateURI = "/containers/create"
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

	log.Println("AuthZReq Request URL path ", reqURL.Path)

	// Checking reqURL Path for the request type
	// If request type is not /containers/create, then passthrough the request
	if !strings.HasSuffix(reqURL.Path, containerCreateURI) {
		// Passthrough request - not a run request
		return authorization.Response{Allow: true}
	}

	// Request path contains /containers/create request so request body will be parsed
	// Extract image reference from request
	imageRef := util.GetImageRef(req)

	// Kubelet refers to images in the local cache by the sha256 hash,
	// so we need to convert this back to the image name before proceeding
	/*imageName, err := util.GetImageName(imageRef)
	if err != nil {
		log.Println("Error retrieving the image name and tag - %v", err)
		return authorization.Response{Allow: false}
	}*/

	// Image ID is needed to fetch image flavor
	/*imageID, err := util.GetImageID(imageRef)
	if err != nil {
		log.Println("Error retrieving the image id - %v", err)
		return authorization.Response{Allow: false}
	}*/

	// Convert image ref into uuid format
	imageUUID := util.GetUUIDFromImageRef(imageRef)

	// Get Image flavor
	flavor, err := util.GetImageFlavor(imageUUID)
	if err != nil {
		log.Println("Error retrieving the image flavor - %v", err)
		return authorization.Response{Allow: false}
	}

	if flavor.Image.Meta.ID == "" {
		log.Println("Flavor does not exist for the image ", imageUUID)
		return authorization.Response{Allow: true}
	}

	integritych := make(chan bool)
	keyfetchch := make(chan bool)

	integrityVerified := true
	keyFetched := true

	integrityRequired := flavor.Image.IntegrityEnforced
	keyfetchRequired := flavor.Image.EncryptionRequired

	if integrityRequired {
		notaryURL := flavor.Image.Integrity.NotaryURL
		go integrity.VerifyIntegrity(notaryURL, imageRef, integritych)
	}

	if keyfetchRequired {
		keyURL := flavor.Image.Encryption.KeyURL
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

	// Allowed by default.
	return authorization.Response{Allow: true}
}
