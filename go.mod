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

	intel/isecl/lib/common => github.com/intel-secl/common v0.0.0-20190920051932-22c16ec493a3

	intel/isecl/lib/flavor => github.com/intel-secl/flavor v0.0.0-20190915015315-7d9923b58ff3

	intel/isecl/lib/platform-info => github.com/intel-secl/platform-info v0.0.0-20190918083246-1f72bff4f238

	intel/isecl/lib/vml => github.com/intel-secl/volume-management-library v0.0.0-20190915022206-560299d2b8e9
)
