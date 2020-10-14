module github.com/cloud-barista/cb-apigw/restapigw

go 1.13

// CB-STORE 관련
replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.3
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	github.com/xujiajun/nutsdb v0.5.0 => github.com/xujiajun/nutsdb v0.5.1-0.20200320023740-0cc84000d103
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	contrib.go.opencensus.io/exporter/jaeger v0.1.0
	github.com/cloud-barista/cb-log v0.2.0-cappuccino.0.20201008023843-31002c0a088d
	github.com/cloud-barista/cb-store v0.2.0-cappuccino.0.20201014054737-e2310432d256
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/sse v0.0.0-20170109093832-22d885f9ecc7 // indirect
	github.com/gin-gonic/gin v1.1.5-0.20170702092826-d459835d2b07
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/influxdata/influxdb v1.7.8
	github.com/json-iterator/go v1.1.10
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.4.0
	github.com/ugorji/go v1.1.7 // indirect
	github.com/unrolled/secure v1.0.4
	go.opencensus.io v0.22.4
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)
