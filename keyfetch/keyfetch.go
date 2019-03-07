package keyfetch

import (
	"log"
	"os/exec"
	"strings"
	"time"

	pool "github.com/Appboy/worker-pools"
)

type keyInfo struct {
	keyURL    string
	imageUUID string
}

// Global channel to communicate between secure-docker-plugin and CacheKey goroutine.
var ch = make(chan keyInfo, 1)

func getKeyHandle(keyURL string) string {

	keyURLSplit := strings.Split(keyURL, "/")
	keyID := keyURLSplit[len(keyURLSplit)-2]
	return keyID
}

// cacheDecryptionKey is used to cache decryption key on Workload Agent
func cacheDecryptionKey(imageUUID, keyHandle string) error {

	_, err := exec.Command("wlagent", "cache-key", imageUUID, keyHandle).Output()
	if err != nil {
		log.Println("Unable to cache decryption key on the workload agent", err)
		return err
	}
	return nil
}

// CacheKey is used to call workload agent to fetch the kmskey and unwrap the key and store actual key  in kernel keyring
func CacheKey() {
	var (
		keyHandle string
		err       error
	)
	maxConcurrentWorkloads := 2
	stalePoolExpiration := 5 * time.Minute
	maxPoolLifetime := 10 * time.Minute

	// workerpool manager is initialized
	poolManager := pool.NewWorkerPoolManager(
		maxConcurrentWorkloads, stalePoolExpiration, maxPoolLifetime,
	)
	for {
		pool, doneUsing := poolManager.GetPool("pool 1", 1)
		defer close(doneUsing)
		pool.Submit(func() {
			// Do anything here. Only maxConcurrentWorkloads will be allowed to execute concurrently per pool.
			// This is useful for limiting concurrent usage of external resources.

			// waiting for image id and key url from the plugin
			keyInfo := <-ch

			keyHandle = getKeyHandle(keyInfo.keyURL)
			if keyHandle == "" {
				log.Println("Unable to get keyHandle from key url", keyInfo.keyURL)
			}

			if keyHandle != "" {
				err = cacheDecryptionKey(keyInfo.imageUUID, keyHandle)
				if err != nil {
					log.Println("Unable to store the key in the keyring", err)
				}
				log.Println("Key is present in keyring", keyHandle)
			}
		})
	}
}

// FetchKey is used to fetch the decryption key from workload agent and store wrapped key in kernel keyring
func FetchKey(keyURL, imageUUID string, keyfetchch chan bool) {

	if keyURL == "" {
		log.Println("Key URL is not specified in flavor.")
		keyfetchch <- false
		return
	}

	var ki keyInfo
	ki.keyURL = keyURL
	ki.imageUUID = imageUUID

	// Triggering the goroutine by passing image id and key url on channel
	ch <- ki

	keyfetchch <- true
}
