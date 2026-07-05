package app

import (
	"bytes"
	"html/template"
	"net/http"
)

func RenderEventDetail(w http.ResponseWriter, view *EventDashboardView) {
	tmpl, err := template.ParseFiles("templates/event_detail.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, view); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}
