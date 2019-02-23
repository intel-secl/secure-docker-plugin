package integrity

import (
	"log"
	"os/exec"
	"strings"

	"github.com/docker/distribution/reference"
)

const (
	dockerPullCmd            = "docker pull"
	dockerContentTrust       = "export DOCKER_CONTENT_TRUST=1"
	dockerContentTrustServer = dockerContentTrust + "; export DOCKER_CONTENT_TRUST_SERVER="
	imageTagShaSeparator     = "@sha256:"
)

func getAPITagFromNamedRef(ref reference.Named) string {

	ref = reference.TagNameOnly(ref)
	if tagged, ok := ref.(reference.Tagged); ok {
		return tagged.Tag()
	}

	return ""
}

// getRegistryAddr parses the image name from container create request,
// returns image, tag and registry address
func getRegistryAddr(imageRef string) (string, string, string, error) {

	ref, err := reference.ParseNormalizedNamed(imageRef)
	if err != nil {
		return "", "", "", err
	}

	return reference.Domain(ref), reference.Path(ref), getAPITagFromNamedRef(ref), nil
}

// VerifyIntegrity is used for veryfing signature with notary server
func VerifyIntegrity(notaryServerURL, imageRef string, integritych chan bool) {

	if notaryServerURL == "" {
		log.Println("Notary URL is not specified in flavor.")
		integritych <- false
		return
	}

	// Make sense of the image reference
	registryAddr, imageName, tag, err := getRegistryAddr(imageRef)
	if err != nil {
		log.Println("Failed in parsing Registry Address from Image reference.", err, imageRef)
		integritych <- false
		return
	}

	finalImageRef := ""

	// Hnadling use case where the tag name is specified with the sha256sum,
	// with this the tag parsed by getImageRef is blank,
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
		integritych <- false
		return
	}

	integritych <- true
}
