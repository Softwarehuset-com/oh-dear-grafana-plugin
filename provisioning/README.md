# Provisioned test environment

This folder provisions a working test environment for the Oh Dear data source plugin, for local development and for the Grafana plugin review process.

## What is provisioned

- **Data source** `Oh Dear` (uid `ohdear`) in `datasources/datasources.yml`, pointing at `https://ohdear.app/api`
- **Sample dashboard** `Oh Dear – Sample` in `dashboards/`, querying every monitor in your account: monitors table, uptime, downtime periods, HTTP metrics and Lighthouse scores

## Requirements

- Docker
- An [Oh Dear](https://ohdear.app) account. The sample dashboard shows real data for the monitors in the account the token belongs to, so any account with at least one monitor works.

## Running it

1. Create an API token: Oh Dear → Settings → API tokens
2. Start Grafana with the plugin:

   ```bash
   OHDEAR_API_TOKEN=<your-token> npm run server
   ```

   This builds the plugin and starts the Grafana container from `docker-compose.yaml` with the plugin mounted and this folder mounted as `/etc/grafana/provisioning`.

3. Open http://localhost:3000 and select the **Oh Dear – Sample** dashboard.

The data source is editable in the UI (Connections → Data sources → Oh Dear) if you need to point it at a different Oh Dear API endpoint or replace the token.
