# Oh Dear data source for Grafana

Visualize your [Oh Dear](https://ohdear.app) monitoring data in Grafana: uptime, downtime, HTTP/ping/TCP response timings and Lighthouse scores, for a single monitor or every monitor in your account.

## Requirements

- An Oh Dear account and an API token (ohdear.app → Settings → API tokens)
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

## Learn more

- [Oh Dear API documentation](https://ohdear.app/docs/integrations/the-oh-dear-api)
