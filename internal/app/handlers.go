package app

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type PageData struct {
	Members          []MemberView
	Attendance       []AttendanceView
	AttendanceGroups []AttendanceGroup
	Events           []Event
	Fines            []Fine
	Dues             []DuesRecord
	Contributions    []Contribution
	Stats            StatsView
}

type StatsView struct {
	TotalMembers       int
	ProbationMembers   int
	OutstandingBalance float64
	OpenEvents         int
	OutstandingFines   int
}

type MemberView struct {
	ID      int
	Name    string
	Email   string
	Phone   string
	Status  string
	Balance float64
}

type AttendanceView struct {
	MemberName  string
	MeetingDate string
	Status      string
	Note        string
}

type MemberDashboardView struct {
	Member        Member
	Summary       MemberDashboardSummary
	Attendance    []AttendanceRecord
	Dues          []DuesRecord
	Fines         []Fine
	Contributions []Contribution
	Events        []Event
}

type EventDashboardView struct {
	Event          Event
	Contributions  []Contribution
	Members        []MemberView
	TotalCollected float64
}

func RenderIndex(w http.ResponseWriter, r *http.Request, store *Store) error {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return err
	}

	memberViews := make([]MemberView, 0, len(store.Members))
	for _, member := range store.Members {
		memberViews = append(memberViews, MemberView{ID: member.ID, Name: member.Name, Email: member.Email, Phone: member.Phone, Status: member.Status, Balance: store.MemberBalance(member.ID)})
	}

	attendanceViews := make([]AttendanceView, 0, len(store.Attendance))
	for _, record := range store.Attendance {
		memberName := "Unknown"
		for _, member := range store.Members {
			if member.ID == record.MemberID {
				memberName = member.Name
				break
			}
		}
		attendanceViews = append(attendanceViews, AttendanceView{MemberName: memberName, MeetingDate: record.MeetingDate, Status: record.Status, Note: record.Note})
	}

	attendanceGroups := groupAttendanceByDate(store.Attendance)

	stats := StatsView{TotalMembers: len(memberViews)}
	for _, member := range memberViews {
		if member.Status == "probation" {
			stats.ProbationMembers++
		}
		stats.OutstandingBalance += store.MemberBalance(member.ID)
	}
	for _, event := range store.Events {
		if event.Status == "open" {
			stats.OpenEvents++
		}
	}
	for _, fine := range store.Fines {
		if fine.Status == "outstanding" {
			stats.OutstandingFines++
		}
	}

	data := PageData{Members: memberViews, Attendance: attendanceViews, AttendanceGroups: attendanceGroups, Events: store.Events, Fines: store.Fines, Dues: store.Dues, Contributions: store.Contributions, Stats: stats}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = w.Write(buf.Bytes())
	return err
}

func HandleIndex(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method == http.MethodGet {
		if err := RenderIndex(w, r, store); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func buildMemberDashboardView(store *Store, memberID int) *MemberDashboardView {
	member := Member{}
	for _, candidate := range store.Members {
		if candidate.ID == memberID {
			member = candidate
			break
		}
	}
	if member.ID == 0 {
		return nil
	}

	var attendance []AttendanceRecord
	var dues []DuesRecord
	var fines []Fine
	var contributions []Contribution
	for _, record := range store.Attendance {
		if record.MemberID == memberID {
			attendance = append(attendance, record)
		}
	}
	for _, due := range store.Dues {
		if due.MemberID == memberID {
			dues = append(dues, due)
		}
	}
	for _, fine := range store.Fines {
		if fine.MemberID == memberID {
			fines = append(fines, fine)
		}
	}
	for _, contribution := range store.Contributions {
		if contribution.MemberID == memberID {
			contributions = append(contributions, contribution)
		}
	}

	summary := store.MemberDashboardSummary(memberID, 30)
	summary.AbsenceFineAmount = store.Settings.AbsenceFineAmount
	return &MemberDashboardView{Member: member, Summary: summary, Attendance: attendance, Dues: dues, Fines: fines, Contributions: contributions, Events: store.Events}
}

func buildEventDashboardView(store *Store, eventID int) *EventDashboardView {
	event := Event{}
	for _, candidate := range store.Events {
		if candidate.ID == eventID {
			event = candidate
			break
		}
	}
	if event.ID == 0 {
		return nil
	}

	var contributions []Contribution
	var members []MemberView
	var totalCollected float64
	for _, contribution := range store.Contributions {
		if contribution.EventID == eventID {
			contributions = append(contributions, contribution)
			totalCollected += contribution.Amount
		}
	}
	for _, member := range store.Members {
		memberView := MemberView{ID: member.ID, Name: member.Name, Email: member.Email, Status: member.Status, Balance: 0}
		for _, contribution := range contributions {
			if contribution.MemberID == member.ID {
				memberView.Balance = contribution.Amount
				memberView.Status = contribution.Status
				break
			}
		}
		if memberView.Status == "" {
			memberView.Status = "not_paid"
		}
		members = append(members, memberView)
	}
	return &EventDashboardView{Event: event, Contributions: contributions, Members: members, TotalCollected: totalCollected}
}

func HandleMembers(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	member := Member{Name: name, Email: r.FormValue("email"), Phone: r.FormValue("phone"), Status: r.FormValue("status"), JoinedAt: time.Now().Format(time.RFC3339)}
	if err := store.CreateMember(member); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleAttendance(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	meetingDate := r.FormValue("meeting_date")
	if meetingDate == "" {
		http.Error(w, "meeting date is required", http.StatusBadRequest)
		return
	}

	if memberIDStr := r.FormValue("member_id"); memberIDStr != "" {
		memberID, err := strconv.Atoi(memberIDStr)
		if err != nil || memberID <= 0 {
			http.Error(w, "valid member id is required", http.StatusBadRequest)
			return
		}
		status := r.FormValue("status")
		if status == "" {
			http.Error(w, "status is required", http.StatusBadRequest)
			return
		}
		if err := store.RecordAttendance(memberID, meetingDate, status, r.FormValue("note")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	note := r.FormValue("note")
	for _, member := range store.Members {
		presentKey := "present_" + strconv.Itoa(member.ID)
		duesKey := "dues_" + strconv.Itoa(member.ID)
		absenteeismKey := "absenteeism_" + strconv.Itoa(member.ID)
		present := len(r.Form[presentKey]) > 0
		duesPaid := len(r.Form[duesKey]) > 0
		absenteeism := len(r.Form[absenteeismKey]) > 0
		status := attendanceStatusFromSelection(present, absenteeism)
		if err := recordAttendanceAndDues(store, member.ID, meetingDate, status, note, duesPaid, 1000); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleDues(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	memberID, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil || memberID <= 0 {
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "valid amount is required", http.StatusBadRequest)
		return
	}
	if err := store.AddDues(DuesRecord{MemberID: memberID, Amount: amount, Status: "pending", DueDate: r.FormValue("due_date")}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleEvents(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	minAmountExpected, err := strconv.ParseFloat(r.FormValue("min_amount_expected"), 64)
	if err != nil {
		http.Error(w, "valid minimum amount expected is required", http.StatusBadRequest)
		return
	}
	if err := store.AddEvent(Event{Title: r.FormValue("title"), Description: r.FormValue("description"), Date: r.FormValue("date"), MinAmountExpected: minAmountExpected, Status: "open"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleContribution(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	eventID, err := strconv.Atoi(r.FormValue("event_id"))
	if err != nil || eventID <= 0 {
		http.Error(w, "valid event id is required", http.StatusBadRequest)
		return
	}
	memberID, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil || memberID <= 0 {
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "valid amount is required", http.StatusBadRequest)
		return
	}
	status := r.FormValue("status")
	if status == "" {
		status = "paid"
	}
	if err := store.AddContribution(Contribution{EventID: eventID, MemberID: memberID, Amount: amount, Status: status}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/?member_id="+strconv.Itoa(memberID), http.StatusSeeOther)
}

func HandleMemberDetail(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	memberID, err := strconv.Atoi(r.URL.Query().Get("member_id"))
	if err != nil || memberID <= 0 {
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	view := buildMemberDashboardView(store, memberID)
	if view == nil {
		http.NotFound(w, r)
		return
	}
	RenderMemberDetail(w, view, store)
}

func HandleEventDetail(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	eventID, err := strconv.Atoi(r.URL.Query().Get("event_id"))
	if err != nil || eventID <= 0 {
		http.Error(w, "valid event id is required", http.StatusBadRequest)
		return
	}
	view := buildEventDashboardView(store, eventID)
	if view == nil {
		http.NotFound(w, r)
		return
	}
	RenderEventDetail(w, view)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}
