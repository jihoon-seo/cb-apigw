module github.com/cloud-barista/cb-apigw/restapigw

go 1.13

require (
	contrib.go.opencensus.io/exporter/jaeger v0.1.0
	github.com/Azure/azure-sdk-for-go v36.1.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/aws/aws-sdk-go v1.25.32 // indirect
	github.com/bramvdbogaerde/go-scp v0.0.0-20191005185035-c96fe084709e // indirect
	github.com/cloud-barista/cb-log v0.0.0-20190829061936-c402c97c951a
	github.com/cloud-barista/cb-spider v0.0.0-20191112051624-b4ed1ced75b9 // indirect
	github.com/cloud-barista/cb-store v0.0.0-20191106041549-06b54d823c60 // indirect
	github.com/coreos/etcd v3.3.17+incompatible // indirect
	github.com/gin-contrib/sse v0.0.0-20170109093832-22d885f9ecc7 // indirect
	github.com/gin-gonic/gin v1.1.5-0.20170702092826-d459835d2b07
	github.com/google/uuid v1.1.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/influxdata/influxdb v1.7.8
	github.com/labstack/echo v3.3.10+incompatible // indirect
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/racker/perigee v0.1.0 // indirect
	github.com/rackspace/gophercloud v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.2.0
	github.com/snowzach/rotatefilehook v0.0.0-20180327172521-2f64f265f58c // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/unrolled/secure v1.0.4
	github.com/valyala/fasttemplate v1.1.0 // indirect
	github.com/xujiajun/nutsdb v0.4.0 // indirect
	go.etcd.io/etcd v3.3.17+incompatible // indirect
	go.opencensus.io v0.22.1
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5 // indirect
	golang.org/x/net v0.0.0-20190921015927-1a5e07d1ff72 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	gopkg.in/yaml.v3 v3.0.0-20190924164351-c8b7dadae555
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
