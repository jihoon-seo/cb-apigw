module github.com/cloud-barista/cb-apigw/restapigw

go 1.13

require (
	contrib.go.opencensus.io/exporter/jaeger v0.1.0
	github.com/cloud-barista/cb-log v0.0.0-20190829061936-c402c97c951a
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/cosiner/argv v0.0.1 // indirect
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/devopsfaith/krakend v1.1.1 // indirect
	github.com/gin-gonic/gin v1.1.5-0.20170702092826-d459835d2b07
	github.com/go-delve/delve v1.4.0 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/influxdata/influxdb v1.7.8
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/peterh/liner v1.2.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/snowzach/rotatefilehook v0.0.0-20180327172521-2f64f265f58c // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.4.0
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	github.com/unrolled/secure v1.0.4
	go.opencensus.io v0.22.1
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	golang.org/x/arch v0.0.0-20200312215426-ff8b605520f4 // indirect
	golang.org/x/sys v0.0.0-20200317113312-5766fd39f98d // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 v3.0.0-20191120175047-4206685974f2
)

replace (
	github.com/cloud-barista/cb-apigw/restapigw => /Users/morris/Workspaces/ETRI/sources/cb-apigw/restapigw
	github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
)
