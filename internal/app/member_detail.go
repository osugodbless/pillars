package app

import (
	"bytes"
	"net/http"
)

func RenderMemberDetail(w http.ResponseWriter, r *http.Request, view *MemberDashboardView, store *Store) {
	tmpl, err := parseTemplates("templates/base.html", "templates/sidebar.html", "templates/member_detail.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Title      string
		ActivePage string
		View       *MemberDashboardView
		Store      *Store
	}{Title: view.Member.Name, ActivePage: "members", View: view, Store: store}
	var buf bytes.Buffer
	if r != nil && r.Header.Get("HX-Request") != "" {
		err = tmpl.ExecuteTemplate(&buf, "content", data)
	} else {
		err = tmpl.ExecuteTemplate(&buf, "base.html", data)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}
