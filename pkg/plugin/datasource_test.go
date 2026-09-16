package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/softwarehuset/oh-dear/pkg/models"
)

const monitorsPayload = `{
  "data": [
    {"id": 1, "type": "http", "url": "https://example.com", "label": "Example", "group_name": "Web", "latest_run_date": "2026-09-16 08:00:00", "summarized_check_result": "succeeded"},
    {"id": 2, "type": "http", "url": "https://api.example.com", "label": "API", "group_name": "Web", "latest_run_date": "2026-09-16 08:01:00", "summarized_check_result": "failed"}
  ],
  "links": {"next": null}
}`

func newTestDatasource(t *testing.T, handler http.HandlerFunc) (*Datasource, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	settings := &models.PluginSettings{
		BaseURL: server.URL,
		Secrets: &models.SecretPluginSettings{Token: "test-token"},
	}

	return &Datasource{settings: settings, client: NewClient(settings)}, server
}

func timeRange() backend.TimeRange {
	return backend.TimeRange{
		From: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC),
	}
}

func queryJSON(t *testing.T, qm queryModel) backend.DataQuery {
	t.Helper()

	raw, err := json.Marshal(qm)
	if err != nil {
		t.Fatal(err)
	}
	return backend.DataQuery{RefID: "A", JSON: raw, TimeRange: timeRange()}
}

func TestQueryDataUnknownQueryType(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{queryJSON(t, queryModel{QueryType: "nope"})},
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.Responses["A"].Error == nil {
		t.Fatal("expected an error for an unknown query type")
	}
}

func TestQueryMonitors(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/monitors" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("unexpected authorization header %q", got)
		}
		_, _ = w.Write([]byte(monitorsPayload))
	})

	resp := ds.query(context.Background(), queryJSON(t, queryModel{QueryType: QueryTypeMonitors}))
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	if len(resp.Frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(resp.Frames))
	}

	frame := resp.Frames[0]
	if frame.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", frame.Rows())
	}
	if got := frame.Fields[0].At(0); got != "Example" {
		t.Errorf("expected first monitor to be Example, got %v", got)
	}
	if got := frame.Fields[5].At(1); got != "failed" {
		t.Errorf("expected second monitor status failed, got %v", got)
	}
}

func TestQueryUptimeForAllMonitors(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/monitors":
			_, _ = w.Write([]byte(monitorsPayload))
		case "/monitors/1/uptime", "/monitors/2/uptime":
			if got := r.URL.Query().Get("split"); got != "day" {
				t.Errorf("expected split=day, got %q", got)
			}
			_, _ = w.Write([]byte(`[{"datetime": "2026-09-15 00:00:00", "uptime_percentage": 99.9}]`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})

	resp := ds.query(context.Background(), queryJSON(t, queryModel{QueryType: QueryTypeUptime, GroupBy: "day"}))
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	if len(resp.Frames) != 2 {
		t.Fatalf("expected one frame per monitor, got %d", len(resp.Frames))
	}
	if resp.Frames[0].Name != "Example" {
		t.Errorf("expected frame named after the monitor, got %q", resp.Frames[0].Name)
	}
	if got := resp.Frames[0].Fields[1].At(0); got != 99.9 {
		t.Errorf("expected uptime 99.9, got %v", got)
	}
}

func TestQueryUptimeForSingleMonitor(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/monitors/2":
			_, _ = w.Write([]byte(`{"id": 2, "label": "API", "url": "https://api.example.com"}`))
		case "/monitors/2/uptime":
			_, _ = w.Write([]byte(`[{"datetime": "2026-09-15 00:00:00", "uptime_percentage": 100}]`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})

	resp := ds.query(context.Background(), queryJSON(t, queryModel{QueryType: QueryTypeUptime, MonitorID: 2}))
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	if len(resp.Frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(resp.Frames))
	}
	if resp.Frames[0].Name != "API" {
		t.Errorf("expected frame named API, got %q", resp.Frames[0].Name)
	}
}

func TestQueryDowntime(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/monitors/1":
			_, _ = w.Write([]byte(`{"id": 1, "label": "Example"}`))
		case "/monitors/1/downtime":
			_, _ = w.Write([]byte(`{"data": [{"id": 9, "started_at": "2026-09-15 10:00:00", "ended_at": "2026-09-15 10:30:00", "notes_markdown": "outage"}]}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})

	resp := ds.query(context.Background(), queryJSON(t, queryModel{QueryType: QueryTypeDowntime, MonitorID: 1}))
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	frame := resp.Frames[0]
	if frame.Rows() != 1 {
		t.Fatalf("expected 1 row, got %d", frame.Rows())
	}
	if got := frame.Fields[3].At(0); got != float64(30) {
		t.Errorf("expected a 30 minute outage, got %v", got)
	}
}

func TestQueryHTTPMetricsConvertsSecondsToMilliseconds(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/monitors/1":
			_, _ = w.Write([]byte(`{"id": 1, "label": "Example"}`))
		case "/monitors/1/http-uptime-metrics":
			if got := r.URL.Query().Get("filter[group_by]"); got != "hour" {
				t.Errorf("expected group_by=hour, got %q", got)
			}
			_, _ = w.Write([]byte(`{"data": [{"date": "2026-09-15 10:00:00", "total_time_in_seconds": 0.25, "dns_time_in_seconds": 0.01}]}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})

	resp := ds.query(context.Background(), queryJSON(t, queryModel{QueryType: QueryTypeHTTPMetrics, MonitorID: 1}))
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	if got := resp.Frames[0].Fields[1].At(0); got != float64(250) {
		t.Errorf("expected 250 ms total time, got %v", got)
	}
}

func TestQueryReportsUnauthorized(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	resp := ds.query(context.Background(), queryJSON(t, queryModel{QueryType: QueryTypeMonitors}))
	if resp.Error == nil {
		t.Fatal("expected an error")
	}
	if resp.ErrorSource != backend.ErrorSourceDownstream {
		t.Errorf("expected a downstream error, got %v", resp.ErrorSource)
	}
}

func TestCheckHealth(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id": 1, "name": "Ada", "email": "ada@example.com"}`))
	})

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != backend.HealthStatusOk {
		t.Fatalf("expected ok, got %v: %s", result.Status, result.Message)
	}
	if result.Message != "Connected as Ada" {
		t.Errorf("unexpected message %q", result.Message)
	}
}

func TestCheckHealthWithInvalidToken(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != backend.HealthStatusError || result.Message != "Invalid API token" {
		t.Errorf("unexpected health result %v: %s", result.Status, result.Message)
	}
}

func TestCheckHealthWithoutToken(t *testing.T) {
	settings := &models.PluginSettings{BaseURL: "https://ohdear.app/api", Secrets: &models.SecretPluginSettings{}}
	ds := &Datasource{settings: settings, client: NewClient(settings)}

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Message != "API token is missing" {
		t.Errorf("unexpected message %q", result.Message)
	}
}

type resourceSender struct {
	response *backend.CallResourceResponse
}

func (s *resourceSender) Send(response *backend.CallResourceResponse) error {
	s.response = response
	return nil
}

func TestCallResourceMonitors(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(monitorsPayload))
	})

	sender := &resourceSender{}
	if err := ds.CallResource(context.Background(), &backend.CallResourceRequest{Path: "/monitors"}, sender); err != nil {
		t.Fatal(err)
	}

	if sender.response.Status != http.StatusOK {
		t.Fatalf("expected 200, got %d", sender.response.Status)
	}

	var items []map[string]any
	if err := json.Unmarshal(sender.response.Body, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(items))
	}
	if items[0]["label"] != "Example" {
		t.Errorf("unexpected first monitor %v", items[0])
	}
}

func TestCallResourceUnknownPath(t *testing.T) {
	ds, _ := newTestDatasource(t, func(w http.ResponseWriter, r *http.Request) {})

	sender := &resourceSender{}
	if err := ds.CallResource(context.Background(), &backend.CallResourceRequest{Path: "/nope"}, sender); err != nil {
		t.Fatal(err)
	}

	if sender.response.Status != http.StatusNotFound {
		t.Errorf("expected 404, got %d", sender.response.Status)
	}
}
