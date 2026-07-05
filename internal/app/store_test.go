package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMemberBalanceIncludesOutstandingDuesAndFines(t *testing.T) {
	store := NewStore()
	store.Settings.AbsenceFineAmount = 25

	store.Dues = append(store.Dues, DuesRecord{MemberID: 1, Amount: 100, Status: "pending"})
	store.Fines = append(store.Fines, Fine{MemberID: 1, Amount: 30, Status: "outstanding"})

	balance := store.MemberBalance(1)
	if balance != 130 {
		t.Fatalf("expected balance 130, got %.2f", balance)
	}
}

func TestRecordAttendanceCreatesFineForUnapprovedAbsence(t *testing.T) {
	store := NewStore()
	store.Settings.AbsenceFineAmount = 15

	err := store.RecordAttendance(1, "2026-07-05", "absent_without_permission", "No notice")
	if err != nil {
		t.Fatalf("record attendance returned error: %v", err)
	}

	if len(store.Fines) != 1 {
		t.Fatalf("expected one fine to be created, got %d", len(store.Fines))
	}

	fine := store.Fines[0]
	if fine.MemberID != 1 || fine.Amount != 15 || fine.Status != "outstanding" || fine.FineDate != "2026-07-05" {
		t.Fatalf("unexpected fine: %+v", fine)
	}
}

func TestMemberDashboardSummaryAggregatesMemberData(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.Attendance = []AttendanceRecord{{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"}, {MemberID: 1, MeetingDate: "2026-07-02", Status: "absent_without_permission"}}
	store.Dues = []DuesRecord{{MemberID: 1, Amount: 100, Status: "paid"}, {MemberID: 1, Amount: 50, Status: "pending"}}
	store.Fines = []Fine{{MemberID: 1, Amount: 20, Status: "paid", Reason: "late coming"}, {MemberID: 1, Amount: 30, Status: "outstanding", Reason: "misconduct"}}
	store.Contributions = []Contribution{{MemberID: 1, Amount: 40, Status: "paid"}, {MemberID: 1, Amount: 25, Status: "pending"}}

	summary := store.MemberDashboardSummary(1, 30)
	if summary.AttendancePresent != 1 {
		t.Fatalf("expected 1 present attendance, got %d", summary.AttendancePresent)
	}
	if summary.DuesPaid != 100 {
		t.Fatalf("expected dues paid 100, got %.2f", summary.DuesPaid)
	}
	if summary.DuesOwed != 50 {
		t.Fatalf("expected dues owed 50, got %.2f", summary.DuesOwed)
	}
	if summary.ContributionsPaid != 40 {
		t.Fatalf("expected contributions paid 40, got %.2f", summary.ContributionsPaid)
	}
	if summary.ContributionsOwed != 25 {
		t.Fatalf("expected contributions owed 25, got %.2f", summary.ContributionsOwed)
	}
	if summary.FinesOwed != 30 {
		t.Fatalf("expected fines owed 30, got %.2f", summary.FinesOwed)
	}
}

func TestRecordAttendanceAndDuesMarksUnpaidDuesAsPending(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}

	err := recordAttendanceAndDues(store, 1, "2026-07-01", "absent_with_permission", "", false, 1000)
	if err != nil {
		t.Fatalf("record attendance and dues returned error: %v", err)
	}
	if len(store.Dues) != 1 {
		t.Fatalf("expected one dues record when dues are not paid, got %d", len(store.Dues))
	}
	if store.Dues[0].Status != "pending" {
		t.Fatalf("expected pending dues status, got %s", store.Dues[0].Status)
	}
}

func TestAttendanceStatusFromSelection(t *testing.T) {
	if got := attendanceStatusFromSelection(false, false); got != "absent_with_permission" {
		t.Fatalf("expected absent_with_permission, got %s", got)
	}
	if got := attendanceStatusFromSelection(true, false); got != "present" {
		t.Fatalf("expected present, got %s", got)
	}
	if got := attendanceStatusFromSelection(false, true); got != "absent_without_permission" {
		t.Fatalf("expected absent_without_permission, got %s", got)
	}
	if got := attendanceStatusFromSelection(true, true); got != "present" {
		t.Fatalf("expected present to take precedence when both are checked, got %s", got)
	}
}

func TestListFinesHandlesNullFineDate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, err := NewStoreWithPath(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	if _, err := store.db.Exec(`INSERT INTO fines(member_id, amount, status, reason, fine_date) VALUES (?, ?, ?, ?, ?)`, 1, 10.0, "outstanding", "reason", nil); err != nil {
		t.Fatalf("insert fine with null date: %v", err)
	}

	fines, err := store.listFinesFromDB()
	if err != nil {
		t.Fatalf("list fines returned error: %v", err)
	}
	if len(fines) != 1 {
		t.Fatalf("expected one fine, got %d", len(fines))
	}
	if fines[0].FineDate != "" {
		t.Fatalf("expected empty fine date, got %q", fines[0].FineDate)
	}
}

func TestGroupAttendanceByDateGroupsMonthlyBatches(t *testing.T) {
	records := []AttendanceRecord{
		{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"},
		{MemberID: 2, MeetingDate: "2026-07-01", Status: "present"},
		{MemberID: 3, MeetingDate: "2026-08-01", Status: "absent_with_permission"},
	}

	groups := groupAttendanceByDate(records)
	if len(groups) != 2 {
		t.Fatalf("expected 2 monthly groups, got %d", len(groups))
	}
	if groups[0].MeetingDate != "2026-07-01" || groups[0].Count != 2 {
		t.Fatalf("unexpected first group: %+v", groups[0])
	}
	if groups[1].MeetingDate != "2026-08-01" || groups[1].Count != 1 {
		t.Fatalf("unexpected second group: %+v", groups[1])
	}
}

func TestStorePersistsMembersAndAttendance(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")

	store, err := NewStoreWithPath(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	member := Member{Name: "Ada", Email: "ada@example.com", Phone: "555-0100", Status: "probation"}
	if err := store.CreateMember(member); err != nil {
		t.Fatalf("create member: %v", err)
	}

	if err := store.RecordAttendance(1, "2026-07-05", "present", ""); err != nil {
		t.Fatalf("record attendance: %v", err)
	}

	reopened, err := NewStoreWithPath(dbPath)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}

	if len(reopened.Members) != 1 {
		t.Fatalf("expected one member after reload, got %d", len(reopened.Members))
	}
	if len(reopened.Attendance) != 1 {
		t.Fatalf("expected one attendance record after reload, got %d", len(reopened.Attendance))
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database file should exist: %v", err)
	}
}
