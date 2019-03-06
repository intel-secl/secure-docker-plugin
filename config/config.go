package config

import (
	"os"
	"strconv"
)

// EnvVariables is a struct containing data required for contacting docker registry server
type EnvVariables struct {
	Username   string
	Password   string
	SchemeType string
	SkipVerify bool
}

// GetEnv will read environment variables from securedockerplugin.conf if set else set default values for variables...
func (ev *EnvVariables) GetEnv() {
	insecureSkipVerify := ""

	ev.Username = os.Getenv("REGISTRY_USERNAME")
	ev.Password = os.Getenv("REGISTRY_PASSWORD")

	ev.SchemeType = os.Getenv("REGISTRY_SCHEME_TYPE")
	if ev.SchemeType == "" {
		ev.SchemeType = "https"
	}

	insecureSkipVerify = os.Getenv("INSECURE_SKIP_VERIFY")
	if len(insecureSkipVerify) == 0 {
		//if variable is not set in env default value false will be set
		insecureSkipVerify = "false"
	}
	ev.SkipVerify, _ = strconv.ParseBool(insecureSkipVerify)
}
