package app

import (
	"bytes"
	"html/template"
	"net/http"
)

func RenderMemberDetail(w http.ResponseWriter, view *MemberDashboardView, store *Store) {
	tmpl, err := template.ParseFiles("templates/member_detail.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		View  *MemberDashboardView
		Store *Store
	}{View: view, Store: store}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}
