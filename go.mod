module terraform-provider-omni

go 1.25.0

replace (
	// forked go-yaml that introduces RawYAML interface, which can be used to populate YAML fields using bytes
	// which are then encoded as a valid YAML blocks with proper indentiation
	gopkg.in/yaml.v3 => github.com/unix4ever/yaml v0.0.0-20220527175918-f17b0f05cf2c
	// our fork for tcpproxy with fixes
	inet.af/tcpproxy => github.com/smira/tcpproxy v0.0.0-20200125044825-b6bb9b5b8252
)

require (
	github.com/cosi-project/runtime v1.11.0
	github.com/hashicorp/terraform-plugin-framework v1.15.1
	github.com/hashicorp/terraform-plugin-framework-timetypes v0.5.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.18.0
	github.com/hashicorp/terraform-plugin-go v0.28.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.37.0
	github.com/hashicorp/terraform-plugin-testing v1.13.2
	github.com/siderolabs/gen v0.8.5
	github.com/siderolabs/omni/client v1.2.1
	gopkg.in/yaml.v3 v3.0.3
)

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/ProtonMail/go-mime v0.0.0-20230322103455-7d82a3887f2f // indirect
	github.com/ProtonMail/gopenpgp/v2 v2.9.0 // indirect
	github.com/adrg/xdg v0.5.3 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/containerd/go-cni v1.1.12 // indirect
	github.com/containernetworking/cni v1.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/gertd/go-pluralize v0.2.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.26.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty v1.5.0 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.6.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/hc-install v0.9.2 // indirect
	github.com/hashicorp/hcl/v2 v2.23.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.23.0 // indirect
	github.com/hashicorp/terraform-json v0.25.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.5 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/jsimonetti/rtnetlink/v2 v2.0.5 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mdlayher/ethtool v0.4.0 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.3-0.20250113171957-fbb4dce95f42 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/onsi/ginkgo/v2 v2.21.0 // indirect
	github.com/onsi/gomega v1.35.1 // indirect
	github.com/opencontainers/runtime-spec v1.2.1 // indirect
	github.com/petermattis/goid v0.0.0-20250508124226-395b08cebbdb // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20241121165744-79df5c4772f2 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/siderolabs/crypto v0.6.3 // indirect
	github.com/siderolabs/go-api-signature v0.3.8 // indirect
	github.com/siderolabs/go-pointer v1.0.1 // indirect
	github.com/siderolabs/net v0.4.0 // indirect
	github.com/siderolabs/proto-codec v0.1.2 // indirect
	github.com/siderolabs/protoenc v0.2.3 // indirect
	github.com/siderolabs/talos/pkg/machinery v1.11.1 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/zclconf/go-cty v1.16.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/time v0.13.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)
