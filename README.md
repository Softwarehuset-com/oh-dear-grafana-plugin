# Oh Dear data source for Grafana

A Grafana data source plugin for [Oh Dear](https://ohdear.app). It queries the Oh Dear API from the plugin backend and exposes monitors, uptime, downtime, HTTP/ping/TCP response metrics and Lighthouse reports as Grafana data frames.

See [src/README.md](./src/README.md) for the user facing documentation.

## Layout

- `pkg/plugin/ohdear_client.go` — Oh Dear API client (monitors, uptime, downtime, metrics, Lighthouse, `/me`)
- `pkg/plugin/datasource.go` — query handling, health check and the `/monitors` resource used by the query editor
- `pkg/models/settings.go` — data source settings, with the API token in `secureJsonData` as `ohDearToken`
- `src/components/` — config and query editors

## Development

```bash
npm install
npm run dev          # frontend in watch mode
mage -v              # backend binaries
npm run server       # Grafana in Docker with the plugin mounted
```

Provide a token for the provisioned data source through the `OHDEAR_API_TOKEN` environment variable.

```bash
npm run typecheck
npm run lint
npm run test:ci      # jest
go test ./pkg/...
npm run e2e          # playwright, needs npm run server
```

Changes to `src/plugin.json` require a restart of the Grafana server.

## Signing and releasing

Signing and release are handled by the GitHub workflows in `.github/workflows`. See the Grafana documentation on [publishing a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin).
