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
	mux.HandleFunc("/attendance-detail", func(w http.ResponseWriter, r *http.Request) {
		app.HandleAttendanceDetail(w, r, store)
	})
	mux.HandleFunc("/fines", func(w http.ResponseWriter, r *http.Request) {
		app.HandleAddFine(w, r, store)
	})
	mux.HandleFunc("/deduct-fine", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDeductFine(w, r, store)
	})
	mux.HandleFunc("/mark-fine-paid", func(w http.ResponseWriter, r *http.Request) {
		app.HandleMarkFinePaid(w, r, store)
	})
	mux.HandleFunc("/mark-dues-paid", func(w http.ResponseWriter, r *http.Request) {
		app.HandleMarkDuesPaid(w, r, store)
	})
	mux.HandleFunc("/mark-contribution-paid", func(w http.ResponseWriter, r *http.Request) {
		app.HandleMarkContributionPaid(w, r, store)
	})
	mux.HandleFunc("/export-attendance", func(w http.ResponseWriter, r *http.Request) {
		app.HandleExportAttendancePDF(w, r, store)
	})
	mux.HandleFunc("/export-contributions", func(w http.ResponseWriter, r *http.Request) {
		app.HandleExportContributionsPDF(w, r, store)
	})
	mux.HandleFunc("/settle-event", func(w http.ResponseWriter, r *http.Request) {
		app.HandleSettleEvent(w, r, store)
	})
	mux.HandleFunc("/promote-to-active", func(w http.ResponseWriter, r *http.Request) {
		app.HandlePromoteToActive(w, r, store)
	})
	mux.HandleFunc("/extend-probation", func(w http.ResponseWriter, r *http.Request) {
		app.HandleExtendProbation(w, r, store)
	})
	mux.HandleFunc("/delete-member", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDeleteMember(w, r, store)
	})
	mux.HandleFunc("/reports/committee", func(w http.ResponseWriter, r *http.Request) {
		app.HandleCommitteeReport(w, r, store)
	})
	mux.HandleFunc("/reports/arrears", func(w http.ResponseWriter, r *http.Request) {
		app.HandleArrearsReport(w, r, store)
	})
	mux.HandleFunc("/health", app.HandleHealth)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
