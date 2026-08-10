package notification

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Links struct {
	AtalayaBaseURL     string
	SentryBaseURL      string
	SentryOrganization string
}

func Format(job domain.NotificationJob, links Links) string {
	if job.Kind == "interpreter_degraded" {
		return "⚠️ <b>Atalaya · INTERPRETER DEGRADADO</b>\n\n" +
			escape(job.Explanation) + "\n\nLos eventos siguen almacenándose. Revisá los logs del interpreter y OpenRouter."
	}
	identity := escape(displayName(job.Application))
	if job.Component != "" {
		identity += " · " + escape(displayComponent(job.Component))
	}
	heading := severityIcon(job.Severity) + " <b>" + identity + " · " + strings.ToUpper(escape(job.Severity)) + "</b>"
	if job.Kind == "group_summary" {
		heading = "📍 <b>" + identity + " · ACTUALIZACIÓN AGRUPADA</b>"
	}
	lines := []string{heading, "", "<b>" + escape(job.ErrorType) + "</b>", escape(job.Summary), "", escape(job.Explanation), "",
		fmt.Sprintf("Ocurrencias: <b>%d</b>", job.OccurrenceCount),
		"Primera: " + job.FirstOccurredAt.Format("2006-01-02 15:04:05 MST"),
		"Última: " + job.LastOccurredAt.Format("2006-01-02 15:04:05 MST")}
	if job.Release != "" {
		lines = append(lines, "Release: <code>"+escape(job.Release)+"</code>")
	}
	if len(job.SuggestedActions) > 0 {
		lines = append(lines, "", "<b>Acciones sugeridas</b>")
		for _, action := range job.SuggestedActions {
			lines = append(lines, "• "+escape(action))
		}
	}
	var linkLines []string
	if links.AtalayaBaseURL != "" && job.EventID != "" {
		linkLines = append(linkLines, `<a href="`+escape(strings.TrimRight(links.AtalayaBaseURL, "/")+"/internal/events/"+url.PathEscape(job.EventID))+`">Ver en Atalaya</a>`)
	}
	if job.Source == "sentry" && links.SentryOrganization != "" && job.SourceEventID != "" {
		u := strings.TrimRight(links.SentryBaseURL, "/") + "/organizations/" + url.PathEscape(links.SentryOrganization) + "/issues/?query=" + url.QueryEscape(job.SourceEventID)
		linkLines = append(linkLines, `<a href="`+escape(u)+`">Ver en Sentry</a>`)
	}
	if len(linkLines) > 0 {
		lines = append(lines, "", strings.Join(linkLines, " · "))
	}
	return strings.Join(lines, "\n")
}
func displayComponent(component string) string {
	if component == "backend" {
		return "Backend"
	}
	if component == "frontend" {
		return "Frontend"
	}
	return strings.Title(strings.ReplaceAll(component, "_", " "))
}

func escape(value string) string { return html.EscapeString(value) }
func severityIcon(severity string) string {
	if severity == "critical" {
		return "🔴"
	}
	if severity == "high" {
		return "🟠"
	}
	return "🟡"
}
func displayName(slug string) string {
	if slug == "prensap" {
		return "Prensap"
	}
	if slug == "wheels_house" {
		return "Wheels House"
	}
	return strings.Title(strings.ReplaceAll(slug, "_", " "))
}
