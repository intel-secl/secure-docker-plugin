DESCRIPTION="ISecL K8S Docker Secure Plugin"
SERVICE=secure-docker-plugin
SERVICEINSTALLDIR=/usr/libexec
SERVICESOCKETFILE=${SERVICE}.socket
SERVICECONFIGFILE=${SERVICE}.service
SYSTEMINSTALLDIR=/lib/systemd/system
SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

VERSION := v0.0.0
GITCOMMIT := $(shell git describe --always)
GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD)
TIMESTAMP := $(shell date --iso=seconds)

# LDFLAGS
LDFLAGS=-ldflags "-X main.Version=$(VERSION)-$(GITCOMMIT) -X main.Branch=$(GITBRANCH) -X main.Time=$(TIMESTAMP)"

# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE): $(SOURCES)
	go build ${LDFLAGS} -o ${SERVICE}

# Generate the service config and socket files
.PHONY: config
config: $(SERVICESOCKETFILE) $(SERVICECONFIGFILE)

# Generate the socket file
.PHONY: $(SERVICESOCKETFILE)
$(SERVICESOCKETFILE): 
	@echo -n "" > ${SERVICESOCKETFILE}
	@echo "[Unit]" >> ${SERVICESOCKETFILE}
	@echo "Description=${DESCRIPTION} Socket" >> ${SERVICESOCKETFILE}
	@echo >> ${SERVICESOCKETFILE}
	@echo "[Socket]" >> ${SERVICESOCKETFILE}
	@echo "ListenStream=/run/docker/plugins/${SERVICE}.sock" >> ${SERVICESOCKETFILE}
	@echo >> ${SERVICESOCKETFILE}
	@echo "[Install]" >> ${SERVICESOCKETFILE}
	@echo "WantedBy=sockets.target" >> ${SERVICESOCKETFILE}

# Generate the service file
.PHONY: $(SERVICECONFIGFILE)
$(SERVICECONFIGFILE):
	@echo -n "" > ${SERVICECONFIGFILE}
	@echo "[Unit]" >> ${SERVICECONFIGFILE}
	@echo "Description=${DESCRIPTION}" >> ${SERVICECONFIGFILE}
	@echo "Before=docker.service" >> ${SERVICECONFIGFILE}
	@echo "After=network.target ${SERVICESOCKETFILE}" >> ${SERVICECONFIGFILE}
	@echo "Requires=${SERVICESOCKETFILE} docker.service" >> ${SERVICECONFIGFILE}
	@echo  >> ${SERVICECONFIGFILE}
	@echo "[Service]" >> ${SERVICECONFIGFILE}
	@echo "ExecStart=${SERVICEINSTALLDIR}/${SERVICE}" >> ${SERVICECONFIGFILE}
	@echo  >> ${SERVICECONFIGFILE}
	@echo "[Install]" >> ${SERVICECONFIGFILE}
	@echo "WantedBy=multi-user.target" >> ${SERVICECONFIGFILE}

# Install the service binary and the service config files
.PHONY: install
install:
	@cp -f ${SERVICE} ${SERVICEINSTALLDIR}
	@cp -f ${SERVICESOCKETFILE} ${SYSTEMINSTALLDIR}
	@cp -f ${SERVICECONFIGFILE} ${SYSTEMINSTALLDIR}

# Uninstalls the service binary and the service config files
.PHONY: uninstall
uninstall:
	@rm -f ${SERVICEINSTALLDIR}/${SERVICE}
	@rm -f ${SYSTEMINSTALLDIR}/${SERVICESOCKETFILE}
	@rm -f ${SYSTEMINSTALLDIR}/${SERVICECONFIGFILE}

# Removes the generated service config and binary files
.PHONY: clean
clean:
	@rm -rf src/github.com src/golang.org
	@rm -rf pkg/ bin/
	@rm -f ${SERVICE}
	@rm -f ${SERVICESOCKETFILE}
	@rm -f ${SERVICECONFIGFILE}
