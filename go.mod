module github.com/tetratelabs/getmesh

go 1.16

require (
	cloud.google.com/go/security v1.1.0
	github.com/Masterminds/semver v1.5.0
	github.com/aws/aws-sdk-go v1.42.6
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/kiali/kiali v1.43.0
	github.com/manifoldco/promptui v0.9.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211116061358-0a5406a5449c
	google.golang.org/genproto v0.0.0-20211116182654-e63d96a377c4
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	istio.io/pkg v0.0.0-20211112173506-bca79fa408e9
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
)

replace github.com/tetratelabs/getmesh => ./

replace github.com/kiali/kiali => github.com/kiali/kiali v1.29.1-0.20210125202741-72d2ce2fceb5
