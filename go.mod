module secure-docker-plugin/v3

require (
	github.com/coreos/go-systemd v0.0.0-20190212144455-93d5ec2c7f76 // indirect
	github.com/buger/jsonparser v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v19.03.13
	github.com/docker/go-plugins-helpers v0.0.0-20181025120712-1e6269c305b8
	github.com/google/uuid v1.1.0
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	intel/isecl/lib/platform-info/v3 v3.6.0
	intel/isecl/lib/vml/v3 v3.6.0
	intel/isecl/lib/common/v3 v3.6.0
	intel/isecl/lib/flavor/v3 v3.6.0
)

replace (
	intel/isecl/lib/common/v3 => gitlab.devtools.intel.com/sst/isecl/lib/common.git/v3 v3.6/develop
	intel/isecl/lib/flavor/v3 => gitlab.devtools.intel.com/sst/isecl/lib/flavor.git/v3 v3.6/develop
	intel/isecl/lib/platform-info/v3 => gitlab.devtools.intel.com/sst/isecl/lib/platform-info.git/v3 v3.6/develop
	intel/isecl/lib/vml/v3 => gitlab.devtools.intel.com/sst/isecl/lib/volume-management.git/v3 v3.6/develop
)
