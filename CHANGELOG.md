# Changelog

## 1.0.1 (2026-10-10)

Dependency and tooling updates:

- Update `@grafana/*` packages from 13.1.0 to 13.2.3
- Upgrade react and react-dom from 18 to 19 (required by `@grafana/ui` 13.2.x)
- Pin eslint to 9.x: `eslint-plugin-react` (latest 7.37.5) is not compatible with eslint 10
- Update GitHub Actions: `golangci-lint-action` v9.3.0, `e2e-version` v3.1.3, `wait-for-grafana` v1.0.5, release workflow Go 1.27
- Update `grafana-plugin-sdk-go` to v0.297.0

## 1.0.0 (2026-10-10)

Initial release.

- Query monitors, uptime, downtime, HTTP/ping/TCP response metrics and Lighthouse reports from the Oh Dear API
- Query a single monitor or every monitor in the account, grouped by hour, day or month
- API token stored as encrypted secure json data and verified by the health check
