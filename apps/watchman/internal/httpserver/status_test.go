package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
)

type publicStatusDatabase struct{ stubDatabase }

func (publicStatusDatabase) PublicStatus(context.Context, time.Time) (store.PublicStatus, error) {
	return store.PublicStatus{Applications: []store.PublicApplicationStatus{{Slug: "farmami", DisplayName: "Farmami", Status: "operational", Components: []store.PublicComponent{}}}, Incidents: []store.PublicIncident{{ID: "public-id", Application: "farmami", Title: "Servicio restablecido", Message: "El servicio opera con normalidad.", Status: "resolved", PublishedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}}}, nil
}
func (publicStatusDatabase) SetIncidentPublication(context.Context, string, bool, string, string) error {
	return nil
}

func TestPublicStatusDoesNotRequireSession(t *testing.T) {
	server := newTestServer(publicStatusDatabase{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/public/status", nil)
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, private := range []string{"stack_trace", "last_error", "provider", "url"} {
		if containsText(body, private) {
			t.Fatalf("public response exposes %q: %s", private, body)
		}
	}
	if recorder.Header().Get("Cache-Control") == "" {
		t.Fatal("missing Cache-Control")
	}
}

func TestSystemHealthRequiresSession(t *testing.T) {
	server := newTestServer(stubDatabase{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil)
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != 401 {
		t.Fatalf("status=%d", recorder.Code)
	}
}

func containsText(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
