package app

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
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
	TotalMembers                int
	ProbationMembers            int
	OpenEvents                  int
	OutstandingFines            int
	TotalDuesPaid               float64
	TotalDuesOwed               float64
	TotalFines                  float64
	FinesOwed                   float64
	FinesPaid                   float64
	TotalTreasuryBalance        float64
	TotalOutstandingReceivables float64
	AtRiskMembersCount          int
	EventFundingProgress        []EventFundingView
}

type EventFundingView struct {
	EventID        int
	Title          string
	TotalCollected float64
	GoalAmount     float64
	Percentage     float64
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
	Filter         string
}

type AttendanceRecordView struct {
	MemberName string
	Status     string
	Note       string
}

type AttendanceDetailView struct {
	MeetingDate string
	Records     []AttendanceRecordView
	Filter      string
}

func buildPageData(store *Store) PageData {
	memberViews := make([]MemberView, 0, len(store.Members))
	for _, member := range store.Members {
		if member.Status == "ex-member" {
			continue
		}
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
	for _, due := range store.Dues {
		if due.Status == "paid" || due.Status == "partially_paid" {
			stats.TotalDuesPaid += due.Amount
		} else {
			stats.TotalDuesOwed += due.Amount
		}
	}
	for _, fine := range store.Fines {
		if fine.Status == "outstanding" {
			stats.FinesOwed += fine.Amount
		} else {
			stats.FinesPaid += fine.Amount
		}
	}
	stats.TotalFines = stats.FinesOwed + stats.FinesPaid

	stats.TotalTreasuryBalance = store.TotalTreasuryBalance()
	stats.TotalOutstandingReceivables = store.TotalOutstandingReceivables()
	stats.AtRiskMembersCount = store.AtRiskMembersCount()

	for _, event := range store.Events {
		if event.Status == "open" {
			collected := 0.0
			expectedCount := 0
			for _, contrib := range store.Contributions {
				if contrib.EventID == event.ID {
					expectedCount++
					if contrib.Status == "paid" || contrib.Status == "partially_paid" {
						collected += contrib.Amount
					}
				}
			}
			goal := event.MinAmountExpected * float64(expectedCount)
			percentage := 0.0
			if goal > 0 {
				percentage = (collected / goal) * 100
				if percentage > 100 {
					percentage = 100
				}
			}
			stats.EventFundingProgress = append(stats.EventFundingProgress, EventFundingView{
				EventID:        event.ID,
				Title:          event.Title,
				TotalCollected: collected,
				GoalAmount:     goal,
				Percentage:     percentage,
			})
		}
	}

	return PageData{Members: memberViews, Attendance: attendanceViews, AttendanceGroups: attendanceGroups, Events: store.Events, Fines: store.Fines, Dues: store.Dues, Contributions: store.Contributions, Stats: stats}
}

func RenderIndex(w http.ResponseWriter, r *http.Request, store *Store) error {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return err
	}

	data := buildPageData(store)

	var buf bytes.Buffer
	if r != nil && r.Header.Get("HX-Request") != "" {
		err = tmpl.ExecuteTemplate(&buf, "content", data)
	} else {
		err = tmpl.Execute(&buf, data)
	}
	if err != nil {
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

func buildEventDashboardView(store *Store, eventID int, filter string) *EventDashboardView {
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
			if contribution.Status == "paid" || contribution.Status == "partially_paid" {
				totalCollected += contribution.Amount
			}
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
		if memberView.Status == "" || memberView.Status == "pending" {
			memberView.Status = "not_paid"
		}
		if filter != "" && filter != "all" {
			if filter == "paid" {
				if memberView.Status != "paid" && memberView.Status != "partially_paid" {
					continue
				}
			} else if filter == "not_paid" {
				if memberView.Status != "not_paid" {
					continue
				}
			} else if memberView.Status != filter {
				continue
			}
		}
		members = append(members, memberView)
	}
	return &EventDashboardView{Event: event, Contributions: contributions, Members: members, TotalCollected: totalCollected, Filter: filter}
}

func renderIndexFragment(w http.ResponseWriter, r *http.Request, store *Store, msg string, msgType string) {
	w.Header().Set("HX-Trigger", `{"showToast":{"message":"`+msg+`","type":"`+msgType+`"}}`)
	if err := RenderIndex(w, r, store); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleMembers(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Method not allowed", "error")
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Name is required", "error")
			return
		}
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	member := Member{Name: name, Email: r.FormValue("email"), Phone: r.FormValue("phone"), Status: r.FormValue("status"), JoinedAt: time.Now().Format(time.RFC3339)}
	if err := store.CreateMember(member); err != nil {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, err.Error(), "error")
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		renderIndexFragment(w, r, store, "Member added", "success")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleAttendance(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Method not allowed", "error")
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	meetingDate := r.FormValue("meeting_date")
	if meetingDate == "" {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Meeting date is required", "error")
			return
		}
		http.Error(w, "meeting date is required", http.StatusBadRequest)
		return
	}

	if memberIDStr := r.FormValue("member_id"); memberIDStr != "" {
		memberID, err := strconv.Atoi(memberIDStr)
		if err != nil || memberID <= 0 {
			if r.Header.Get("HX-Request") != "" {
				renderIndexFragment(w, r, store, "Valid member id is required", "error")
				return
			}
			http.Error(w, "valid member id is required", http.StatusBadRequest)
			return
		}
		status := r.FormValue("status")
		if status == "" {
			if r.Header.Get("HX-Request") != "" {
				renderIndexFragment(w, r, store, "Status is required", "error")
				return
			}
			http.Error(w, "status is required", http.StatusBadRequest)
			return
		}
		if err := store.RecordAttendance(memberID, meetingDate, status, r.FormValue("note")); err != nil {
			if r.Header.Get("HX-Request") != "" {
				renderIndexFragment(w, r, store, err.Error(), "error")
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Attendance recorded", "success")
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	note := r.FormValue("note")
	for _, member := range store.Members {
		if member.Status == "ex-member" {
			continue
		}
		presentKey := "present_" + strconv.Itoa(member.ID)
		duesKey := "dues_" + strconv.Itoa(member.ID)
		absenteeismKey := "absenteeism_" + strconv.Itoa(member.ID)
		lateKey := "late_" + strconv.Itoa(member.ID)
		present := len(r.Form[presentKey]) > 0
		duesPaid := len(r.Form[duesKey]) > 0
		absenteeism := len(r.Form[absenteeismKey]) > 0
		isLate := len(r.Form[lateKey]) > 0
		status := attendanceStatusFromSelection(present, absenteeism)
		if err := recordAttendanceAndDues(store, member.ID, meetingDate, status, note, duesPaid, store.Settings.DuesAmount); err != nil {
			if r.Header.Get("HX-Request") != "" {
				renderIndexFragment(w, r, store, err.Error(), "error")
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if isLate {
			if err := store.AddFine(Fine{MemberID: member.ID, Amount: store.Settings.LateFineAmount, Status: "outstanding", Reason: "Lateness", FineDate: meetingDate}); err != nil {
				if r.Header.Get("HX-Request") != "" {
					renderIndexFragment(w, r, store, err.Error(), "error")
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
	if r.Header.Get("HX-Request") != "" {
		renderIndexFragment(w, r, store, "Attendance saved", "success")
		return
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

func HandleAddFine(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Method not allowed", "error")
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	memberID, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil || memberID <= 0 {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Valid member id is required", "error")
			return
		}
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Valid amount is required", "error")
			return
		}
		http.Error(w, "valid amount is required", http.StatusBadRequest)
		return
	}
	reason := r.FormValue("reason")
	if reason == "" {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Reason is required", "error")
			return
		}
		http.Error(w, "reason is required", http.StatusBadRequest)
		return
	}
	if err := store.AddFine(Fine{MemberID: memberID, Amount: amount, Status: "outstanding", Reason: reason, FineDate: r.FormValue("fine_date")}); err != nil {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, err.Error(), "error")
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		renderIndexFragment(w, r, store, "Fine issued", "success")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleDeductFine(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	memberID, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil || memberID <= 0 {
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	if fineIDStr := r.FormValue("fine_id"); fineIDStr != "" {
		fineID, err := strconv.Atoi(fineIDStr)
		if err != nil || fineID <= 0 {
			http.Error(w, "valid fine id is required", http.StatusBadRequest)
			return
		}
		if err := store.DeductFineFromDues(memberID, fineID); err != nil {
			if r.Header.Get("HX-Request") != "" {
				w.Header().Set("HX-Trigger", `{"showToast":{"message":"`+err.Error()+`","type":"error"}}`)
				view := buildMemberDashboardView(store, memberID)
				if view != nil {
					RenderMemberDetail(w, r, view, store)
					return
				}
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if err := store.DeductAllFinesFromDues(memberID); err != nil {
			if r.Header.Get("HX-Request") != "" {
				w.Header().Set("HX-Trigger", `{"showToast":{"message":"`+err.Error()+`","type":"error"}}`)
				view := buildMemberDashboardView(store, memberID)
				if view != nil {
					RenderMemberDetail(w, r, view, store)
					return
				}
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Trigger", `{"showToast":{"message":"Deducted successfully","type":"success"}}`)
		view := buildMemberDashboardView(store, memberID)
		if view != nil {
			RenderMemberDetail(w, r, view, store)
			return
		}
	}
	http.Redirect(w, r, "/member-detail?member_id="+strconv.Itoa(memberID), http.StatusSeeOther)
}

func HandleMarkDuesPaid(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	memberID, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil || memberID <= 0 {
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	duesID, err := strconv.Atoi(r.FormValue("dues_id"))
	if err != nil || duesID <= 0 {
		http.Error(w, "valid dues id is required", http.StatusBadRequest)
		return
	}
	if err := store.MarkDuesPaid(memberID, duesID); err != nil {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Trigger", `{"showToast":{"message":"`+err.Error()+`","type":"error"}}`)
			view := buildMemberDashboardView(store, memberID)
			if view != nil {
				RenderMemberDetail(w, r, view, store)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Trigger", `{"showToast":{"message":"Marked as paid","type":"success"}}`)
		view := buildMemberDashboardView(store, memberID)
		if view != nil {
			RenderMemberDetail(w, r, view, store)
			return
		}
	}
	http.Redirect(w, r, "/member-detail?member_id="+strconv.Itoa(memberID), http.StatusSeeOther)
}

func HandleMarkContributionPaid(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	memberID, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil || memberID <= 0 {
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	eventID, err := strconv.Atoi(r.FormValue("event_id"))
	if err != nil || eventID <= 0 {
		http.Error(w, "valid event id is required", http.StatusBadRequest)
		return
	}
	if err := store.MarkContributionPaid(memberID, eventID); err != nil {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Trigger", `{"showToast":{"message":"`+err.Error()+`","type":"error"}}`)
			view := buildMemberDashboardView(store, memberID)
			if view != nil {
				RenderMemberDetail(w, r, view, store)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Trigger", `{"showToast":{"message":"Contribution deducted","type":"success"}}`)
		view := buildMemberDashboardView(store, memberID)
		if view != nil {
			RenderMemberDetail(w, r, view, store)
			return
		}
	}
	http.Redirect(w, r, "/member-detail?member_id="+strconv.Itoa(memberID), http.StatusSeeOther)
}

func HandleEvents(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Method not allowed", "error")
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	minAmountExpected, err := strconv.ParseFloat(r.FormValue("min_amount_expected"), 64)
	if err != nil {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, "Valid minimum amount expected is required", "error")
			return
		}
		http.Error(w, "valid minimum amount expected is required", http.StatusBadRequest)
		return
	}
	if err := store.AddEvent(Event{Title: r.FormValue("title"), Description: r.FormValue("description"), Date: r.FormValue("date"), MinAmountExpected: minAmountExpected, Status: "open"}); err != nil {
		if r.Header.Get("HX-Request") != "" {
			renderIndexFragment(w, r, store, err.Error(), "error")
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		renderIndexFragment(w, r, store, "Event created", "success")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleContribution(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Trigger", `{"showToast":{"message":"Method not allowed","type":"error"}}`)
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	eventID, err := strconv.Atoi(r.FormValue("event_id"))
	if err != nil || eventID <= 0 {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Trigger", `{"showToast":{"message":"Valid event id is required","type":"error"}}`)
		}
		http.Error(w, "valid event id is required", http.StatusBadRequest)
		return
	}
	memberID, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil || memberID <= 0 {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Trigger", `{"showToast":{"message":"Valid member id is required","type":"error"}}`)
		}
		http.Error(w, "valid member id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Trigger", `{"showToast":{"message":"Valid amount is required","type":"error"}}`)
		}
		http.Error(w, "valid amount is required", http.StatusBadRequest)
		return
	}
	status := r.FormValue("status")
	if status == "" {
		status = "paid"
	}
	if err := store.AddContribution(Contribution{EventID: eventID, MemberID: memberID, Amount: amount, Status: status}); err != nil {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Trigger", `{"showToast":{"message":"`+err.Error()+`","type":"error"}}`)
			view := buildEventDashboardView(store, eventID, "")
			if view != nil {
				RenderEventDetail(w, r, view)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Trigger", `{"showToast":{"message":"Payment recorded","type":"success"}}`)
		view := buildEventDashboardView(store, eventID, "")
		if view != nil {
			RenderEventDetail(w, r, view)
			return
		}
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
	RenderMemberDetail(w, r, view, store)
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
	filter := r.URL.Query().Get("filter")
	view := buildEventDashboardView(store, eventID, filter)
	if view == nil {
		http.NotFound(w, r)
		return
	}
	RenderEventDetail(w, r, view)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

func HandleAttendanceDetail(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date is required", http.StatusBadRequest)
		return
	}
	filter := r.URL.Query().Get("filter")

	recordMap := make(map[int]AttendanceRecord)
	for _, record := range store.Attendance {
		if record.MeetingDate == date {
			recordMap[record.MemberID] = record
		}
	}

	var records []AttendanceRecordView
	for _, member := range store.Members {
		rec, found := recordMap[member.ID]
		status := "not_recorded"
		note := ""
		if found {
			status = rec.Status
			note = rec.Note
		}
		if filter != "" && status != filter {
			continue
		}
		records = append(records, AttendanceRecordView{MemberName: member.Name, Status: status, Note: note})
	}

	view := AttendanceDetailView{MeetingDate: date, Records: records, Filter: filter}

	tmpl, err := template.ParseFiles("templates/attendance_detail.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if r.Header.Get("HX-Request") != "" {
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

func HandleExportAttendancePDF(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Attendance Report")
	pdf.Ln(12)

	if startDate != "" && endDate != "" {
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("Period: %s to %s", startDate, endDate))
		pdf.Ln(10)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(60, 10, "Member Name", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Date", "1", 0, "", false, 0, "")
	pdf.CellFormat(65, 10, "Status", "1", 0, "", false, 0, "")
	pdf.CellFormat(35, 10, "Active?", "1", 0, "", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 12)
	for _, record := range store.Attendance {
		if startDate != "" && record.MeetingDate < startDate {
			continue
		}
		if endDate != "" && record.MeetingDate > endDate {
			continue
		}
		memberName := "Unknown"
		memberStatus := "Unknown"
		for _, m := range store.Members {
			if m.ID == record.MemberID {
				memberName = m.Name
				memberStatus = m.Status
				break
			}
		}
		pdf.CellFormat(60, 10, memberName, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, record.MeetingDate, "1", 0, "", false, 0, "")
		pdf.CellFormat(65, 10, record.Status, "1", 0, "", false, 0, "")
		pdf.CellFormat(35, 10, memberStatus, "1", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="attendance_report.pdf"`)
	err := pdf.Output(w)
	if err != nil {
		http.Error(w, "failed to generate PDF", http.StatusInternalServerError)
	}
}

func HandleDeleteMember(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	memberIDStr := r.FormValue("member_id")
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}
	err = store.DeleteMember(memberID)
	if err != nil {
		http.Error(w, "Failed to delete member", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func HandleExportContributionsPDF(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	eventIDStr := r.URL.Query().Get("event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil || eventID <= 0 {
		http.Error(w, "valid event id is required", http.StatusBadRequest)
		return
	}

	var targetEvent Event
	found := false
	for _, e := range store.Events {
		if e.ID == eventID {
			targetEvent = e
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Contributions Report: %s", targetEvent.Title))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(60, 10, "Member Name", "1", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "Status", "1", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "Amount", "1", 0, "", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 12)
	for _, member := range store.Members {
		status := "not_paid"
		amount := 0.0
		for _, contrib := range store.Contributions {
			if contrib.EventID == eventID && contrib.MemberID == member.ID {
				status = contrib.Status
				amount = contrib.Amount
				break
			}
		}
		if status == "pending" {
			status = "not_paid"
		}
		if status == "not_paid" {
			amount = 0.0
		}
		pdf.CellFormat(60, 10, member.Name, "1", 0, "", false, 0, "")
		pdf.CellFormat(40, 10, status, "1", 0, "", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", amount), "1", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="contributions_event_%d.pdf"`, eventID))
	if err := pdf.Output(w); err != nil {
		http.Error(w, "failed to generate PDF", http.StatusInternalServerError)
	}
}
