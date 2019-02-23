package keyfetch

import (
	"log"
	"os/exec"
	"strings"
)

func getKeyHandle(keyURL string) string {

	keyURLSplit := strings.Split(keyURL, "/")
	keyID := keyURLSplit[len(keyURLSplit)-2]
	return keyID
}

// cacheDecryptionKey is used to cache decryption key on Workload Agent
func cacheDecryptionKey(imageID, keyHandle string) error {

	_, err := exec.Command("wlagent", "cache-key", imageID, keyHandle).Output()
	if err != nil {
		log.Println("Unable to cache decryption key on the workload agent", err)
		return err
	}
	return nil
}

// FetchKey is used to fetch the decryption key from workload agent and store wrapped key in kernel keyring
func FetchKey(keyURL, imageID string, keyfetchch chan bool) {

	if keyURL == "" {
		log.Println("Key URL is not specified in flavor.")
		keyfetchch <- false
		return
	}

	keyHandle := getKeyHandle(keyURL)
	err := cacheDecryptionKey(imageID, keyHandle)
	if err != nil {
		keyfetchch <- false
		return
	}

	keyfetchch <- true
}
