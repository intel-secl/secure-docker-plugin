/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package util

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"secure-docker-plugin/v3/config"
	"github.com/buger/jsonparser"
)

const (
	version           = "/v2/"
	defaultDockerHub  = "https://registry-1.docker.io"
	dockerHubAuthHead = "https://auth.docker.io/token?scope=repository:"
	dockerHubAuthTail = ":pull&service=registry.docker.io"
)

type requestOptions struct {
	url       string
	auth      string
	headers   map[string]string
	transport *http.Transport
}

// GetDigestFromRegistry to get sha value of an docker image from secure,insecure private and public docker registry...
func GetDigestFromRegistry(imageRef string) (string, error) {
	var (
		regAddress string
		image      string
		tag        string
		data       []byte
		token      string
		err        error
	)

	regAddress = ""
	token = ""
	regAddress, image, tag, err = GetRegistryAddr(imageRef)
	if err != nil {
		return "", err
	}

	ev := &config.EnvVariables{}
	ev.GetEnv()
	username, password, scheme, skipVerify := ev.Username, ev.Password, ev.SchemeType, ev.SkipVerify
        
	options := requestOptions{}

//	options.headers = map[string]string{"Accept": "application/vnd.docker.distribution.manifest.v2+json"}
        token, err = getToken(username, password, image)
        if err != nil {
                 return "", err
        }
	if regAddress == "docker.io" {

		options.url = defaultDockerHub + version + image + "/manifests/" + tag
	} else {
		if username != "" && password != "" {
			options.auth = getAuth(username, password)
		}

		options.url = scheme + "://" + regAddress + version + image + "/manifests/" + tag
	}
        options.headers = map[string]string{"Accept": "application/vnd.docker.distribution.manifest.v2+json", "Authorization": "Bearer " + token}
        options.transport = transport(skipVerify)
	data, err = newRequest(options)
	if err != nil {
		return "", err
	}

	digest, _, _, err := jsonparser.Get(data, "config", "digest")
	if err != nil {
		log.Println("Error in parsing digest from json", err)
		return "", err
	}

	return string(digest), nil
}

func getToken(username, password, image string) (string, error) {

	options := requestOptions{}
	options.url = dockerHubAuthHead + image + dockerHubAuthTail
	if username != "" && password != "" { 
        	options.auth = getAuth(username, password)
        }
	data, err := newRequest(options)
	if err != nil {
		return "", err
	}

	token, _, _, err := jsonparser.Get(data, "token")
	if err != nil {
		return "", err
	}

	return string(token), nil
}

func getAuth(username, password string) string {

	auth := username + ":" + password
	return auth
}


func transport(skipVerify bool) *http.Transport {
        tr := &http.Transport{
                        TLSClientConfig: &tls.Config{
                                InsecureSkipVerify: skipVerify,
                        },
                        Proxy: http.ProxyFromEnvironment,
                }
	return tr
}

func newRequest(options requestOptions) ([]byte, error) {

	var client *http.Client
	request, err := http.NewRequest("GET", options.url, nil)
	if err != nil {
		return nil, err
	}

	if options.auth != "" {
		part := strings.Split(options.auth, ":")
		request.SetBasicAuth(part[0], part[1])
	}

	for k, v := range options.headers {
		request.Header.Set(k, v)
	}

	if options.transport != nil {
		client = &http.Client{
			Transport: options.transport,
		}
	} else {
		client = &http.Client{}
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		derr := response.Body.Close()
		if derr != nil {
			log.Println("Error while closing response body", derr)
		}
	}()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
