package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMemberBalanceUsesDuesPaidMinusFinesOwed(t *testing.T) {
	store := NewStore()
	store.Settings.AbsenceFineAmount = 25

	store.Dues = append(store.Dues, DuesRecord{MemberID: 1, Amount: 100, Status: "paid"})
	store.Dues = append(store.Dues, DuesRecord{MemberID: 1, Amount: 50, Status: "pending"})
	store.Fines = append(store.Fines, Fine{MemberID: 1, Amount: 30, Status: "outstanding"})

	balance := store.MemberBalance(1)
	if balance != 70 {
		t.Fatalf("expected balance 70 (100 paid - 30 fines), got %.2f", balance)
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
	store.Events = []Event{{ID: 1, Title: "Test Event", MinAmountExpected: 50}}
	store.Attendance = []AttendanceRecord{{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"}, {MemberID: 1, MeetingDate: "2026-07-02", Status: "absent_without_permission"}}
	store.Dues = []DuesRecord{{MemberID: 1, Amount: 100, Status: "paid"}, {MemberID: 1, Amount: 50, Status: "pending"}}
	store.Fines = []Fine{{MemberID: 1, Amount: 20, Status: "paid", Reason: "late coming"}, {MemberID: 1, Amount: 30, Status: "outstanding", Reason: "misconduct"}}
	store.Contributions = []Contribution{{EventID: 1, MemberID: 1, Amount: 40, Status: "paid"}, {EventID: 1, MemberID: 1, Amount: 25, Status: "pending"}}

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

func TestAddEventCreatesPendingContributionsForAllMembers(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Ada"},
		{ID: 2, Name: "Bob"},
	}

	err := store.AddEvent(Event{Title: "Fundraiser", MinAmountExpected: 100, Status: "open"})
	if err != nil {
		t.Fatalf("add event returned error: %v", err)
	}

	if len(store.Contributions) != 2 {
		t.Fatalf("expected 2 contributions (one per member), got %d", len(store.Contributions))
	}

	for _, c := range store.Contributions {
		if c.Amount != 100 {
			t.Fatalf("expected contribution amount 100, got %.2f", c.Amount)
		}
		if c.Status != "pending" {
			t.Fatalf("expected pending contribution status, got %s", c.Status)
		}
	}
}

func TestAddContributionRejectsAmountBelowMinimum(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	if err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"}); err != nil {
		t.Fatalf("add event: %v", err)
	}

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 30, Status: "paid"})
	if err == nil {
		t.Fatal("expected error for amount below minimum, got nil")
	}
}

func TestAddContributionAllowsAmountAboveMinimum(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	if err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"}); err != nil {
		t.Fatalf("add event: %v", err)
	}

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 75, Status: "paid"})
	if err != nil {
		t.Fatalf("expected no error for amount above minimum, got: %v", err)
	}
}

func TestEventTitle(t *testing.T) {
	store := NewStore()
	store.Events = []Event{{ID: 1, Title: "Fundraiser"}, {ID: 2, Title: "Party"}}

	if title := store.EventTitle(1); title != "Fundraiser" {
		t.Fatalf("expected 'Fundraiser', got %q", title)
	}
	if title := store.EventTitle(2); title != "Party" {
		t.Fatalf("expected 'Party', got %q", title)
	}
	if title := store.EventTitle(99); title != "Event #99" {
		t.Fatalf("expected 'Event #99', got %q", title)
	}
}

func TestAddContributionAccumulatesOnSubsequentPayment(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	if err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"}); err != nil {
		t.Fatalf("add event: %v", err)
	}

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "paid"})
	if err != nil {
		t.Fatalf("first contribution: %v", err)
	}

	err = store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 30, Status: "paid"})
	if err != nil {
		t.Fatalf("second contribution: %v", err)
	}

	found := false
	for _, c := range store.Contributions {
		if c.EventID == 1 && c.MemberID == 1 {
			if c.Amount != 80 {
				t.Fatalf("expected accumulated amount 80, got %.2f", c.Amount)
			}
			if c.Status != "paid" {
				t.Fatalf("expected status 'paid', got %q", c.Status)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("contribution not found")
	}
}

func TestAddContributionRejectsBelowMinOnFirstPayment(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	if err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"}); err != nil {
		t.Fatalf("add event: %v", err)
	}

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 30, Status: "paid"})
	if err == nil {
		t.Fatal("expected error for first payment below minimum, got nil")
	}
}

func TestAddContributionAllowsSmallSubsequentPaymentAfterMinimumMet(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	if err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"}); err != nil {
		t.Fatalf("add event: %v", err)
	}

	if err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "paid"}); err != nil {
		t.Fatalf("first payment: %v", err)
	}

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 5, Status: "paid"})
	if err != nil {
		t.Fatalf("subsequent small payment should be allowed, got: %v", err)
	}

	for _, c := range store.Contributions {
		if c.EventID == 1 && c.MemberID == 1 {
			if c.Amount != 55 {
				t.Fatalf("expected accumulated amount 55, got %.2f", c.Amount)
			}
			break
		}
	}
}
