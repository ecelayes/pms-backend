package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestMetrics_RecordsCounter(t *testing.T) {
	e := echo.New()
	e.Use(Metrics())
	e.GET("/foo", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	before := readCounter(t, "GET", "/foo", "200")

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	after := readCounter(t, "GET", "/foo", "200")
	if after-before < 1 {
		t.Fatalf("counter not incremented: before=%v after=%v", before, after)
	}
}

func TestMetrics_RecordsInFlightGauge(t *testing.T) {
	e := echo.New()
	e.Use(Metrics())
	e.GET("/bar", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	before := readGaugeValue(t, HTTPRequestsInFlight)

	req := httptest.NewRequest(http.MethodGet, "/bar", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	after := readGaugeValue(t, HTTPRequestsInFlight)
	if after != before {
		t.Fatalf("gauge should return to baseline; before=%v after=%v", before, after)
	}
}

func TestMetrics_UnmatchedPath(t *testing.T) {
	e := echo.New()
	e.Use(Metrics())
	e.GET("/notfound", func(c echo.Context) error {
		return c.NoContent(http.StatusNotFound)
	})

	before := readCounter(t, "GET", "/notfound", "404")

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	after := readCounter(t, "GET", "/notfound", "404")
	if after-before < 1 {
		t.Fatalf("counter not incremented: before=%v after=%v", before, after)
	}
}

func TestMetrics_DurationObserved(t *testing.T) {
	e := echo.New()
	e.Use(Metrics())
	e.GET("/slow", func(c echo.Context) error {
		time.Sleep(5 * time.Millisecond)
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	count, sum := readHistogram(t, "GET", "/slow", "200")
	if count < 1 {
		t.Fatalf("expected sample count >= 1, got %d", count)
	}
	if sum < 0.001 {
		t.Fatalf("expected sum >= 1ms, got %v", sum)
	}
}

func readCounter(t *testing.T, labels ...string) float64 {
	t.Helper()
	counter, err := HTTPRequestsTotal.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("get counter: %v", err)
	}
	var m dto.Metric
	if err := counter.(prometheus.Metric).Write(&m); err != nil {
		t.Fatalf("write: %v", err)
	}
	return m.Counter.GetValue()
}

func readGaugeValue(t *testing.T, g prometheus.Gauge) float64 {
	t.Helper()
	var m dto.Metric
	if err := g.Write(&m); err != nil {
		t.Fatalf("write gauge: %v", err)
	}
	return m.Gauge.GetValue()
}

func readHistogram(t *testing.T, labels ...string) (uint64, float64) {
	t.Helper()
	hist, err := HTTPRequestDuration.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("get histogram: %v", err)
	}
	var m dto.Metric
	if err := hist.(prometheus.Metric).Write(&m); err != nil {
		t.Fatalf("write: %v", err)
	}
	return m.Histogram.GetSampleCount(), m.Histogram.GetSampleSum()
}

func TestMetrics_UnmatchedRoute(t *testing.T) {
	e := echo.New()
	e.Use(Metrics())

	req := httptest.NewRequest(http.MethodGet, "/no-such-route", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
