/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package main

import (
	"flag"
	"github.com/docker/go-plugins-helpers/authorization"
	"log"
	"os/user"
	"secure-docker-plugin/v3/plugin"
	"secure-docker-plugin/v3/util"
	"strconv"
)

const (
	defaultDockerHost = "unix:///var/run/docker.sock"
	pluginSocket      = "/run/docker/plugins/secure-docker-plugin.sock"
)

var (
	// Version holds the version number for the WLA binary
	Version = ""
	// BuildDate holds the build date for the WLA binary
	BuildDate = ""
	// GitHash holds the commit hash for the WLA binary
	GitHash = ""
)

func recovery() {
	if r := recover(); r != nil {
		log.Println("Recovered:", r)
	}
}

func main() {
	log.Println("Plugin init")

	flDockerHost := flag.String("host", defaultDockerHost, "Specifies the host where docker is running")
	flag.Parse()

	defer recovery()

	// Create sdp instance
	sdp, err := plugin.NewPlugin(*flDockerHost, util.WlagentSocketFile)
	if err != nil {
		log.Fatal(err)
	}

	// Start service handler on the local sock
	u, err := user.Lookup("root")
	if err != nil {
		log.Fatal(err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		log.Fatal(err)
	}

	handler := authorization.NewHandler(sdp)
	if err := handler.ServeUnix(pluginSocket, gid); err != nil {
		log.Fatal(err)
	}

	log.Println("Plugin exit")
	err = sdp.Cleanup()
	if err != nil {
		log.Printf("%v", err)
	}
}
