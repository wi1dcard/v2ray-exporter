module github.com/wi1dcard/v2ray-exporter

go 1.13

require (
	github.com/jessevdk/go-flags v1.4.0
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/grpc v1.29.1
	v2ray.com/core v4.19.1+incompatible
)

replace v2ray.com/core => github.com/v2ray/v2ray-core v1.24.5-0.20200610141238-f9935d0e93ea
