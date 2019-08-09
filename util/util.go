package util

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"log"
	"os/exec"
	"strings"

	"intel/isecl/lib/flavor"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
	"github.com/google/uuid"
)

type securityMetaData struct {
	RequiresConfidentiality bool
	RequiresIntegrity       bool
	KeyHandle               string
	KeySize                 string
	KeyType                 string
	CryptCipher             string
	KeyFilePath             string
	IsEmptyLayer            bool
	IsSecurityTransformed   bool
}

// GetUUIDFromImageID is used to convert image id into uuid format
func GetUUIDFromImageID(imageID string) string {

	var NameSpaceDNS = uuid.Must(uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	imageUUID := uuid.NewHash(md5.New(), NameSpaceDNS, []byte(imageID), 4)
	return imageUUID.String()
}

// GetImageFlavor is used to retrieve image flavor from Workload Agent
func GetImageFlavor(imageUUID string) (flavor.SignedImageFlavor, error) {

	var signedFlavor flavor.SignedImageFlavor
	log.Println(imageUUID)
	f, err := exec.Command("wlagent", "fetch-flavor", imageUUID, "CONTAINER_IMAGE").Output()
	if err != nil {
		log.Println("Unable to fetch image flavor from the workload agent - %v", err)
		return signedFlavor, err
	}

	json.Unmarshal(f, &signedFlavor)
	return signedFlavor, nil
}

// GetImageRef returns the image reference for a container image
func GetImageRef(req authorization.Request) string {

	var config container.Config
	json.Unmarshal(req.RequestBody, &config)
	return config.Image
}

// GetImageMetadata returns the image metadata for a container image
func GetImageMetadata(imageRef string) (*types.ImageInspect, error) {

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	imageInspect, _, err := cli.ImageInspectWithRaw(context.Background(), imageRef)
	if err != nil {
		log.Println("Couldn't retrieve image metadata from name.", err)
		return nil, nil
	}

	return &imageInspect, nil
}

// GetImageID returns the image id for a container image
func GetImageID(imageRef string) (string, error) {

	imageMetadata, err := GetImageMetadata(imageRef)
	if err != nil {
		log.Println("Couldn't retrieve image metadata.", err)
		return "", err
	}

	if imageMetadata == nil {
		//Get the sha of an image using docker api
		digest, err := GetDigestFromRegistry(imageRef)
		if err != nil {
			log.Println("Couldn't retrieve image digest from registry.", err)
			return "", err
		}
		return strings.Split(digest, ":")[1], nil
	}

	return strings.Split(imageMetadata.ID, ":")[1], nil
}

func getAPITagFromNamedRef(ref reference.Named) string {

	ref = reference.TagNameOnly(ref)
	if tagged, ok := ref.(reference.Tagged); ok {
		return tagged.Tag()
	}

	return ""
}

// GetRegistryAddr parses the image name from container create request,
// returns image, tag and registry address
func GetRegistryAddr(imageRef string) (string, string, string, error) {

	ref, err := reference.ParseNormalizedNamed(imageRef)
	if err != nil {
		return "", "", "", err
	}

	return reference.Domain(ref), reference.Path(ref), getAPITagFromNamedRef(ref), nil
}

// GetSecurityMetaData gets data related to container confidentiality and integrity by providing image ID
/* Sample security meta data JSON:
{
	"RequiresConfidentiality": true,
	"RequiresIntegrity": false,
	"KeyHandle": "f2c863b9-8fa7-4831-98cd-3cdd82567243",
	"KeySize": "256",
	"KeyType": "key-type-keyrings",
	"CryptCipher": "aes-xts-plain",
	"KeyFilePath": "/tmp/wrappedKey_f2c863b9-8fa7-4831-98cd-3cdd82567243_694390072",
	"IsEmptyLayer": false,
	"IsSecurityTransformed": true
}
*/
func GetSecurityMetaData(imageID string) (*securityMetaData, error) {

	var secData securityMetaData
	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	imageInfo, _, err := client.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		return nil, err
	}
	securityMetaData, _ := json.Marshal(imageInfo.GraphDriver.Data["security-meta-data"])
	replacer := strings.NewReplacer("\\", "", "\"{", "{", "}\"", "}")
	err = json.Unmarshal([]byte(replacer.Replace(string(securityMetaData))), &secData)
	if err != nil {
		return nil, err
	}
	return &secData, nil
}
