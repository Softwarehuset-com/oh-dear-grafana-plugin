package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Softwarehuset-com/oh-dear-grafana-plugin/pkg/models"
)

// Monitor represents an Oh Dear monitor.
type Monitor struct {
	ID                    int64  `json:"id"`
	Type                  string `json:"type"`
	URL                   string `json:"url"`
	Label                 string `json:"label"`
	GroupName             string `json:"group_name"`
	LatestRunDate         string `json:"latest_run_date"`
	SummarizedCheckResult string `json:"summarized_check_result"`
}

type UptimePoint struct {
	Datetime         string  `json:"datetime"`
	UptimePercentage float64 `json:"uptime_percentage"`
}

type DowntimeRecord struct {
	ID            int64  `json:"id"`
	StartedAt     string `json:"started_at"`
	EndedAt       string `json:"ended_at"`
	NotesHTML     string `json:"notes_html"`
	NotesMarkdown string `json:"notes_markdown"`
}

type HTTPMetricPoint struct {
	DNSTimeInSeconds                    float64 `json:"dns_time_in_seconds"`
	TCPTimeInSeconds                    float64 `json:"tcp_time_in_seconds"`
	SSLHandshakeTimeInSeconds           float64 `json:"ssl_handshake_time_in_seconds"`
	RemoteServerProcessingTimeInSeconds float64 `json:"remote_server_processing_time_in_seconds"`
	DownloadTimeInSeconds               float64 `json:"download_time_in_seconds"`
	TotalTimeInSeconds                  float64 `json:"total_time_in_seconds"`
	Date                                string  `json:"date"`
}

type PingMetricPoint struct {
	MinimumTimeInMS    float64 `json:"minimum_time_in_ms"`
	MaximumTimeInMS    float64 `json:"maximum_time_in_ms"`
	AverageTimeInMS    float64 `json:"average_time_in_ms"`
	PacketLossPercent  float64 `json:"packet_loss_percentage"`
	UptimePercentage   float64 `json:"uptime_percentage"`
	DowntimePercentage float64 `json:"downtime_percentage"`
	Date               string  `json:"date"`
}

type TCPMetricPoint struct {
	TimeToConnectInMS  float64 `json:"time_to_connect_in_ms"`
	UptimePercentage   float64 `json:"uptime_percentage"`
	DowntimePercentage float64 `json:"downtime_percentage"`
	Date               string  `json:"date"`
}

type LighthouseReport struct {
	PerformanceScore           float64 `json:"performance_score"`
	AccessibilityScore         float64 `json:"accessibility_score"`
	BestPracticesScore         float64 `json:"best_practices_score"`
	SEOScore                   float64 `json:"seo_score"`
	FirstContentfulPaintInMS   float64 `json:"first_contentful_paint_in_ms"`
	SpeedIndexInMS             float64 `json:"speed_index_in_ms"`
	LargestContentfulPaintInMS float64 `json:"largest_contentful_paint_in_ms"`
	TimeToInteractiveInMS      float64 `json:"time_to_interactive_in_ms"`
	TotalBlockingTimeInMS      float64 `json:"total_blocking_time_in_ms"`
	CumulativeLayoutShift      float64 `json:"cumulative_layout_shift"`
	CreatedAt                  string  `json:"created_at"`
}

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	PhotoURL string `json:"photo_url"`
}

type paginated struct {
	Data  json.RawMessage `json:"data"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

// Client is an HTTP client for the Oh Dear API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(settings *models.PluginSettings) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(settings.BaseURL, "/"),
		token:   settings.Secrets.Token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Error is an Oh Dear API error with the HTTP status code.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unauthorized() bool {
	return e.Status == http.StatusUnauthorized
}

func (e *Error) Forbidden() bool {
	return e.Status == http.StatusForbidden
}

func (e *Error) NotFound() bool {
	return e.Status == http.StatusNotFound
}

// Get performs an authenticated GET request and decodes the JSON response into out.
func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request to Oh Dear failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("reading Oh Dear response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &Error{Status: resp.StatusCode, Message: apiErrorMessage(resp.StatusCode, body)}
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("decoding Oh Dear response: %w", err)
		}
	}

	return nil
}

func apiErrorMessage(status int, body []byte) string {
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Message != "" {
		return fmt.Sprintf("Oh Dear API returned %d: %s", status, payload.Message)
	}
	return fmt.Sprintf("Oh Dear API returned %d: %s", status, strings.TrimSpace(string(body)))
}

// GetMonitors fetches all monitors for the account.
func (c *Client) GetMonitors(ctx context.Context) ([]Monitor, error) {
	var monitors []Monitor
	pageNumber := 1
	for {
		query := url.Values{}
		query.Set("page[number]", fmt.Sprintf("%d", pageNumber))
		query.Set("page[size]", "200")

		var result paginated
		if err := c.Get(ctx, "/monitors", query, &result); err != nil {
			return nil, err
		}

		var page []Monitor
		if err := json.Unmarshal(result.Data, &page); err != nil {
			return nil, err
		}
		monitors = append(monitors, page...)

		if result.Links.Next == "" || pageNumber >= 10 {
			break
		}
		pageNumber++
	}
	return monitors, nil
}

// GetMonitor fetches a single monitor.
func (c *Client) GetMonitor(ctx context.Context, monitorID int64) (*Monitor, error) {
	var monitor Monitor
	if err := c.Get(ctx, fmt.Sprintf("/monitors/%d", monitorID), nil, &monitor); err != nil {
		return nil, err
	}
	return &monitor, nil
}

// GetUptime fetches uptime percentages for a monitor in [from, to].
func (c *Client) GetUptime(ctx context.Context, monitorID int64, from, to time.Time, split string) ([]UptimePoint, error) {
	query := url.Values{}
	query.Set("filter[started_at]", from.UTC().Format("20060102150405"))
	query.Set("filter[ended_at]", to.UTC().Format("20060102150405"))
	if split != "" {
		query.Set("split", split)
	}

	var points []UptimePoint
	if err := c.Get(ctx, fmt.Sprintf("/monitors/%d/uptime", monitorID), query, &points); err != nil {
		return nil, err
	}
	return points, nil
}

// GetDowntime fetches downtime periods for a monitor in [from, to].
func (c *Client) GetDowntime(ctx context.Context, monitorID int64, from, to time.Time) ([]DowntimeRecord, error) {
	query := url.Values{}
	query.Set("filter[started_at]", from.UTC().Format("20060102150405"))
	query.Set("filter[ended_at]", to.UTC().Format("20060102150405"))

	var result paginated
	if err := c.Get(ctx, fmt.Sprintf("/monitors/%d/downtime", monitorID), query, &result); err != nil {
		return nil, err
	}

	var records []DowntimeRecord
	if err := json.Unmarshal(result.Data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// GetHTTPUptimeMetrics fetches HTTP response timing metrics for a monitor.
func (c *Client) GetHTTPUptimeMetrics(ctx context.Context, monitorID int64, from, to time.Time, groupBy string) ([]HTTPMetricPoint, error) {
	var points []HTTPMetricPoint
	if err := c.getMetrics(ctx, monitorID, "http-uptime-metrics", from, to, groupBy, &points); err != nil {
		return nil, err
	}
	return points, nil
}

// GetPingUptimeMetrics fetches ping metrics for a monitor.
func (c *Client) GetPingUptimeMetrics(ctx context.Context, monitorID int64, from, to time.Time, groupBy string) ([]PingMetricPoint, error) {
	var points []PingMetricPoint
	if err := c.getMetrics(ctx, monitorID, "ping-uptime-metrics", from, to, groupBy, &points); err != nil {
		return nil, err
	}
	return points, nil
}

// GetTCPUptimeMetrics fetches TCP metrics for a monitor.
func (c *Client) GetTCPUptimeMetrics(ctx context.Context, monitorID int64, from, to time.Time, groupBy string) ([]TCPMetricPoint, error) {
	var points []TCPMetricPoint
	if err := c.getMetrics(ctx, monitorID, "tcp-uptime-metrics", from, to, groupBy, &points); err != nil {
		return nil, err
	}
	return points, nil
}

func (c *Client) getMetrics(ctx context.Context, monitorID int64, endpoint string, from, to time.Time, groupBy string, out any) error {
	query := url.Values{}
	query.Set("filter[start]", from.UTC().Format("20060102150405"))
	query.Set("filter[end]", to.UTC().Format("20060102150405"))
	if groupBy != "" {
		query.Set("filter[group_by]", groupBy)
	}

	var result struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.Get(ctx, fmt.Sprintf("/monitors/%d/%s", monitorID, endpoint), query, &result); err != nil {
		return err
	}
	if err := json.Unmarshal(result.Data, out); err != nil {
		return err
	}
	return nil
}

// GetLighthouseReports fetches Lighthouse reports created in [from, to] for a monitor.
func (c *Client) GetLighthouseReports(ctx context.Context, monitorID int64, from, to time.Time) ([]LighthouseReport, error) {
	var reports []LighthouseReport
	pageNumber := 1
	for {
		query := url.Values{}
		query.Set("filter[created_at]", from.UTC().Format("20060102150405"))
		query.Set("page[number]", fmt.Sprintf("%d", pageNumber))
		query.Set("page[size]", "200")

		var result paginated
		if err := c.Get(ctx, fmt.Sprintf("/monitors/%d/lighthouse-reports", monitorID), query, &result); err != nil {
			return nil, err
		}

		var page []LighthouseReport
		if err := json.Unmarshal(result.Data, &page); err != nil {
			return nil, err
		}
		for _, r := range page {
			if created, err := parseOhDearTime(r.CreatedAt); err == nil && !created.After(to) {
				reports = append(reports, r)
			}
		}

		if result.Links.Next == "" || pageNumber >= 5 {
			break
		}
		pageNumber++
	}
	return reports, nil
}

// GetMe fetches the authenticated user.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	var user User
	if err := c.Get(ctx, "/me", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// parseOhDearTime parses the timestamp formats used by the Oh Dear API.
func parseOhDearTime(value string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized timestamp %q", value)
}
