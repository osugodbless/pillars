package app

import (
	"bytes"
	"html/template"
	"net/http"
)

func RenderMemberDetail(w http.ResponseWriter, r *http.Request, view *MemberDashboardView, store *Store) {
	tmpl, err := template.ParseFiles("templates/member_detail.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		View  *MemberDashboardView
		Store *Store
	}{View: view, Store: store}
	var buf bytes.Buffer
	if r != nil && r.Header.Get("HX-Request") != "" {
		err = tmpl.ExecuteTemplate(&buf, "content", data)
	} else {
		err = tmpl.Execute(&buf, data)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}
