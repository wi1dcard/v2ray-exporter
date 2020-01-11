# V2Ray Exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/wi1dcard/v2ray-exporter)][goreportcard]
[![Build Status](https://travis-ci.com/wi1dcard/v2ray-exporter.svg?branch=master)][build-status]

A Exporter that collect V2Ray metrics over its [Stats API][stats-api] and export them to Prometheus.

![grafana.png][grafana-screenshot]

[stats-api]: https://www.v2ray.com/chapter_02/stats.html
[goreportcard]: https://goreportcard.com/report/github.com/wi1dcard/v2ray-exporter
[grafana-screenshot]: https://i.loli.net/2020/01/09/uXrgSG4VaiJqCIE.png
[build-status]: https://travis-ci.com/wi1dcard/v2ray-exporter

## Installation

The latest binaries are made available on GitHub [releases][github-releases] page:

```bash
wget -O /tmp/v2ray-exporter https://github.com/wi1dcard/v2ray-exporter/releases/latest/download/v2ray-exporter_linux_amd64
mv /tmp/v2ray-exporter /usr/local/bin/v2ray-exporter
chmod +x /usr/local/bin/v2ray-exporter
```

You can also find the docker images built automatically by CI from [Docker Hub](https://hub.docker.com/r/wi1dcard/v2ray-exporter). The images are made for multi-arch. You can run it from your Raspberry Pi or any other ARM, ARM64 devices without changing the image name:

```bash
docker run --rm -it wi1dcard/v2ray-exporter:<TAG>
```

Please note that `latest` tag is not available. Use `master` instead if you want the latest build of master branch.

## Usage

Before we start, let's assume you have already set up Prometheus and Grafana.

Firstly, you will need to make sure the API and statistics related features has been enabled in your V2Ray config file. For example:

```json
{
    "stats": {},
    "api": {
        "tag": "api",
        "services": [
            "StatsService"
        ]
    },
    "policy": {
        "levels": {
            "0": {
                "statsUserUplink": true,
                "statsUserDownlink": true
            }
        },
        "system": {
            "statsInboundUplink": true,
            "statsInboundDownlink": true
        }
    },
    "inbounds": [
        {
            "tag": "tcp",
            "port": 12345,
            "protocol": "vmess",
            "settings": {
                "clients": [
                    {
                        "email": "foo@example.com",
                        "id": "e731f153-4f31-49d3-9e8f-ff8f396135ef",
                        "level": 0,
                        "alterId": 4
                    },
                    {
                        "email": "bar@example.com",
                        "id": "e731f153-4f31-49d3-9e8f-ff8f396135ee",
                        "level": 0,
                        "alterId": 4
                    }
                ]
            }
        },
        {
            "tag": "api",
            "listen": "127.0.0.1",
            "port": 54321,
            "protocol": "dokodemo-door",
            "settings": {
                "address": "127.0.0.1"
            }
        }
    ],
    "outbounds": [
        {
            "protocol": "freedom",
            "settings": {}
        }
    ],
    "routing": {
        "rules": [
            {
                "inboundTag": [
                    "api"
                ],
                "outboundTag": "api",
                "type": "field"
            }
        ]
    }
}
```

You can see we open two inbounds in the configuration above. The first inbound accepts VMess connections from user `foo@example.com` and `bar@example.com`, and the second one listens port 54321 on localhost and handles the API calls, which is the endpoint of the exporter scrapes. If you'd like to run V2Ray and exporter in different machine, consider use `0.0.0.0` instead of `127.0.0.1` and be careful with security risks.

Additionally, we should also enable `stats`, `api`, `policy` and setup proper routing rules in order to get traffic statistics works. For more information, please see [The Beginner's Guide of V2Ray][v2ray-beginners-guide].

The next step is to start the exporter:

```bash
v2ray-exporter --v2ray-endpoint "127.0.0.1:54321"
## Or
docker run --rm -d wi1dcard/v2ray-exporter --v2ray-endpoint "127.0.0.1:54321"
```

The logs should looks like:

```
V2Ray Exporter v0.2.0-110e82d (built 2020-01-09T16:07:56Z)
INFO[0000] Server is ready to handle incoming scrape requests.
```

Now the exporter is listening on `:9550`, and the option `--listen` allows you changing the listen address or port.

You can now open `http://IP:9550` in your browser:

![browser.png][browser-screenshot]

Click the `Scrape V2Ray Metrics`, the exporter will expose all metrics including V2Ray runtime and statistics in the Prometheus metrics format, like:

```
...
# HELP v2ray_up Indicate scrape succeeded or not
# TYPE v2ray_up gauge
v2ray_up 1
# HELP v2ray_uptime_seconds V2Ray uptime in seconds
# TYPE v2ray_uptime_seconds gauge
v2ray_uptime_seconds 150624
...
```

If you can't find `v2ray_up 1` in the response which means the scrape was failed, please review the logs (stdout/stderr) of V2Ray Exporter for more detailed information.

We have the metrics exposed. Now let Prometheus scrapes these data points and visualize them with Grafana. Here is an example Promtheus configuration:

```yaml
global:
  scrape_interval: 15s
  scrape_timeout: 5s

scrape_configs:
  - job_name: v2ray
    metrics_path: /scrape
    static_configs:
      - targets: [IP:9550]
```

To learn more about Prometheus, please see the [official docs][prometheus-docs].

A simple Grafana dashboard is also available [here][grafana-dashboard]. Please refer to the [Grafana docs][grafana-importing-dashboard] to get the steps about importing dashboards form JSON files.

## Digging Deeper

We did not keep the original metrci names from V2Ray intentionally. You may find out why in the [comments][explaination-of-metric-names].

For users who do not really care about internal changes but only need a mapping table, here you got it:

| Runtime Metric   | Exposed Metric                     |
| :--------------- | :--------------------------------- |
| `uptime`         | `v2ray_uptime_seconds`             |
| `num_goroutine`  | `v2ray_goroutines`                 |
| `alloc`          | `v2ray_memstats_alloc_bytes`       |
| `total_alloc`    | `v2ray_memstats_alloc_bytes_total` |
| `sys`            | `v2ray_memstats_sys_bytes`         |
| `mallocs`        | `v2ray_memstats_mallocs_total`     |
| `frees`          | `v2ray_memstats_frees_total`       |
| `live_objects`   | Removed. See the appendix below.   |
| `num_gc`         | `v2ray_memstats_num_gc`            |
| `pause_total_ns` | `v2ray_memstats_pause_total_ns`    |

| Statistic Metric                          | Exposed Metric                                                              |
| :---------------------------------------- | :-------------------------------------------------------------------------- |
| `inbound>>>tag-name>>>traffic>>>uplink`   | `v2ray_traffic_uplink_bytes_total{dimension="inbound",target="tag-name"}`   |
| `inbound>>>tag-name>>>traffic>>>downlink` | `v2ray_traffic_downlink_bytes_total{dimension="inbound",target="tag-name"}` |
| `user>>>user-email>>traffic>>>uplink`     | `v2ray_traffic_uplink_bytes_total{dimension="user",target="user-email"}`    |
| `user>>>user-email>>>traffic>>>downlink`  | `v2ray_traffic_downlink_bytes_total{dimension="user",target="user-email"}`  |
| ...                                       | ...                                                                         |

- The value of `live_objects` can be calculated by using `memstats_mallocs_total - memstats_frees_total`.

## TODOs

- GitHub Action

## Special Thanks

- <https://github.com/schweikert/fping-exporter>
- <https://github.com/oliver006/redis_exporter>
- <https://github.com/roboll/helmfile>

## License

MIT

[github-releases]: https://github.com/wi1dcard/v2ray-exporter/releases
[v2ray-beginners-guide]: https://guide.v2fly.org/en_US/advanced/traffic.html
[browser-screenshot]: https://i.loli.net/2020/01/11/ZVtNEU8iqMrFGKm.png
[prometheus-docs]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/
[grafana-dashboard]: ./dashboard.json
[grafana-importing-dashboard]: https://grafana.com/docs/grafana/latest/reference/export_import/#importing-a-dashboard
[explaination-of-metric-names]: https://github.com/wi1dcard/v2ray-exporter/blob/110e82dfefb1b51f4da3966ddd1945b5d0dac203/exporter.go#L134
