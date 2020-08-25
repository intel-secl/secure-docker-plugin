package integrity

/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

import (
	"log"
	"os/exec"
	"strings"

	"secure-docker-plugin/v3/util"
)

const (
	dockerPullCmd            = "docker pull"
	dockerContentTrust       = "export DOCKER_CONTENT_TRUST=1"
	dockerContentTrustServer = dockerContentTrust + "; export DOCKER_CONTENT_TRUST_SERVER="
	imageNameShaPrefix       = "sha256:"
	imageTagShaSeparator     = "@sha256:"
)

// getImageName returns the image name and tag for a conatiner image
func getImageName(imageRef string) (string, error) {

	imageMetadata, err := util.GetImageMetadata(imageRef)
	if err != nil {
		log.Println("Couldn't retrieve image metadata.", err)
		return "", err
	}

	if imageMetadata == nil {
		return imageRef, nil
	}

	return imageMetadata.RepoTags[0], nil
}

// VerifyIntegrity is used for veryfing signature with notary server
func VerifyIntegrity(notaryServerURL, imageRef string) bool {

	if notaryServerURL == "" {
		log.Println("Notary URL is not specified in flavor.")
		return false
	}

	// Kubelet passes along image references as sha sums
	// we need to convert these back to readable names to proceed further
	if strings.HasPrefix(imageRef, imageNameShaPrefix) {
		image, err := getImageName(imageRef)
		if err != nil {
			log.Println("Error retrieving the image name and tag.", err)
			return false
		}
		imageRef = image
	}

	// Make sense of the image reference
	registryAddr, imageName, tag, err := util.GetRegistryAddr(imageRef)
	if err != nil {
		log.Println("Failed in parsing Registry Address from Image reference.", err, imageRef)
		return false
	}

	finalImageRef := ""

	// Handling use case where the tag name is specified with the sha256sum,
	// with this the tag parsed by GetRegistryAddr is blank,
	// we pass it as is in the form: registry:[port]/imagename@sha256:shasum
	if strings.Contains(imageRef, imageTagShaSeparator) {
		finalImageRef = registryAddr + "/" + imageName + imageTagShaSeparator + strings.Split(imageRef, imageTagShaSeparator)[1]
	} else {
		finalImageRef = registryAddr + "/" + imageName + ":" + tag
	}

	trustPullCmd := dockerContentTrustServer + notaryServerURL + ";" + dockerPullCmd + " " + finalImageRef
	log.Println("Docker trusted pull command: ", trustPullCmd)

	trustPullCmdOut, trustPullCmdErr := exec.Command("bash", "-c", trustPullCmd).Output()
	log.Println("Trusted pull returned: ", string(trustPullCmdOut))

	// Was there an error? if yes, then assume not trusted and don't allow
	if trustPullCmdErr != nil {
		log.Println("Trust Inspect returned error: ", trustPullCmdErr.Error())
		return false
	}

	return true
}
