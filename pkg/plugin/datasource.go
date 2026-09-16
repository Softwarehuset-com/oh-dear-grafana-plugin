package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/Softwarehuset-com/oh-dear-grafana-plugin/pkg/models"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Query types supported by the Oh Dear data source.
const (
	// QueryTypeMonitors returns a table of all monitors and their current status.
	QueryTypeMonitors = "monitors"
	// QueryTypeUptime returns uptime percentages over time.
	QueryTypeUptime = "uptime"
	// QueryTypeDowntime returns downtime periods.
	QueryTypeDowntime = "downtime"
	// QueryTypeHTTPMetrics returns HTTP response timing metrics.
	QueryTypeHTTPMetrics = "http-uptime-metrics"
	// QueryTypePingMetrics returns ping metrics.
	QueryTypePingMetrics = "ping-uptime-metrics"
	// QueryTypeTCPMetrics returns TCP connection metrics.
	QueryTypeTCPMetrics = "tcp-uptime-metrics"
	// QueryTypeLighthouse returns Lighthouse performance scores.
	QueryTypeLighthouse = "lighthouse"
)

// Datasource is an Oh Dear datasource instance.
type Datasource struct {
	settings *models.PluginSettings
	client   *Client
}

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	parsed, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}

	return &Datasource{
		settings: parsed,
		client:   NewClient(parsed),
	}, nil
}

// Dispose tells the plugin SDK that the plugin wants to clean up resources
// when a new instance is created.
func (d *Datasource) Dispose() {
	// Nothing to clean up.
}

type queryModel struct {
	QueryType string `json:"queryType"`
	MonitorID int64  `json:"monitorId"`
	GroupBy   string `json:"groupBy"`
}

// QueryData handles multiple queries and returns multiple responses.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.query(ctx, q)
	}

	return response, nil
}

func (d *Datasource) query(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	var qm queryModel
	if err := json.Unmarshal(q.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err))
	}

	switch qm.QueryType {
	case QueryTypeMonitors:
		return d.queryMonitors(ctx)
	case QueryTypeUptime:
		return d.queryUptime(ctx, qm, q.TimeRange)
	case QueryTypeDowntime:
		return d.queryDowntime(ctx, qm, q.TimeRange)
	case QueryTypeHTTPMetrics:
		return d.queryHTTPMetrics(ctx, qm, q.TimeRange)
	case QueryTypePingMetrics:
		return d.queryPingMetrics(ctx, qm, q.TimeRange)
	case QueryTypeTCPMetrics:
		return d.queryTCPMetrics(ctx, qm, q.TimeRange)
	case QueryTypeLighthouse:
		return d.queryLighthouse(ctx, qm)
	default:
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unknown query type %q", qm.QueryType))
	}
}

// monitorTargets resolves the monitors a query applies to. When MonitorID is
// zero, all monitors in the account are used.
func (d *Datasource) monitorTargets(ctx context.Context, qm queryModel) ([]Monitor, error) {
	if qm.MonitorID > 0 {
		monitor, err := d.client.GetMonitor(ctx, qm.MonitorID)
		if err != nil {
			return nil, err
		}
		return []Monitor{*monitor}, nil
	}

	monitors, err := d.client.GetMonitors(ctx)
	if err != nil {
		return nil, err
	}
	if len(monitors) == 0 {
		return nil, errors.New("no monitors found in your Oh Dear account")
	}
	return monitors, nil
}

func (d *Datasource) queryMonitors(ctx context.Context) backend.DataResponse {
	monitors, err := d.client.GetMonitors(ctx)
	if err != nil {
		return apiErrorResponse(err)
	}

	ids := make([]int64, 0, len(monitors))
	types := make([]string, 0, len(monitors))
	labels := make([]string, 0, len(monitors))
	groups := make([]string, 0, len(monitors))
	urls := make([]string, 0, len(monitors))
	statuses := make([]string, 0, len(monitors))
	latestRuns := make([]time.Time, 0, len(monitors))

	for _, m := range monitors {
		ids = append(ids, m.ID)
		types = append(types, m.Type)
		labels = append(labels, m.Label)
		groups = append(groups, m.GroupName)
		urls = append(urls, m.URL)
		statuses = append(statuses, m.SummarizedCheckResult)

		if t, err := parseOhDearTime(m.LatestRunDate); err == nil {
			latestRuns = append(latestRuns, t)
		} else {
			latestRuns = append(latestRuns, time.Time{})
		}
	}

	frame := data.NewFrame("Monitors",
		data.NewField("Monitor", nil, labels),
		data.NewField("ID", nil, ids),
		data.NewField("Type", nil, types),
		data.NewField("Group", nil, groups),
		data.NewField("URL", nil, urls),
		data.NewField("Status", nil, statuses),
		data.NewField("Latest check", nil, latestRuns),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) queryUptime(ctx context.Context, qm queryModel, timeRange backend.TimeRange) backend.DataResponse {
	monitors, err := d.monitorTargets(ctx, qm)
	if err != nil {
		return apiErrorResponse(err)
	}

	split := qm.GroupBy
	if split == "" {
		split = "hour"
	}

	frames := make([]*data.Frame, 0, len(monitors))
	for _, m := range monitors {
		points, err := d.client.GetUptime(ctx, m.ID, timeRange.From, timeRange.To, split)
		if err != nil {
			return apiErrorResponse(err)
		}

		times := make([]time.Time, 0, len(points))
		values := make([]float64, 0, len(points))
		for _, p := range points {
			t, err := parseOhDearTime(p.Datetime)
			if err != nil {
				continue
			}
			times = append(times, t)
			values = append(values, p.UptimePercentage)
		}

		frames = append(frames, newTimeSeriesFrame(m.Label, "Uptime (%)", times, values))
	}

	return framesResponse(frames)
}

func (d *Datasource) queryDowntime(ctx context.Context, qm queryModel, timeRange backend.TimeRange) backend.DataResponse {
	monitors, err := d.monitorTargets(ctx, qm)
	if err != nil {
		return apiErrorResponse(err)
	}

	monitorNames := make([]string, 0)
	starts := make([]time.Time, 0)
	ends := make([]time.Time, 0)
	durations := make([]float64, 0)
	notes := make([]string, 0)

	for _, m := range monitors {
		records, err := d.client.GetDowntime(ctx, m.ID, timeRange.From, timeRange.To)
		if err != nil {
			return apiErrorResponse(err)
		}

		for _, r := range records {
			start, err1 := parseOhDearTime(r.StartedAt)
			end, err2 := parseOhDearTime(r.EndedAt)
			if err1 != nil || err2 != nil {
				continue
			}
			monitorNames = append(monitorNames, m.Label)
			starts = append(starts, start)
			ends = append(ends, end)
			durations = append(durations, end.Sub(start).Seconds()/60)
			if r.NotesMarkdown != "" {
				notes = append(notes, r.NotesMarkdown)
			} else {
				notes = append(notes, "")
			}
		}
	}

	frame := data.NewFrame("Downtime",
		data.NewField("Monitor", nil, monitorNames),
		data.NewField("Started at", nil, starts),
		data.NewField("Ended at", nil, ends),
		data.NewField("Duration (min)", nil, durations),
		data.NewField("Notes", nil, notes),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) queryHTTPMetrics(ctx context.Context, qm queryModel, timeRange backend.TimeRange) backend.DataResponse {
	monitors, err := d.monitorTargets(ctx, qm)
	if err != nil {
		return apiErrorResponse(err)
	}

	groupBy := defaultGroupBy(qm.GroupBy, "hour")

	frames := make([]*data.Frame, 0, len(monitors))
	for _, m := range monitors {
		points, err := d.client.GetHTTPUptimeMetrics(ctx, m.ID, timeRange.From, timeRange.To, groupBy)
		if err != nil {
			return apiErrorResponse(err)
		}

		times := make([]time.Time, 0, len(points))
		totals := make([]float64, 0, len(points))
		dns := make([]float64, 0, len(points))
		tcp := make([]float64, 0, len(points))
		ssl := make([]float64, 0, len(points))
		server := make([]float64, 0, len(points))
		download := make([]float64, 0, len(points))

		for _, p := range points {
			t, err := parseOhDearTime(p.Date)
			if err != nil {
				continue
			}
			times = append(times, t)
			totals = append(totals, p.TotalTimeInSeconds*1000)
			dns = append(dns, p.DNSTimeInSeconds*1000)
			tcp = append(tcp, p.TCPTimeInSeconds*1000)
			ssl = append(ssl, p.SSLHandshakeTimeInSeconds*1000)
			server = append(server, p.RemoteServerProcessingTimeInSeconds*1000)
			download = append(download, p.DownloadTimeInSeconds*1000)
		}

		frames = append(frames, newMultiValueFrame(m.Label, "HTTP timing (ms)", len(monitors) > 1, times,
			"Total", totals,
			"DNS", dns,
			"TCP", tcp,
			"SSL handshake", ssl,
			"Server processing", server,
			"Download", download,
		))
	}

	return framesResponse(frames)
}

func (d *Datasource) queryPingMetrics(ctx context.Context, qm queryModel, timeRange backend.TimeRange) backend.DataResponse {
	monitors, err := d.monitorTargets(ctx, qm)
	if err != nil {
		return apiErrorResponse(err)
	}

	groupBy := defaultGroupBy(qm.GroupBy, "hour")

	frames := make([]*data.Frame, 0, len(monitors))
	for _, m := range monitors {
		points, err := d.client.GetPingUptimeMetrics(ctx, m.ID, timeRange.From, timeRange.To, groupBy)
		if err != nil {
			return apiErrorResponse(err)
		}

		times := make([]time.Time, 0, len(points))
		avg := make([]float64, 0, len(points))
		mins := make([]float64, 0, len(points))
		maxs := make([]float64, 0, len(points))
		packetLoss := make([]float64, 0, len(points))
		uptime := make([]float64, 0, len(points))

		for _, p := range points {
			t, err := parseOhDearTime(p.Date)
			if err != nil {
				continue
			}
			times = append(times, t)
			avg = append(avg, p.AverageTimeInMS)
			mins = append(mins, p.MinimumTimeInMS)
			maxs = append(maxs, p.MaximumTimeInMS)
			packetLoss = append(packetLoss, p.PacketLossPercent)
			uptime = append(uptime, p.UptimePercentage)
		}

		frames = append(frames, newMultiValueFrame(m.Label, "Ping metrics", len(monitors) > 1, times,
			"Average RTT (ms)", avg,
			"Min RTT (ms)", mins,
			"Max RTT (ms)", maxs,
			"Packet loss (%)", packetLoss,
			"Uptime (%)", uptime,
		))
	}

	return framesResponse(frames)
}

func (d *Datasource) queryTCPMetrics(ctx context.Context, qm queryModel, timeRange backend.TimeRange) backend.DataResponse {
	monitors, err := d.monitorTargets(ctx, qm)
	if err != nil {
		return apiErrorResponse(err)
	}

	groupBy := defaultGroupBy(qm.GroupBy, "hour")

	frames := make([]*data.Frame, 0, len(monitors))
	for _, m := range monitors {
		points, err := d.client.GetTCPUptimeMetrics(ctx, m.ID, timeRange.From, timeRange.To, groupBy)
		if err != nil {
			return apiErrorResponse(err)
		}

		times := make([]time.Time, 0, len(points))
		connect := make([]float64, 0, len(points))
		uptime := make([]float64, 0, len(points))

		for _, p := range points {
			t, err := parseOhDearTime(p.Date)
			if err != nil {
				continue
			}
			times = append(times, t)
			connect = append(connect, p.TimeToConnectInMS)
			uptime = append(uptime, p.UptimePercentage)
		}

		frames = append(frames, newMultiValueFrame(m.Label, "TCP metrics", len(monitors) > 1, times,
			"Time to connect (ms)", connect,
			"Uptime (%)", uptime,
		))
	}

	return framesResponse(frames)
}

func (d *Datasource) queryLighthouse(ctx context.Context, qm queryModel) backend.DataResponse {
	monitors, err := d.monitorTargets(ctx, qm)
	if err != nil {
		return apiErrorResponse(err)
	}

	frames := make([]*data.Frame, 0, len(monitors))
	for _, m := range monitors {
		reports, err := d.client.GetLighthouseReports(ctx, m.ID)
		if err != nil {
			return apiErrorResponse(err)
		}

		times := make([]time.Time, 0, len(reports))
		performance := make([]float64, 0, len(reports))
		accessibility := make([]float64, 0, len(reports))
		bestPractices := make([]float64, 0, len(reports))
		seo := make([]float64, 0, len(reports))
		fcp := make([]float64, 0, len(reports))
		lcp := make([]float64, 0, len(reports))
		tti := make([]float64, 0, len(reports))

		for _, r := range reports {
			t, err := parseOhDearTime(r.CreatedAt)
			if err != nil {
				continue
			}
			times = append(times, t)
			performance = append(performance, r.PerformanceScore)
			accessibility = append(accessibility, r.AccessibilityScore)
			bestPractices = append(bestPractices, r.BestPracticesScore)
			seo = append(seo, r.SEOScore)
			fcp = append(fcp, r.FirstContentfulPaintInMS)
			lcp = append(lcp, r.LargestContentfulPaintInMS)
			tti = append(tti, r.TimeToInteractiveInMS)
		}

		frames = append(frames, newMultiValueFrame(m.Label, "Lighthouse", len(monitors) > 1, times,
			"Performance", performance,
			"Accessibility", accessibility,
			"Best practices", bestPractices,
			"SEO", seo,
			"FCP (ms)", fcp,
			"LCP (ms)", lcp,
			"TTI (ms)", tti,
		))
	}

	return framesResponse(frames)
}

// newTimeSeriesFrame builds a time series frame with a single value field.
// The frame name doubles as the series name so legends show the monitor
// when multiple monitors are queried.
func newTimeSeriesFrame(label, field string, times []time.Time, values []float64) *data.Frame {
	return data.NewFrame(label,
		data.NewField("Time", nil, times),
		data.NewField(field, nil, values),
	)
}

// newMultiValueFrame builds a time series frame with multiple value fields.
// Pairs of (name, values) follow the shared time field. When prefixLabel is
// set, the monitor label is prefixed to each field name so series from
// different monitors stay distinguishable in legends.
func newMultiValueFrame(label, name string, prefixLabel bool, times []time.Time, pairs ...any) *data.Frame {
	fields := make([]*data.Field, 0, 1+len(pairs)/2)
	fields = append(fields, data.NewField("Time", nil, times))

	for i := 0; i < len(pairs); i += 2 {
		fieldName := pairs[i].(string)
		values := pairs[i+1].([]float64)
		if prefixLabel {
			fieldName = fmt.Sprintf("%s - %s", label, fieldName)
		}
		fields = append(fields, data.NewField(fieldName, nil, values))
	}

	frame := data.NewFrame(name, fields...)
	frame.Name = label
	return frame
}

func framesResponse(frames []*data.Frame) backend.DataResponse {
	if len(frames) == 0 {
		return backend.ErrDataResponse(backend.StatusNotFound, "no data returned by Oh Dear")
	}
	return backend.DataResponse{Frames: frames}
}

func defaultGroupBy(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// apiErrorResponse converts Oh Dear client errors into data responses with
// helpful messages.
func apiErrorResponse(err error) backend.DataResponse {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.Unauthorized():
			return backend.ErrDataResponseWithSource(backend.StatusUnauthorized, backend.ErrorSourceDownstream, "Your Oh Dear API token was rejected. Check the token in the data source configuration.")
		case apiErr.Forbidden():
			return backend.ErrDataResponseWithSource(backend.StatusForbidden, backend.ErrorSourceDownstream, "Your Oh Dear API token does not have permission for this monitor. It may be scoped to a subset of monitors.")
		case apiErr.NotFound():
			return backend.ErrDataResponseWithSource(backend.StatusNotFound, backend.ErrorSourceDownstream, "Monitor not found in your Oh Dear account.")
		}
		return backend.ErrDataResponseWithSource(backend.StatusBadGateway, backend.ErrorSourceDownstream, apiErr.Message)
	}
	return backend.ErrDataResponse(backend.StatusInternal, err.Error())
}

// CheckHealth handles health checks sent from Grafana to the plugin.
func (d *Datasource) CheckHealth(_ context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if d.settings.Secrets.Token == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "API token is missing",
		}, nil
	}

	user, err := d.client.GetMe(context.Background())
	if err != nil {
		var apiErr *Error
		if errors.As(err, &apiErr) && apiErr.Unauthorized() {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "Invalid API token",
			}, nil
		}
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Unable to reach Oh Dear: %v", err),
		}, nil
	}

	name := user.Name
	if name == "" {
		name = user.Email
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Connected as %s", name),
	}, nil
}

// CallResource exposes the monitor list to the frontend for the query editor.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	path := strings.TrimPrefix(req.Path, "/resources")

	switch {
	case path == "/monitors" || path == "":
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			message := err.Error()
			var apiErr *Error
			if errors.As(err, &apiErr) {
				if apiErr.Unauthorized() {
					message = "Invalid Oh Dear API token"
				}
				return sender.Send(jsonResponse(apiErr.Status, message))
			}
			return sender.Send(jsonResponse(502, message))
		}

		items := make([]map[string]any, 0, len(monitors))
		for _, m := range monitors {
			items = append(items, map[string]any{
				"id":     m.ID,
				"label":  m.Label,
				"url":    m.URL,
				"type":   m.Type,
				"status": m.SummarizedCheckResult,
			})
		}

		body, err := json.Marshal(items)
		if err != nil {
			return err
		}
		return sender.Send(&backend.CallResourceResponse{
			Status:  200,
			Headers: jsonContentType(),
			Body:    body,
		})
	default:
		return sender.Send(jsonResponse(404, "resource not found"))
	}
}

func jsonContentType() map[string][]string {
	return map[string][]string{"Content-Type": {"application/json"}}
}

func jsonResponse(status int, message string) *backend.CallResourceResponse {
	body, _ := json.Marshal(map[string]string{"error": message})
	return &backend.CallResourceResponse{
		Status:  status,
		Headers: jsonContentType(),
		Body:    body,
	}
}
