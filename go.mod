module secure-docker-plugin/v3

require (
	github.com/coreos/go-systemd v0.0.0-20190212144455-93d5ec2c7f76 // indirect
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v19.03.13
	github.com/docker/go-plugins-helpers v0.0.0-20181025120712-1e6269c305b8
	github.com/google/uuid v1.1.0
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	intel/isecl/lib/platform-info/v3 v3.3.0
	intel/isecl/lib/vml/v3 v3.3.0
	intel/isecl/lib/common/v3 v3.3.0
	intel/isecl/lib/flavor/v3 v3.3.0
)

replace (
	intel/isecl/lib/common/v3 => github.com/intel-secl/common/v3 v3.3.0
	intel/isecl/lib/flavor/v3 => github.com/intel-secl/flavor/v3 v3.3.0
	intel/isecl/lib/platform-info/v3 => github.com/intel-secl/platform-info/v3 v3.3.0
	intel/isecl/lib/vml/v3 => github.com/intel-secl/volume-management-library/v3 v3.3.0
)
