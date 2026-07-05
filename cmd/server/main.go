package main

import (
	"log"
	"net/http"

	"pillars/internal/app"
)

func main() {
	store, err := app.NewStoreWithPath("./data/pillars.db")
	if err != nil {
		log.Fatalf("create store: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		app.HandleIndex(w, r, store)
	})
	mux.HandleFunc("/members", func(w http.ResponseWriter, r *http.Request) {
		app.HandleMembers(w, r, store)
	})
	mux.HandleFunc("/attendance", func(w http.ResponseWriter, r *http.Request) {
		app.HandleAttendance(w, r, store)
	})
	mux.HandleFunc("/dues", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDues(w, r, store)
	})
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		app.HandleEvents(w, r, store)
	})
	mux.HandleFunc("/contributions", func(w http.ResponseWriter, r *http.Request) {
		app.HandleContribution(w, r, store)
	})
	mux.HandleFunc("/member-detail", func(w http.ResponseWriter, r *http.Request) {
		app.HandleMemberDetail(w, r, store)
	})
	mux.HandleFunc("/event-detail", func(w http.ResponseWriter, r *http.Request) {
		app.HandleEventDetail(w, r, store)
	})
	mux.HandleFunc("/health", app.HandleHealth)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
