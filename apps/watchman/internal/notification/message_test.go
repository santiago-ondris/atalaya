package notification

import (
	"strings"
	"testing"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

func TestFormatEventEscapesContentAndIncludesLinks(t *testing.T) {
	job := domain.NotificationJob{EventID: "event-id", Kind: "event_alert", Application: "prensap", Component: "backend", Source: "sentry", SourceEventID: "source-id",
		ErrorType: "TypeError <script>", Summary: "No carga & falla", Explanation: "Explicación", Severity: "high", OccurrenceCount: 3,
		FirstOccurredAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC), LastOccurredAt: time.Date(2026, 8, 7, 10, 5, 0, 0, time.UTC), SuggestedActions: []string{"Revisar <logs>"}}
	message := Format(job, Links{AtalayaBaseURL: "https://atalaya.example", SentryBaseURL: "https://sentry.io", SentryOrganization: "acme"})
	for _, expected := range []string{"Prensap · Backend · HIGH", "TypeError &lt;script&gt;", "No carga &amp; falla", "Ocurrencias: <b>3</b>", "Ver en Atalaya", "Ver en Sentry"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("message missing %q:\n%s", expected, message)
		}
	}
	if strings.Contains(message, "<script>") {
		t.Fatal("unescaped HTML in message")
	}
}

func TestFormatInterpreterDegradedDoesNotCreateRecursiveLink(t *testing.T) {
	message := Format(domain.NotificationJob{Kind: "interpreter_degraded", Explanation: "Falló OpenRouter"}, Links{AtalayaBaseURL: "https://atalaya.example"})
	if !strings.Contains(message, "INTERPRETER DEGRADADO") || strings.Contains(message, "Ver en") {
		t.Fatalf("unexpected message: %s", message)
	}
}
