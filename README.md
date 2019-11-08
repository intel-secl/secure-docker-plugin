##### Intel<sup>®</sup> Security Libraries for Data Center  - Secure Docker Plugin
Secure Docker Plugin (SDP) allows the secure docker daemon to run according to the policy set in the container image flavor. 
The SDP is created based on the authorization plugin architecture supported by the docker engine

## System Requirements
- RHEL 7.5/7.6
- Epel 7 Repo
- Proxy settings if applicable

## Software requirements
- git
- makeself
- `go` version >= `go1.11.4` & <= `go1.12.12`

### Install `go` version >= `go1.11.4` & <= `go1.12.12`
The `secure-docker-plugin` requires Go version 1.11.4 that has support for `go modules`. The build was validated with the latest version 1.12.12 of `go`. It is recommended that you use 1.12.12 version of `go`. More recent versions may introduce compatibility issues. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.12.12.linux-amd64.tar.gz
tar -xzf go1.12.12.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

### Build
```console
> make
```

### Deploy
The Secure Docker Plugin is bundled with ISecL workload agent, is depolyed when workload agent is installed with container security.

### Manage service
* Start service
    * systemctl start secure-docker-plugin
* Stop service
    * systemctl stop secure-docker-plugin
* Restart service
    * systemctl restart secure-docker-plugin
* Status of service
    * systemctl status secure-docker-plugin

