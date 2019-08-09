module secure-docker-plugin

require (
	github.com/Appboy/worker-pools v0.0.0-20190116202358-b3004bb29cb5
	github.com/Sirupsen/logrus v1.3.0 // indirect
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23
	github.com/coreos/go-systemd v0.0.0-20190212144455-93d5ec2c7f76 // indirect
	github.com/diegobernardes/ttlcache v0.0.0-20190225130438-dcd0e0a06d7f // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-plugins-helpers v0.0.0-20181025120712-1e6269c305b8
	github.com/docker/go-units v0.3.3 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/google/uuid v1.1.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	intel/isecl/lib/common v0.0.0
	intel/isecl/lib/flavor v0.0.0
	intel/isecl/lib/platform-info v0.0.0
	intel/isecl/lib/vml v0.0.0
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.3.0

	intel/isecl/lib/common => gitlab.devtools.intel.com/sst/isecl/lib/common.git v0.0.0-20190807232609-b7b7cd045b1c

	intel/isecl/lib/flavor => gitlab.devtools.intel.com/sst/isecl/lib/flavor.git v0.0.0-20190808115139-3baf114b27c1

	intel/isecl/lib/platform-info => gitlab.devtools.intel.com/sst/isecl/lib/platform-info.git v0.0.0-20190724162312-1400c15ab9ee

	intel/isecl/lib/vml => gitlab.devtools.intel.com/sst/isecl/lib/volume-management.git v0.0.0-20190724045147-f917328f3b16
)
