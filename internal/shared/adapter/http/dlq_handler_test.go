package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecelayes/pms-backend/internal/shared/adapter/redis"
	"github.com/labstack/echo/v4"
)

type mockDLQStore struct {
	stats       *redis.DLQStats
	statsErr    error
	messages    *redis.DLQStats
	messagesErr error
	deleted     int64
	deleteErr   error
	purgeCount  int64
	purgeErr    error
}

func (m *mockDLQStore) GetStats(ctx context.Context) (*redis.DLQStats, error) {
	return m.stats, m.statsErr
}

func (m *mockDLQStore) GetMessages(ctx context.Context, limit int64) (*redis.DLQStats, error) {
	return m.messages, m.messagesErr
}

func (m *mockDLQStore) DeleteMessage(ctx context.Context, id string) (int64, error) {
	return m.deleted, m.deleteErr
}

func (m *mockDLQStore) Purge(ctx context.Context) (int64, error) {
	return m.purgeCount, m.purgeErr
}

func TestDLQHandler_GetStats_Success(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{stats: &redis.DLQStats{TotalMessages: 5}}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodGet, "/admin/dlq/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.GetStats(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestDLQHandler_GetStats_Error(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{statsErr: context.DeadlineExceeded}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodGet, "/admin/dlq/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.GetStats(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestDLQHandler_GetMessages_Success(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{messages: &redis.DLQStats{TotalMessages: 0}}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodGet, "/admin/dlq/messages", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.GetMessages(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestDLQHandler_GetMessages_Error(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{messagesErr: context.Canceled}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodGet, "/admin/dlq/messages", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.GetMessages(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestDLQHandler_DeleteMessage_Success(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{deleted: 1}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodDelete, "/admin/dlq/messages/abc123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues("abc123")
	_ = h.DeleteMessage(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestDLQHandler_DeleteMessage_NotFound(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{deleted: 0}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodDelete, "/admin/dlq/messages/abc123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues("abc123")
	_ = h.DeleteMessage(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestDLQHandler_DeleteMessage_Error(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{deleteErr: context.DeadlineExceeded}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodDelete, "/admin/dlq/messages/abc123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues("abc123")
	_ = h.DeleteMessage(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestDLQHandler_PurgeDLQ_Success(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{purgeCount: 10}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodDelete, "/admin/dlq/purge", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.PurgeDLQ(c)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestDLQHandler_PurgeDLQ_Error(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{purgeErr: context.DeadlineExceeded}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodDelete, "/admin/dlq/purge", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.PurgeDLQ(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestNewDLQHandler(t *testing.T) {
	h := NewDLQHandler(nil)
	if h == nil {
		t.Error("Expected non-nil handler")
	}
}

func TestDLQHandler_DeleteMessage_EmptyIDWithError(t *testing.T) {
	e := echo.New()
	store := &mockDLQStore{deleteErr: context.DeadlineExceeded}
	h := &DLQHandler{store: store}

	req := httptest.NewRequest(http.MethodDelete, "/admin/dlq/messages/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")
	_ = h.DeleteMessage(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty ID with err, got %d", rec.Code)
	}
}
