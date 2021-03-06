/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package util

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
	"github.com/google/uuid"
	"intel/isecl/lib/flavor/v3"
	"log"
	"net/rpc"
	"strings"
)

const WlagentSocketFile = "/var/run/workload-agent/wlagent.sock"
const WlagentSocketDialTimeout = "5s"

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
	imageUUID := uuid.NewHash(md5.New(), uuid.NameSpaceDNS, []byte(imageID), 4)
	return imageUUID.String()
}

type FlavorInfo struct {
	ImageID string
}

type OutFlavor struct {
	ReturnCode  bool
	ImageFlavor string
}

// GetImageFlavor is used to retrieve image flavor from Workload Agent
func GetImageFlavor(client *rpc.Client, imageUUID string) (flavor.Image, error) {

	var flvr flavor.Image
	var err error

	log.Printf("Fetching flavor for image id %s", imageUUID)

	var outFlavor OutFlavor
	var args = FlavorInfo{
		ImageID: imageUUID,
	}
	err = client.Call("VirtualMachine.FetchFlavor", &args, &outFlavor)
	if err != nil {
		log.Printf("Unable to fetch image flavor from the workload agent - %v", err)
		return flvr, err
	}
	if len(outFlavor.ImageFlavor) == 0 {
		log.Printf("There is no flavor for given image id %s", imageUUID)
		return flvr, nil
	}
	err = json.Unmarshal([]byte(outFlavor.ImageFlavor), &flvr)
	if err != nil {
		log.Printf("Unable to unmarshal image flavor - %v", err)
		return flvr, err
	}
	return flvr, nil
}

// GetImageRef returns the image reference for a container image
func GetImageRef(req authorization.Request) string {

	var config container.Config
	err := json.Unmarshal(req.RequestBody, &config)
	if err != nil {
		log.Printf("Unable to unmarshal request - %v", err)
	}
	return config.Image
}

// GetImageMetadata returns the image metadata for a container image
func GetImageMetadata(dc *client.Client, imageRef string) (*types.ImageInspect, error) {
	imageInspect, _, err := dc.ImageInspectWithRaw(context.Background(), imageRef)
	if err != nil {
		log.Printf("Couldn't retrieve image metadata from name: %v", err)
		return nil, nil
	}

	return &imageInspect, nil
}

// GetImageID returns the image id for a container image
func GetImageID(dc *client.Client, imageRef string) (string, error) {
	imageMetadata, err := GetImageMetadata(dc, imageRef)
	if err != nil {
		log.Printf("Couldn't retrieve image metadata: %v", err)
		return "", err
	}

	if imageMetadata == nil {
		//Get the sha of an image using docker api
		digest, err := GetDigestFromRegistry(imageRef)
		if err != nil {
			log.Printf("Couldn't retrieve image digest from registry: %v", err)
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
func GetSecurityMetaData(dc *client.Client, imageID string) (*securityMetaData, error) {
	var secData securityMetaData
	imageInfo, _, err := dc.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		return nil, err
	}

	if imageInfo.GraphDriver.Data["security-meta-data"] == "" {
		return nil, nil
	}
	securityMetaData, err := json.Marshal(imageInfo.GraphDriver.Data["security-meta-data"])
	if err != nil {
		return nil, err
	}
	replacer := strings.NewReplacer("\\", "", "\"{", "{", "}\"", "}")
	err = json.Unmarshal([]byte(replacer.Replace(string(securityMetaData))), &secData)
	if err != nil {
		return nil, err
	}
	return &secData, nil
}
