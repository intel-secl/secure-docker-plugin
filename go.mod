module secure-docker-plugin/v3

require (
	github.com/buger/jsonparser v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v19.03.13
	github.com/docker/go-plugins-helpers v0.0.0-20181025120712-1e6269c305b8
	github.com/google/uuid v1.1.1
	github.com/pkg/errors v0.9.1
	gopkg.in/retry.v1 v1.0.3
	intel/isecl/lib/flavor/v3 v3.6.0
	intel/isecl/lib/platform-info/v3 v3.6.0
	intel/isecl/lib/vml/v3 v3.6.0
)

replace (
	github.com/intel-secl/intel-secl/v3 => github.com/intel-secl/intel-secl/v3 v3.6.0
	intel/isecl/lib/common/v3 => github.com/intel-secl/common/v3 v3.6.0
	intel/isecl/lib/flavor/v3 => github.com/intel-secl/flavor/v3 v3.6.0
	intel/isecl/lib/platform-info/v3 => github.com/intel-secl/platform-info/v3 v3.6.0
	intel/isecl/lib/vml/v3 => github.com/intel-secl/volume-management-library/v3 v3.6.0
)
