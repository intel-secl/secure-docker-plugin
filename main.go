/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package main

import (
	"flag"
	"log"
	"os/user"
	"strconv"

	"secure-docker-plugin/v2/plugin"

	"github.com/docker/go-plugins-helpers/authorization"
)

const (
	defaultDockerHost = "unix:///var/run/docker.sock"
	pluginSocket      = "/run/docker/plugins/secure-docker-plugin.sock"
)

func recovery() {

	if r := recover(); r != nil {
		log.Println("Recovered:", r)
	}
}

func main() {
	flDockerHost := flag.String("host", defaultDockerHost, "Specifies the host where docker is running")
	flag.Parse()

	defer recovery()

	// Create plugin instance
	plugin, err := plugin.NewPlugin(*flDockerHost)
	if err != nil {
		log.Fatal(err)
	}

	// Start service handler on the local sock
	u, _ := user.Lookup("root")
	gid, _ := strconv.Atoi(u.Gid)
	handler := authorization.NewHandler(plugin)
	if err := handler.ServeUnix(pluginSocket, gid); err != nil {
		log.Fatal(err)
	}

	wait := make(chan struct{})
	<-wait
}
