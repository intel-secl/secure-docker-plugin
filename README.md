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
- Go 11.4 or newer

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

