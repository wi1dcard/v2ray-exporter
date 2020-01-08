module github.com/wi1dcard/v2ray-exporter

go 1.13

require (
	github.com/gogo/protobuf v1.1.1
	github.com/golang/protobuf v1.3.2
	github.com/jessevdk/go-flags v1.4.0
	github.com/prometheus/client_golang v1.3.0
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/grpc v1.26.0
	v2ray.com/core v4.19.1+incompatible
)

replace v2ray.com/core => github.com/v2ray/v2ray-core v4.22.1+incompatible
