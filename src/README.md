# Oh Dear data source for Grafana

[![Grafana plugin](https://img.shields.io/badge/Grafana-plugin-blue)](https://grafana.com/plugins/softwarehuset-ohdear-datasource)
[![CI](https://img.shields.io/github/actions/workflow/status/Softwarehuset-com/oh-dear-grafana-plugin/ci.yml)](https://github.com/Softwarehuset-com/oh-dear-grafana-plugin/actions)

Visualize your [Oh Dear](https://ohdear.app) monitoring data in Grafana: uptime, downtime, HTTP/ping/TCP response timings and Lighthouse scores, for a single monitor or every monitor in your account.

![Oh Dear sample dashboard in Grafana](https://raw.githubusercontent.com/Softwarehuset-com/oh-dear-grafana-plugin/main/src/img/screenshot-dashboard.png)

| | |
| --- | --- |
| ![Uptime](https://raw.githubusercontent.com/Softwarehuset-com/oh-dear-grafana-plugin/main/src/img/screenshot-uptime.png) | ![HTTP metrics](https://raw.githubusercontent.com/Softwarehuset-com/oh-dear-grafana-plugin/main/src/img/screenshot-http-metrics.png) |
| ![Monitors](https://raw.githubusercontent.com/Softwarehuset-com/oh-dear-grafana-plugin/main/src/img/screenshot-monitors.png) | |

## Requirements

- An [Oh Dear](https://ohdear.app) account and an API token (Oh Dear → Settings → API tokens)
- Grafana 12.3 or later

## Configuration

Add the data source, paste your API token and click **Save & test**. The token is stored encrypted and is only ever used by the plugin backend. The API URL only needs changing if you are pointing at a different Oh Dear endpoint than `https://ohdear.app/api`.

## Queries

| Query type   | Returns                                                                 |
| ------------ | ----------------------------------------------------------------------- |
| Monitors     | Table of all monitors with type, group, URL, status and last check time |
| Uptime       | Uptime percentage over time                                             |
| Downtime     | Table of downtime periods with duration and notes                       |
| HTTP metrics | DNS, TCP, SSL handshake, server processing, download and total timings  |
| Ping metrics | Min/max/average round trip time, packet loss and uptime                 |
| TCP metrics  | Time to connect and uptime                                              |
| Lighthouse   | Performance, accessibility, best practices and SEO scores plus timings  |

Leave the monitor picker on **All monitors** to get one series per monitor. Time based query types can be grouped by hour, day or month.

## Sample dashboard

The repository provisions a sample dashboard that queries every monitor in your account. See [provisioning/README.md](https://github.com/Softwarehuset-com/oh-dear-grafana-plugin/blob/main/provisioning/README.md) for how to run it.

## Learn more

- [Oh Dear API documentation](https://ohdear.app/docs/integrations/the-oh-dear-api)

## Disclaimer and license

This is an independent community plugin. It is not affiliated with, endorsed by, or sponsored by Oh Dear (ohdear.app). The plugin is provided "as is", without warranty of any kind. It is licensed under the [MIT License](https://github.com/Softwarehuset-com/oh-dear-grafana-plugin/blob/main/LICENSE).
