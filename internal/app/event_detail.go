package app

import (
	"bytes"
	"html/template"
	"net/http"
)

func RenderEventDetail(w http.ResponseWriter, r *http.Request, view *EventDashboardView) {
	tmpl, err := template.ParseFiles("templates/event_detail.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if r != nil && r.Header.Get("HX-Request") != "" {
		err = tmpl.ExecuteTemplate(&buf, "content", view)
	} else {
		err = tmpl.Execute(&buf, view)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}
