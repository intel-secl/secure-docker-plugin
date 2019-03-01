package util

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"log"
	"os/exec"

	"intel/isecl/lib/flavor"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
	"github.com/google/uuid"
)

const (
	imageNameShaPrefix = "sha256:"
)

// GetUUIDFromImageRef is used to convert image id into uuid format
func GetUUIDFromImageRef(imageRef string) string {

	var NameSpaceDNS = uuid.Must(uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	imageUUID := uuid.NewHash(md5.New(), NameSpaceDNS, []byte(imageRef), 4)
	return imageUUID.String()
}

// GetImageFlavor is used to retrieve image flavor from Workload Agent
func GetImageFlavor(imageUUID string) (flavor.ImageFlavor, error) {

	var flavor flavor.ImageFlavor

	f, err := exec.Command("wlagent", "fetch-flavor", imageUUID, "CONTAINER_IMAGE").Output()
	if err != nil {
		log.Println("Unable to fetch image flavor from the workload agent - %v", err)
		return flavor, err
	}

	json.Unmarshal(f, &flavor)
	return flavor, nil
}

// GetImageRef returns the image reference for a container image
func GetImageRef(req authorization.Request) string {

	var config container.Config
	json.Unmarshal(req.RequestBody, &config)
	return config.Image
}

// GetImageName returns the image name and tag for a conatiner image
func GetImageName(image string) (string, error) {

	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	imageInspect, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		log.Println("Couldn't retrieve image name from sha - %v", err)
		return "", err
	}

	log.Println("Extracted image name ", imageInspect.RepoTags[0])
	return imageInspect.RepoTags[0], nil
}

// GetImageID retruns the image id for a container image
func GetImageID(image string) (string, error) {

	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	imageInspect, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		log.Println("Couldn't retrieve image sha from name - %v", err)
		return "", err
	}

	log.Println("Extracted image hash ", imageInspect.ID)
	return imageInspect.ID, nil
}
