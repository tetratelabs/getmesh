module github.com/tetratelabs/getmesh

go 1.16

require (
	cloud.google.com/go v0.74.0
	github.com/Azure/go-autorest/autorest v0.11.15 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.10 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/aws/aws-sdk-go v1.36.20
	github.com/elazarl/goproxy v0.0.0-20201021153353-00ad82a08272 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kiali/kiali v1.29.1-0.20210125202741-72d2ce2fceb5
	github.com/kr/pretty v0.2.1 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20201214210602-f9fddec55a1e
	google.golang.org/genproto v0.0.0-20201210142538-e3217bee35cc
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
	istio.io/pkg v0.0.0-20201113155828-7cd7caba2919
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
)

replace github.com/kiali/kiali => github.com/kiali/kiali v1.29.1-0.20210125202741-72d2ce2fceb5
