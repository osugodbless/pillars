package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestMemberBalancePartiallyPaid(t *testing.T) {
	store := NewStore()
	store.Dues = append(store.Dues, DuesRecord{MemberID: 1, Amount: 100, Status: "partially_paid"})
	store.Fines = append(store.Fines, Fine{MemberID: 1, Amount: 30, Status: "paid"})

	balance := store.MemberBalance(1)
	if balance != 100 {
		t.Fatalf("expected balance 100, got %.2f", balance)
	}
}

func TestMemberBalanceZero(t *testing.T) {
	store := NewStore()
	balance := store.MemberBalance(999)
	if balance != 0 {
		t.Fatalf("expected balance 0, got %.2f", balance)
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

func TestRecordAttendancePresentNoFine(t *testing.T) {
	store := NewStore()
	err := store.RecordAttendance(1, "2026-07-05", "present", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 0 {
		t.Fatalf("expected no fines for present, got %d", len(store.Fines))
	}
}

func TestRecordAttendanceWithPermissionNoFine(t *testing.T) {
	store := NewStore()
	err := store.RecordAttendance(1, "2026-07-05", "absent_with_permission", "Had appointment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 0 {
		t.Fatalf("expected no fines for permitted absence, got %d", len(store.Fines))
	}
}

func TestRecordAttendanceInvalidMemberID(t *testing.T) {
	store := NewStore()
	err := store.RecordAttendance(0, "2026-07-05", "present", "")
	if err == nil {
		t.Fatal("expected error for member ID 0")
	}
	err = store.RecordAttendance(-1, "2026-07-05", "present", "")
	if err == nil {
		t.Fatal("expected error for negative member ID")
	}
}

func TestRecordAttendanceEmptyStatus(t *testing.T) {
	store := NewStore()
	err := store.RecordAttendance(1, "2026-07-05", "", "")
	if err == nil {
		t.Fatal("expected error for empty status")
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

func TestMemberDashboardSummaryZeroDays(t *testing.T) {
	store := NewStore()
	store.Attendance = []AttendanceRecord{{MemberID: 1, MeetingDate: "", Status: "present"}}
	summary := store.MemberDashboardSummary(1, 0)
	if summary.AttendanceTotal != 1 {
		t.Fatalf("expected 1 attendance total with 0 days, got %d", summary.AttendanceTotal)
	}
}

func TestMemberDashboardSummaryNoMatch(t *testing.T) {
	store := NewStore()
	summary := store.MemberDashboardSummary(999, 30)
	if summary.AttendanceTotal != 0 {
		t.Fatalf("expected 0, got %d", summary.AttendanceTotal)
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

func TestRecordAttendanceAndDuesPaid(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}

	err := recordAttendanceAndDues(store, 1, "2026-07-01", "present", "", true, 2000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Dues) != 1 {
		t.Fatalf("expected one dues record, got %d", len(store.Dues))
	}
	if store.Dues[0].Status != "paid" {
		t.Fatalf("expected paid dues status, got %s", store.Dues[0].Status)
	}
	if store.Dues[0].Amount != 2000 {
		t.Fatalf("expected dues amount 2000, got %.2f", store.Dues[0].Amount)
	}
}

func TestRecordAttendanceAndDuesRecordFails(t *testing.T) {
	store := NewStore()
	err := recordAttendanceAndDues(store, 0, "2026-07-01", "", "", false, 1000)
	if err == nil {
		t.Fatal("expected error from recordAttendanceAndDues")
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

func TestGroupAttendanceByDateEmpty(t *testing.T) {
	groups := groupAttendanceByDate(nil)
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}

func TestGroupAttendanceByDateSkipsEmptyEntries(t *testing.T) {
	records := []AttendanceRecord{
		{MemberID: 1, MeetingDate: "", Status: "present"},
		{MemberID: 2, MeetingDate: "2026-07-01", Status: "present"},
	}
	groups := groupAttendanceByDate(records)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
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

func TestAddEventSkipsExMembers(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Ada", Status: "active"},
		{ID: 2, Name: "Bob", Status: "ex-member"},
	}
	err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Contributions) != 1 {
		t.Fatalf("expected 1 contribution (ex-member skipped), got %d", len(store.Contributions))
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

func TestFormatNaira(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "₦0.00"},
		{100, "₦100.00"},
		{1000, "₦1,000.00"},
		{1234567, "₦1,234,567.00"},
		{1234.56, "₦1,234.56"},
		{-500, "-₦500.00"},
		{0.1, "₦0.10"},
		{99.99, "₦99.99"},
	}

	for _, tt := range tests {
		result := FormatNaira(tt.input)
		if result != tt.expected {
			t.Errorf("FormatNaira(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestAgingBucket(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		date     string
		expected string
	}{
		{"", "0-30"},
		{now.Format("2006-01-02"), "0-30"},
		{now.AddDate(0, 0, -15).Format("2006-01-02"), "0-30"},
		{now.AddDate(0, 0, -31).Format("2006-01-02"), "31-60"},
		{now.AddDate(0, 0, -45).Format("2006-01-02"), "31-60"},
		{now.AddDate(0, 0, -61).Format("2006-01-02"), "61-90"},
		{now.AddDate(0, 0, -75).Format("2006-01-02"), "61-90"},
		{now.AddDate(0, 0, -91).Format("2006-01-02"), "90+"},
		{now.AddDate(0, 0, -120).Format("2006-01-02"), "90+"},
		{"invalid-date", "0-30"},
	}

	for _, tt := range tests {
		result := AgingBucket(tt.date)
		if result != tt.expected {
			t.Errorf("AgingBucket(%q) = %q, want %q", tt.date, result, tt.expected)
		}
	}
}

func TestMemberFinancialSummaries(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Ada", Status: "active"},
		{ID: 2, Name: "Bob", Status: "active"},
		{ID: 3, Name: "Carol", Status: "ex-member"},
	}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-01-15"},
		{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-02-15"},
		{MemberID: 2, Amount: 2000, Status: "paid", DueDate: "2026-01-15"},
		{MemberID: 3, Amount: 2000, Status: "paid", DueDate: "2026-01-15"},
	}
	store.Fines = []Fine{
		{MemberID: 1, Amount: 500, Status: "outstanding", FineDate: "2026-01-20"},
		{MemberID: 2, Amount: 300, Status: "paid", FineDate: "2026-01-25"},
	}
	store.Events = []Event{{ID: 1, Date: "2026-01-10"}}
	store.Contributions = []Contribution{
		{EventID: 1, MemberID: 1, Amount: 1000, Status: "paid"},
		{EventID: 1, MemberID: 2, Amount: 1000, Status: "pending"},
	}

	summaries, err := store.MemberFinancialSummaries("2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 3 {
		t.Fatalf("expected 3 summaries (active + ex-member with records in range), got %d", len(summaries))
	}

	ada := summaries[0]
	if ada.MemberName != "Ada" {
		t.Errorf("expected Ada, got %s", ada.MemberName)
	}
	if ada.DuesExpected != 2000 {
		t.Errorf("expected DuesExpected 2000, got %.2f", ada.DuesExpected)
	}
	if ada.DuesPaid != 2000 {
		t.Errorf("expected DuesPaid 2000, got %.2f", ada.DuesPaid)
	}
	if ada.FinesLevied != 500 {
		t.Errorf("expected FinesLevied 500, got %.2f", ada.FinesLevied)
	}
	if ada.FinesPaid != 0 {
		t.Errorf("expected FinesPaid 0, got %.2f", ada.FinesPaid)
	}
	if ada.ContributionsExpected != 1000 {
		t.Errorf("expected ContributionsExpected 1000, got %.2f", ada.ContributionsExpected)
	}
	if ada.ContributionsPaid != 1000 {
		t.Errorf("expected ContributionsPaid 1000, got %.2f", ada.ContributionsPaid)
	}

	bob := summaries[1]
	if bob.DuesExpected != 2000 {
		t.Errorf("expected Bob DuesExpected 2000, got %.2f", bob.DuesExpected)
	}
	if bob.FinesPaid != 300 {
		t.Errorf("expected Bob FinesPaid 300, got %.2f", bob.FinesPaid)
	}
	if bob.ContributionsPaid != 0 {
		t.Errorf("expected Bob ContributionsPaid 0, got %.2f", bob.ContributionsPaid)
	}
}

func TestMemberFinancialSummariesExMemberWithDuesOutOfRange(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "ex-member"}}
	store.Dues = []DuesRecord{{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2025-01-01"}}

	summaries, err := store.MemberFinancialSummaries("2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries (out of range), got %d", len(summaries))
	}
}

func TestMemberFinancialSummariesPartiallyPaidDues(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	store.Dues = []DuesRecord{{MemberID: 1, Amount: 1500, Deducted: 500, Status: "partially_paid", DueDate: "2026-01-15"}}

	summaries, err := store.MemberFinancialSummaries("2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.DuesExpected != 2000 {
		t.Errorf("expected DuesExpected 2000 (1500+500), got %.2f", s.DuesExpected)
	}
	if s.DuesPaid != 1500 {
		t.Errorf("expected DuesPaid 1500, got %.2f", s.DuesPaid)
	}
	if s.DuesDeducted != 500 {
		t.Errorf("expected DuesDeducted 500, got %.2f", s.DuesDeducted)
	}
}

func TestArrearsByMember(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Ada", Status: "active"},
		{ID: 2, Name: "Bob", Status: "active"},
		{ID: 3, Name: "Carol", Status: "active"},
	}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-01-15"},
		{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-02-15"},
		{MemberID: 2, Amount: 4000, Status: "owed", DueDate: "2025-12-01"},
	}
	store.Fines = []Fine{
		{MemberID: 1, Amount: 500, Status: "outstanding", FineDate: "2026-01-20"},
		{MemberID: 3, Amount: 1000, Status: "outstanding", FineDate: ""},
	}

	arrears, err := store.ArrearsByMember()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(arrears) != 3 {
		t.Fatalf("expected 3 members with debt, got %d", len(arrears))
	}

	if arrears[0].TotalOwed < arrears[1].TotalOwed {
		t.Error("arrears should be sorted by total owed descending")
	}

	bob := arrears[0]
	if bob.MemberName != "Bob" {
		t.Errorf("expected Bob, got %s", bob.MemberName)
	}
	if bob.DuesOwed != 4000 {
		t.Errorf("expected Bob DuesOwed 4000, got %.2f", bob.DuesOwed)
	}
	if bob.TotalOwed != 4000 {
		t.Errorf("expected Bob TotalOwed 4000, got %.2f", bob.TotalOwed)
	}

	ada := arrears[1]
	if ada.DuesOwed != 2000 {
		t.Errorf("expected Ada DuesOwed 2000, got %.2f", ada.DuesOwed)
	}
	if ada.FinesOwed != 500 {
		t.Errorf("expected Ada FinesOwed 500, got %.2f", ada.FinesOwed)
	}
	if ada.TotalOwed != 2500 {
		t.Errorf("expected Ada TotalOwed 2500, got %.2f", ada.TotalOwed)
	}
}

func TestArrearsByMemberNoDebt(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	arrears, err := store.ArrearsByMember()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arrears) != 0 {
		t.Fatalf("expected 0 arrears, got %d", len(arrears))
	}
}

func TestArrearsByMemberWithContributions(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	store.Events = []Event{{ID: 1, Date: "2026-01-10"}}
	store.Contributions = []Contribution{{EventID: 1, MemberID: 1, Amount: 500, Status: "not_paid"}}

	arrears, err := store.ArrearsByMember()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arrears) != 1 {
		t.Fatalf("expected 1 arrear, got %d", len(arrears))
	}
	if arrears[0].ContribOwed != 500 {
		t.Errorf("expected ContribOwed 500, got %.2f", arrears[0].ContribOwed)
	}
}

// === NEW TESTS: Store methods with 0% coverage ===

func TestProbationReviewDue(t *testing.T) {
	store := NewStore()
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	store.Members = []Member{
		{ID: 1, Name: "Ada", Status: "probation", ProbationEnds: yesterday},
		{ID: 2, Name: "Bob", Status: "probation", ProbationEnds: tomorrow},
		{ID: 3, Name: "Carol", Status: "active"},
		{ID: 4, Name: "Dave", Status: "probation", ProbationEnds: today},
	}

	due := store.ProbationReviewDue()
	if len(due) != 2 {
		t.Fatalf("expected 2 members due for review, got %d", len(due))
	}
}

func TestProbationReviewDueEmpty(t *testing.T) {
	store := NewStore()
	due := store.ProbationReviewDue()
	if len(due) != 0 {
		t.Fatalf("expected 0, got %d", len(due))
	}
}

func TestProbationReviewDueEmptyProbationEnds(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "probation", ProbationEnds: ""}}
	due := store.ProbationReviewDue()
	if len(due) != 0 {
		t.Fatalf("expected 0 (empty ProbationEnds), got %d", len(due))
	}
}

func TestPromoteToActiveInMemory(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "probation"}}

	err := store.PromoteToActive(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].Status != "active" {
		t.Fatalf("expected status 'active', got %q", store.Members[0].Status)
	}
	if !store.Members[0].IsBonafide {
		t.Fatal("expected IsBonafide true")
	}
}

func TestPromoteToActiveNotFound(t *testing.T) {
	store := NewStore()
	err := store.PromoteToActive(999)
	if err == nil {
		t.Fatal("expected error for nonexistent member")
	}
}

func TestPromoteToActiveNotOnProbation(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	err := store.PromoteToActive(1)
	if err == nil {
		t.Fatal("expected error for non-probation member")
	}
}

func TestPromoteToActiveWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, err := NewStoreWithPath(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	store.CreateMember(Member{Name: "Ada", Status: "probation", JoinedAt: time.Now().Format(time.RFC3339)})

	err = store.PromoteToActive(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].Status != "active" {
		t.Fatalf("expected status 'active', got %q", store.Members[0].Status)
	}

	reopened, _ := NewStoreWithPath(dbPath)
	if reopened.Members[0].Status != "active" {
		t.Fatalf("expected status 'active' after reopen, got %q", reopened.Members[0].Status)
	}
}

func TestExtendProbationInMemory(t *testing.T) {
	store := NewStore()
	futureDate := time.Now().AddDate(0, 3, 0).Format("2006-01-02")
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "probation", ProbationEnds: futureDate}}

	err := store.ExtendProbation(1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, _ := time.Parse("2006-01-02", futureDate)
	expected := parsed.AddDate(0, 2, 0).Format("2006-01-02")
	if store.Members[0].ProbationEnds != expected {
		t.Fatalf("expected ProbationEnds %s, got %s", expected, store.Members[0].ProbationEnds)
	}
}

func TestExtendProbationEmptyEndDate(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "probation", ProbationEnds: ""}}

	err := store.ExtendProbation(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].ProbationEnds == "" {
		t.Fatal("expected ProbationEnds to be set")
	}
}

func TestExtendProbationNotFound(t *testing.T) {
	store := NewStore()
	err := store.ExtendProbation(999, 1)
	if err == nil {
		t.Fatal("expected error for nonexistent member")
	}
}

func TestExtendProbationNotOnProbation(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	err := store.ExtendProbation(1, 1)
	if err == nil {
		t.Fatal("expected error for non-probation member")
	}
}

func TestExtendProbationInvalidDate(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "probation", ProbationEnds: "not-a-date"}}
	err := store.ExtendProbation(1, 1)
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestExtendProbationWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "probation", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.ExtendProbation(1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reopened, _ := NewStoreWithPath(dbPath)
	if reopened.Members[0].ProbationEnds == "" {
		t.Fatal("expected ProbationEnds to persist")
	}
}

func TestAddDuesInMemory(t *testing.T) {
	store := NewStore()
	err := store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-07-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Dues) != 1 {
		t.Fatalf("expected 1 due, got %d", len(store.Dues))
	}
	if store.Dues[0].ID != 1 {
		t.Fatalf("expected ID 1, got %d", store.Dues[0].ID)
	}
}

func TestAddDuesWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	err := store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-07-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Dues) != 1 {
		t.Fatalf("expected 1 due, got %d", len(store.Dues))
	}
	if store.Dues[0].ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestAddFineInMemory(t *testing.T) {
	store := NewStore()
	err := store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 1 {
		t.Fatalf("expected 1 fine, got %d", len(store.Fines))
	}
	if store.Fines[0].ID != 1 {
		t.Fatalf("expected ID 1, got %d", store.Fines[0].ID)
	}
}

func TestAddFineWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	err := store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 1 {
		t.Fatalf("expected 1 fine, got %d", len(store.Fines))
	}
}

func TestMarkFinePaidInMemory(t *testing.T) {
	store := NewStore()
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}

	err := store.MarkFinePaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected status 'paid', got %q", store.Fines[0].Status)
	}
}

func TestMarkFinePaidNotFound(t *testing.T) {
	store := NewStore()
	err := store.MarkFinePaid(1, 999)
	if err == nil {
		t.Fatal("expected error for nonexistent fine")
	}
}

func TestMarkFinePaidAlreadyPaid(t *testing.T) {
	store := NewStore()
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "paid"}}
	err := store.MarkFinePaid(1, 1)
	if err == nil {
		t.Fatal("expected error for already-paid fine")
	}
}

func TestMarkFinePaidWrongMember(t *testing.T) {
	store := NewStore()
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	err := store.MarkFinePaid(2, 1)
	if err == nil {
		t.Fatal("expected error for wrong member")
	}
}

func TestMarkFinePaidWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})

	err := store.MarkFinePaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected status 'paid', got %q", store.Fines[0].Status)
	}
}

func TestMarkDuesPaidInMemory(t *testing.T) {
	store := NewStore()
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "pending"}}

	err := store.MarkDuesPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "paid" {
		t.Fatalf("expected status 'paid', got %q", store.Dues[0].Status)
	}
}

func TestMarkDuesPaidOwed(t *testing.T) {
	store := NewStore()
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "owed"}}

	err := store.MarkDuesPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "paid" {
		t.Fatalf("expected status 'paid', got %q", store.Dues[0].Status)
	}
}

func TestMarkDuesPaidPartiallyPaid(t *testing.T) {
	store := NewStore()
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 1500, Status: "partially_paid"}}

	err := store.MarkDuesPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "paid" {
		t.Fatalf("expected status 'paid', got %q", store.Dues[0].Status)
	}
}

func TestMarkDuesPaidNotFound(t *testing.T) {
	store := NewStore()
	err := store.MarkDuesPaid(1, 999)
	if err == nil {
		t.Fatal("expected error for nonexistent dues")
	}
}

func TestMarkDuesPaidAlreadyPaid(t *testing.T) {
	store := NewStore()
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}
	err := store.MarkDuesPaid(1, 1)
	if err == nil {
		t.Fatal("expected error for already-paid dues")
	}
}

func TestMarkDuesPaidWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-07-01"})

	err := store.MarkDuesPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "paid" {
		t.Fatalf("expected status 'paid', got %q", store.Dues[0].Status)
	}
}

func TestMarkContributionPaidInMemoryNoPaidDues(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 100, Status: "pending"})

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range store.Contributions {
		if c.EventID == 1 && c.MemberID == 1 {
			if c.Status != "paid" {
				t.Fatalf("expected status 'paid', got %q", c.Status)
			}
			break
		}
	}
}

func TestMarkContributionPaidNotFound(t *testing.T) {
	store := NewStore()
	err := store.MarkContributionPaid(1, 1)
	if err == nil {
		t.Fatal("expected error for nonexistent contribution")
	}
}

func TestMarkContributionPaidWithPaidDuesFullCoverage(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 100, Status: "pending"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "partially_paid" {
		t.Fatalf("expected dues partially_paid, got %q", store.Dues[0].Status)
	}
	if store.Dues[0].Amount != 1900 {
		t.Fatalf("expected dues amount 1900, got %.2f", store.Dues[0].Amount)
	}
	if store.Dues[0].Deducted != 100 {
		t.Fatalf("expected dues deducted 100, got %.2f", store.Dues[0].Deducted)
	}
}

func TestMarkContributionPaidWithPartialDues(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.Settings.DuesAmount = 2000
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 500, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 500, Status: "pending"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 100, Status: "paid", DueDate: "2026-07-01"})

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The paid dues (100) < contribution (500), so dues gets fully consumed, remaining 400 goes to owed
	if store.Dues[0].Status != "pending" {
		t.Fatalf("expected dues reset to pending, got %q", store.Dues[0].Status)
	}
	// Check that an "owed" record was created for the remainder
	foundOwed := false
	for _, d := range store.Dues {
		if d.Status == "owed" && d.MemberID == 1 {
			foundOwed = true
			if d.Amount != 400 {
				t.Fatalf("expected owed amount 400, got %.2f", d.Amount)
			}
		}
	}
	if !foundOwed {
		t.Fatal("expected owed dues record for remainder")
	}
}

func TestMarkContributionPaidWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 100, Status: "pending"})

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkContributionPaidWithDBAndPaidDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 100, Status: "pending"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkEventSettledInMemory(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	// Mark all contributions as paid
	for i := range store.Contributions {
		store.Contributions[i].Status = "paid"
	}

	err := store.MarkEventSettled(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Events[0].Status != "settled" {
		t.Fatalf("expected status 'settled', got %q", store.Events[0].Status)
	}
	for _, c := range store.Contributions {
		if c.EventID == 1 && c.Status != "settled" {
			t.Fatalf("expected contribution status 'settled', got %q", c.Status)
		}
	}
}

func TestMarkEventSettledNotFound(t *testing.T) {
	store := NewStore()
	err := store.MarkEventSettled(999)
	if err == nil {
		t.Fatal("expected error for nonexistent event")
	}
}

func TestMarkEventSettledNotOpen(t *testing.T) {
	store := NewStore()
	store.Events = []Event{{ID: 1, Title: "Event", Status: "settled"}}
	err := store.MarkEventSettled(1)
	if err == nil {
		t.Fatal("expected error for non-open event")
	}
}

func TestMarkEventSettledPendingContributions(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	// Leave contributions as pending
	err := store.MarkEventSettled(1)
	if err == nil {
		t.Fatal("expected error for pending contributions")
	}
}

func TestMarkEventSettledPartiallyPaidContributions(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"
	store.Contributions[1].Status = "partially_paid"

	err := store.MarkEventSettled(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkEventSettledWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	for i := range store.Contributions {
		store.Contributions[i].Status = "paid"
	}

	err := store.MarkEventSettled(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Events[0].Status != "settled" {
		t.Fatalf("expected 'settled', got %q", store.Events[0].Status)
	}
}

func TestDeductFineFromDuesInMemoryNoRemainder(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected fine status 'paid', got %q", store.Fines[0].Status)
	}
	if store.Dues[0].Status != "partially_paid" {
		t.Fatalf("expected dues status 'partially_paid', got %q", store.Dues[0].Status)
	}
	if store.Dues[0].Amount != 1500 {
		t.Fatalf("expected dues amount 1500, got %.2f", store.Dues[0].Amount)
	}
}

func TestDeductFineFromDuesInMemoryFullConsumption(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 2000, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "pending" {
		t.Fatalf("expected dues reset to pending, got %q", store.Dues[0].Status)
	}
}

func TestDeductFineFromDuesInMemoryWithRemainder(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 3000, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundOwed := false
	for _, d := range store.Dues {
		if d.Status == "owed" {
			foundOwed = true
			if d.Amount != 1000 {
				t.Fatalf("expected owed amount 1000, got %.2f", d.Amount)
			}
		}
	}
	if !foundOwed {
		t.Fatal("expected owed record for remainder")
	}
}

func TestDeductFineFromDuesNotFound(t *testing.T) {
	store := NewStore()
	err := store.DeductFineFromDues(1, 999)
	if err == nil {
		t.Fatal("expected error for nonexistent fine")
	}
}

func TestDeductFineFromDuesWithPartiallyPaidDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 100, Status: "partially_paid"}}

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected fine status 'paid', got %q", store.Fines[0].Status)
	}
}

func TestDeductFineFromDuesWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductFineFromDuesWithDBFullConsumption(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 2000, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductFineFromDuesWithDBRemainder(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 3000, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductAllFinesFromDuesInMemory(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Fines = []Fine{
		{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"},
		{ID: 2, MemberID: 1, Amount: 300, Status: "outstanding"},
	}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range store.Fines {
		if f.Status != "paid" {
			t.Fatalf("expected all fines paid, got %q", f.Status)
		}
	}
}

func TestDeductAllFinesFromDuesNoFines(t *testing.T) {
	store := NewStore()
	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductAllFinesFromDuesWithRemainder(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 3000, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundOwed := false
	for _, d := range store.Dues {
		if d.Status == "owed" {
			foundOwed = true
		}
	}
	if !foundOwed {
		t.Fatal("expected owed record for remainder")
	}
}

func TestDeductAllFinesFromDuesWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductAllFinesFromDuesWithDBRemainder(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 3000, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMemberInMemory(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}

	err := store.DeleteMember(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].Status != "ex-member" {
		t.Fatalf("expected status 'ex-member', got %q", store.Members[0].Status)
	}
}

func TestDeleteMemberWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.DeleteMember(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].Status != "ex-member" {
		t.Fatalf("expected 'ex-member', got %q", store.Members[0].Status)
	}

	reopened, _ := NewStoreWithPath(dbPath)
	if reopened.Members[0].Status != "ex-member" {
		t.Fatalf("expected 'ex-member' after reopen, got %q", reopened.Members[0].Status)
	}
}

func TestTotalTreasuryBalance(t *testing.T) {
	store := NewStore()
	store.Dues = []DuesRecord{
		{Amount: 2000, Status: "paid"},
		{Amount: 1500, Status: "partially_paid"},
		{Amount: 1000, Status: "pending"},
	}
	store.Fines = []Fine{
		{Amount: 500, Status: "paid"},
		{Amount: 300, Status: "outstanding"},
	}
	store.Contributions = []Contribution{
		{Amount: 1000, Status: "paid"},
		{Amount: 500, Status: "not_paid"},
	}

	balance := store.TotalTreasuryBalance()
	expected := 2000.0 + 1500.0 + 500.0 + 1000.0
	if balance != expected {
		t.Fatalf("expected treasury balance %.2f, got %.2f", expected, balance)
	}
}

func TestTotalOutstandingReceivables(t *testing.T) {
	store := NewStore()
	store.Dues = []DuesRecord{
		{Amount: 2000, Status: "pending"},
		{Amount: 1000, Status: "owed"},
		{Amount: 500, Status: "paid"},
	}
	store.Fines = []Fine{
		{Amount: 500, Status: "outstanding"},
		{Amount: 300, Status: "paid"},
	}
	store.Contributions = []Contribution{
		{Amount: 1000, Status: "pending"},
		{Amount: 500, Status: "paid"},
	}

	owed := store.TotalOutstandingReceivables()
	expected := 2000.0 + 1000.0 + 500.0 + 1000.0
	if owed != expected {
		t.Fatalf("expected outstanding %.2f, got %.2f", expected, owed)
	}
}

func TestAtRiskMembersCount(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000

	// Member with >3 months dues AND 3 consecutive absence fines
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
	}
	store.Fines = []Fine{
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "2026-01-15"},
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "2026-02-15"},
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "2026-03-15"},
	}

	count := store.AtRiskMembersCount()
	if count != 1 {
		t.Fatalf("expected 1 at-risk member, got %d", count)
	}
}

func TestAtRiskMembersCountNoMatch(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000

	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "pending"},
	}

	count := store.AtRiskMembersCount()
	if count != 0 {
		t.Fatalf("expected 0 at-risk members, got %d", count)
	}
}

func TestAtRiskMembersCountNonConsecutiveFines(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000

	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
	}
	store.Fines = []Fine{
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "2026-01-15"},
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "2026-05-15"},
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "2026-09-15"},
	}

	count := store.AtRiskMembersCount()
	if count != 0 {
		t.Fatalf("expected 0 (non-consecutive), got %d", count)
	}
}

func TestAtRiskMembersCountFinesWithInvalidDate(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Ada", Status: "active"}}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
		{MemberID: 1, Amount: 2000, Status: "pending"},
	}
	store.Fines = []Fine{
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "bad-date"},
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "bad-date"},
		{MemberID: 1, Amount: 1000, Status: "outstanding", Reason: "Unapproved absence", FineDate: "bad-date"},
	}

	count := store.AtRiskMembersCount()
	if count != 0 {
		t.Fatalf("expected 0 (invalid dates), got %d", count)
	}
}

func TestCreateMemberEmptyName(t *testing.T) {
	store := NewStore()
	err := store.CreateMember(Member{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateMemberProbationEndsCalculation(t *testing.T) {
	store := NewStore()
	store.Settings.ProbationPeriodDays = 90
	joinedAt := time.Now().Format(time.RFC3339)

	err := store.CreateMember(Member{Name: "Ada", Status: "probation", JoinedAt: joinedAt})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].ProbationEnds == "" {
		t.Fatal("expected ProbationEnds to be calculated")
	}
}

func TestCreateMemberWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	err := store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestCreateMemberWithDBProbation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	err := store.CreateMember(Member{Name: "Ada", Status: "probation", JoinedAt: time.Now().Format(time.RFC3339)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].ProbationEnds == "" {
		t.Fatal("expected ProbationEnds to be set")
	}
}

func TestNewStoreWithPathEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewStoreWithPath(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNewStoreWithPathDefault(t *testing.T) {
	// Test empty path defaults to ./data/pillars.db
	// We can't easily test this without affecting the real DB, so skip
	t.Skip("skipping default path test to avoid side effects")
}

func TestMemberDashboardSummaryDuesWithDeducted(t *testing.T) {
	store := NewStore()
	store.Attendance = nil
	store.Dues = []DuesRecord{{MemberID: 1, Amount: 1500, Deducted: 500, Status: "partially_paid"}}
	store.Fines = nil
	store.Contributions = nil

	summary := store.MemberDashboardSummary(1, 0)
	if summary.DuesPaid != 1500 {
		t.Errorf("expected DuesPaid 1500, got %.2f", summary.DuesPaid)
	}
	if summary.DuesOwed != 500 {
		t.Errorf("expected DuesOwed 500 (from Deducted), got %.2f", summary.DuesOwed)
	}
}

func TestContainsHelper(t *testing.T) {
	if !contains("duplicate column name", "duplicate column name") {
		t.Error("expected contains to return true for exact match")
	}
	if !contains("abc duplicate column name xyz", "duplicate column name") {
		t.Error("expected contains to return true for substring")
	}
	if contains("abc", "abcde") {
		t.Error("expected contains to return false for shorter needle")
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("expected 1 for true")
	}
	if boolToInt(false) != 0 {
		t.Error("expected 0 for false")
	}
}

func TestAddContributionNoEventMinimum(t *testing.T) {
	store := NewStore()
	// No event, no minimum — should succeed
	err := store.AddContribution(Contribution{EventID: 999, MemberID: 1, Amount: 50, Status: "paid"})
	if err != nil {
		t.Fatalf("expected no error when no event minimum, got: %v", err)
	}
}

func TestAddContributionUpdateExistingPending(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	// First contribution
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "pending"})
	// Update existing pending contribution with amount >= minimum
	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 75, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range store.Contributions {
		if c.EventID == 1 && c.MemberID == 1 {
			if c.Amount != 75 {
				t.Fatalf("expected amount 75, got %.2f", c.Amount)
			}
			if c.Status != "paid" {
				t.Fatalf("expected status 'paid', got %q", c.Status)
			}
		}
	}
}

func TestAddContributionUpdateExistingPendingBelowMin(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Ada"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "pending"})
	// Update with amount below minimum
	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 30, Status: "paid"})
	if err == nil {
		t.Fatal("expected error for amount below minimum on update")
	}
}

func TestAddContributionDBNewInsert(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 100, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddContributionDBUpdateExistingPaid(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "paid"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 30, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range store.Contributions {
		if c.EventID == 1 && c.MemberID == 1 {
			if c.Amount != 80 {
				t.Fatalf("expected amount 80, got %.2f", c.Amount)
			}
		}
	}
}

func TestAddContributionDBUpdateExistingPendingBelowMin(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "pending"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 30, Status: "paid"})
	if err == nil {
		t.Fatal("expected error for below minimum on pending update")
	}
}

func TestAddContributionDBBelowMinNewInsert(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 30, Status: "paid"})
	if err == nil {
		t.Fatal("expected error for below minimum on new insert")
	}
}

func TestAddEventWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Contributions) != 1 {
		t.Fatalf("expected 1 contribution, got %d", len(store.Contributions))
	}
}

func TestRecordAttendanceWithDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	err := store.RecordAttendance(1, "2026-07-05", "present", "On time")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Attendance) != 1 {
		t.Fatalf("expected 1 attendance, got %d", len(store.Attendance))
	}
}

func TestRecordAttendanceWithDBAbsent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	err := store.RecordAttendance(1, "2026-07-05", "absent_without_permission", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 1 {
		t.Fatalf("expected 1 fine for absence, got %d", len(store.Fines))
	}
}

func TestListAttendanceFromDBEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	records, err := store.listAttendanceFromDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}

func TestListDuesFromDBEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	dues, err := store.listDuesFromDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dues) != 0 {
		t.Fatalf("expected 0 dues, got %d", len(dues))
	}
}

func TestListEventsFromDBEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	events, err := store.listEventsFromDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestListContributionsFromDBEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	contribs, err := store.listContributionsFromDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contribs) != 0 {
		t.Fatalf("expected 0 contributions, got %d", len(contribs))
	}
}

func TestListMembersFromDBEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)

	members, err := store.listMembersFromDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 0 {
		t.Fatalf("expected 0 members, got %d", len(members))
	}
}

func TestPersistDuesAndFines(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Late", FineDate: "2026-07-01"})

	reopened, _ := NewStoreWithPath(dbPath)
	if len(reopened.Dues) != 1 {
		t.Fatalf("expected 1 due after reopen, got %d", len(reopened.Dues))
	}
	if len(reopened.Fines) != 1 {
		t.Fatalf("expected 1 fine after reopen, got %d", len(reopened.Fines))
	}
}

func TestPersistEventsAndContributions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Ada", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	reopened, _ := NewStoreWithPath(dbPath)
	if len(reopened.Events) != 1 {
		t.Fatalf("expected 1 event after reopen, got %d", len(reopened.Events))
	}
	if len(reopened.Contributions) != 1 {
		t.Fatalf("expected 1 contribution after reopen, got %d", len(reopened.Contributions))
	}
}

// === Additional coverage tests ===

func TestAddContributionInMemoryAccumulatePaid(t *testing.T) {
	store := NewStore()
	store.AddEvent(Event{ID: 1, Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions = append(store.Contributions, Contribution{ID: 1, EventID: 1, MemberID: 1, Amount: 50, Status: "paid"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 75, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Contributions[0].Amount != 125 {
		t.Fatalf("expected accumulated amount 125, got %.2f", store.Contributions[0].Amount)
	}
}

func TestAddContributionInMemoryAccumulatePartialPaid(t *testing.T) {
	store := NewStore()
	store.AddEvent(Event{ID: 1, Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions = append(store.Contributions, Contribution{ID: 1, EventID: 1, MemberID: 1, Amount: 50, Status: "partially_paid"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 75, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Contributions[0].Amount != 125 {
		t.Fatalf("expected accumulated amount 125, got %.2f", store.Contributions[0].Amount)
	}
}

func TestAddContributionInMemoryUpdatePending(t *testing.T) {
	store := NewStore()
	store.AddEvent(Event{ID: 1, Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions = append(store.Contributions, Contribution{ID: 1, EventID: 1, MemberID: 1, Amount: 200, Status: "pending"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 200, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Contributions[0].Amount != 200 {
		t.Fatalf("expected amount 200, got %.2f", store.Contributions[0].Amount)
	}
	if store.Contributions[0].Status != "paid" {
		t.Fatalf("expected 'paid', got %q", store.Contributions[0].Status)
	}
}

func TestAddContributionInMemoryBelowMinimum(t *testing.T) {
	store := NewStore()
	store.AddEvent(Event{ID: 1, Title: "Event", MinAmountExpected: 100, Status: "open"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "pending"})
	if err == nil {
		t.Fatal("expected error for amount below minimum")
	}
}

func TestAddContributionInMemoryNewBelowMinimum(t *testing.T) {
	store := NewStore()
	store.Events = append(store.Events, Event{ID: 99, Title: "Event", MinAmountExpected: 200, Status: "open"})

	err := store.AddContribution(Contribution{EventID: 99, MemberID: 1, Amount: 50, Status: "pending"})
	if err == nil {
		t.Fatal("expected error for amount below minimum")
	}
}

func TestAddContributionInMemoryNoMinimum(t *testing.T) {
	store := NewStore()

	err := store.AddContribution(Contribution{EventID: 99, MemberID: 1, Amount: 0, Status: "pending"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordAttendanceEmptyDate(t *testing.T) {
	store := NewStore()
	err := store.RecordAttendance(1, "", "present", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemberBalanceWithPaidFines(t *testing.T) {
	store := NewStore()
	store.Dues = append(store.Dues, DuesRecord{MemberID: 1, Amount: 100, Status: "paid"})
	store.Fines = append(store.Fines, Fine{MemberID: 1, Amount: 30, Status: "paid"})
	balance := store.MemberBalance(1)
	if balance != 100 {
		t.Fatalf("expected balance 100 (paid fines don't reduce balance), got %.2f", balance)
	}
}

func TestEventTitleNotFound(t *testing.T) {
	store := NewStore()
	title := store.EventTitle(999)
	if title != "Event #999" {
		t.Fatalf("expected 'Event #999', got %q", title)
	}
}

func TestExtendProbationEmptyEnds(t *testing.T) {
	store := NewStore()
	store.Members = append(store.Members, Member{ID: 1, Name: "Bob", Status: "probation"})
	err := store.ExtendProbation(1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkEventSettledNotAllContributed(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"
	store.Contributions[1].Status = "pending"

	err := store.MarkEventSettled(1)
	if err == nil {
		t.Fatal("expected error when not all contributed")
	}
}

func TestAtRiskMembersCountNone(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	count := store.AtRiskMembersCount()
	if count != 0 {
		t.Fatalf("expected 0 at-risk members, got %d", count)
	}
}

func TestArrearsByMemberNoExMembers(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "ex-member"}}
	rows, err := store.ArrearsByMember()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 arrears rows, got %d", len(rows))
	}
}

func TestMarkDuesPaidNotFoundOwed(t *testing.T) {
	store := NewStore()
	store.Dues = append(store.Dues, DuesRecord{ID: 1, MemberID: 1, Amount: 2000, Status: "owed"})
	err := store.MarkDuesPaid(2, 1)
	if err == nil {
		t.Fatal("expected error for wrong member ID")
	}
}

func TestDeleteMemberNotFound(t *testing.T) {
	store := NewStore()
	err := store.DeleteMember(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddContributionInMemoryUpdatePendingBelowMinimum(t *testing.T) {
	store := NewStore()
	store.AddEvent(Event{ID: 1, Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions = append(store.Contributions, Contribution{ID: 1, EventID: 1, MemberID: 1, Amount: 200, Status: "pending"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "paid"})
	if err == nil {
		t.Fatal("expected error for amount below minimum on pending update")
	}
}

func TestMemberDashboardSummaryNoRecords(t *testing.T) {
	store := NewStore()
	store.Members = append(store.Members, Member{ID: 1, Name: "Alice"})
	summary := store.MemberDashboardSummary(1, 30)
	if summary.AttendanceTotal != 0 {
		t.Fatalf("expected 0 attendance, got %d", summary.AttendanceTotal)
	}
}

func TestGroupAttendanceByDateEmptyMeetingDate(t *testing.T) {
	records := []AttendanceRecord{{MeetingDate: "", Status: "present"}}
	groups := groupAttendanceByDate(records)
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups for empty date, got %d", len(groups))
	}
}

func TestGroupAttendanceByDateMultipleDates(t *testing.T) {
	records := []AttendanceRecord{
		{MeetingDate: "2026-07-01", Status: "present"},
		{MeetingDate: "2026-07-01", Status: "present"},
		{MeetingDate: "2026-07-08", Status: "absent"},
	}
	groups := groupAttendanceByDate(records)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Count != 2 {
		t.Fatalf("expected first group count 2, got %d", groups[0].Count)
	}
}

func TestContains(t *testing.T) {
	if !contains("hello world", "world") {
		t.Fatal("expected true")
	}
	if contains("hello", "world") {
		t.Fatal("expected false")
	}
	if !contains("hello", "hello") {
		t.Fatal("expected true for exact match")
	}
	if contains("", "abc") {
		t.Fatal("expected false for empty text")
	}
	if !contains("abc", "") {
		t.Fatal("expected true for empty needle")
	}
}

func TestMemberFinancialSummariesEmpty(t *testing.T) {
	store := NewStore()
	summaries, err := store.MemberFinancialSummaries("2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestMarkDuesPaidAlreadyPaidCov(t *testing.T) {
	store := NewStore()
	store.Dues = append(store.Dues, DuesRecord{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"})
	err := store.MarkDuesPaid(1, 1)
	if err == nil {
		t.Fatal("expected error for already-paid dues")
	}
}

// === DB-backed coverage tests ===

func TestRecordAttendanceAbsentWithoutPermissionDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.Settings.AbsenceFineAmount = 1000
	err := store.RecordAttendance(1, "2026-07-01", "absent_without_permission", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 1 {
		t.Fatalf("expected 1 auto-fine, got %d", len(store.Fines))
	}
	if store.Fines[0].Amount != 1000 {
		t.Fatalf("expected fine amount 1000, got %.2f", store.Fines[0].Amount)
	}
}

func TestRecordAttendancePresentNoFineDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	err := store.RecordAttendance(1, "2026-07-01", "present", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 0 {
		t.Fatalf("expected 0 fines, got %d", len(store.Fines))
	}
}

func TestRecordAttendanceAbsentWithoutPermissionInMemory(t *testing.T) {
	store := NewStore()
	store.Settings.AbsenceFineAmount = 500
	err := store.RecordAttendance(1, "2026-07-01", "absent_without_permission", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 1 {
		t.Fatalf("expected 1 auto-fine, got %d", len(store.Fines))
	}
}

func TestAddContributionDBExistingPaid(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	existing := store.Contributions[0]
	existing.Amount = 50
	existing.Status = "paid"
	store.db.Exec(`UPDATE contributions SET amount = ?, status = ? WHERE id = ?`, existing.Amount, existing.Status, existing.ID)

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 75, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range store.Contributions {
		if c.EventID == 1 && c.MemberID == 1 && c.Status == "paid" {
			if c.Amount != 125 {
				t.Fatalf("expected accumulated amount 125, got %.2f", c.Amount)
			}
			return
		}
	}
}

func TestAddContributionDBExistingPartiallyPaid(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	existing := store.Contributions[0]
	store.db.Exec(`UPDATE contributions SET amount = ?, status = ? WHERE id = ?`, 50.0, "partially_paid", existing.ID)
	store.Contributions[0].Amount = 50
	store.Contributions[0].Status = "partially_paid"

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 75, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddContributionDBExistingPendingBelowMinimum(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 200, Status: "open"})

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "paid"})
	if err == nil {
		t.Fatal("expected error for amount below minimum")
	}
}

func TestAddContributionDBNewBelowMinimum(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 500, Status: "open"})

	// Remove auto-generated contribution
	store.db.Exec(`DELETE FROM contributions WHERE event_id = 1`)
	store.Contributions = nil

	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 50, Status: "pending"})
	if err == nil {
		t.Fatal("expected error for amount below minimum")
	}
}

func TestAddContributionDBNewNoEvent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.AddContribution(Contribution{EventID: 99, MemberID: 1, Amount: 50, Status: "pending"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkContributionPaidDBFullDuesCoverage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 1000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 500, Status: "open"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-06-01"})

	// Set contribution to pending
	store.db.Exec(`UPDATE contributions SET status = 'pending', amount = 1500 WHERE event_id = 1`)
	store.Contributions[0].Status = "pending"
	store.Contributions[0].Amount = 1500

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After paying, the dues should be reset to pending
	foundPending := false
	for _, d := range store.Dues {
		if d.Status == "pending" || d.Status == "owed" {
			foundPending = true
			break
		}
	}
	if !foundPending {
		t.Fatal("expected dues to be reset or carried forward")
	}
}

func TestMarkContributionPaidDBPartialDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 1000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-06-01"})

	// Contribution amount is less than dues
	store.db.Exec(`UPDATE contributions SET status = 'pending', amount = 300 WHERE event_id = 1`)
	store.Contributions[0].Status = "pending"
	store.Contributions[0].Amount = 300

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundPartial := false
	for _, d := range store.Dues {
		if d.Status == "partially_paid" {
			foundPartial = true
			break
		}
	}
	if !foundPartial {
		t.Fatal("expected partially_paid dues")
	}
}

func TestMarkContributionPaidDBNoPaidDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 1000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	store.db.Exec(`UPDATE contributions SET status = 'pending', amount = 500 WHERE event_id = 1`)
	store.Contributions[0].Status = "pending"
	store.Contributions[0].Amount = 500

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create an owed dues record for the remaining
	foundOwed := false
	for _, d := range store.Dues {
		if d.Status == "owed" {
			foundOwed = true
			break
		}
	}
	if !foundOwed {
		t.Fatal("expected owed dues record")
	}
}

func TestMarkContributionPaidDBPartiallyPaidDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 1000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	// Add a partially_paid dues record
	store.AddDues(DuesRecord{MemberID: 1, Amount: 500, Status: "partially_paid", DueDate: "2026-06-01"})
	// Also add a paid one
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-07-01"})

	store.db.Exec(`UPDATE contributions SET status = 'pending', amount = 200 WHERE event_id = 1`)
	store.Contributions[0].Status = "pending"
	store.Contributions[0].Amount = 200

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkEventSettledDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	// Mark contribution as paid
	store.db.Exec(`UPDATE contributions SET status = 'paid' WHERE event_id = 1`)
	store.Contributions[0].Status = "paid"

	err := store.MarkEventSettled(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Events[0].Status != "settled" {
		t.Fatalf("expected settled, got %s", store.Events[0].Status)
	}
}

func TestDeleteMemberDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.DeleteMember(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].Status != "ex-member" {
		t.Fatalf("expected ex-member, got %s", store.Members[0].Status)
	}
}

func TestArrearsByMemberDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "owed", DueDate: "2026-06-01"})

	rows, err := store.ArrearsByMember()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 arrears row, got %d", len(rows))
	}
}

func TestMemberFinancialSummariesDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 200, Status: "open"})

	summaries, err := store.MemberFinancialSummaries("2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
}

func TestTotalTreasuryBalanceDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	balance := store.TotalTreasuryBalance()
	if balance != 2000 {
		t.Fatalf("expected 2000, got %.2f", balance)
	}
}

func TestTotalOutstandingReceivablesDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "owed", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	receivables := store.TotalOutstandingReceivables()
	if receivables != 2500 {
		t.Fatalf("expected 2500, got %.2f", receivables)
	}
}

func TestAtRiskMembersCountDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 100
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	// 4 months of owed dues (> 3 * 100 = 300)
	store.AddDues(DuesRecord{MemberID: 1, Amount: 100, Status: "owed", DueDate: time.Now().AddDate(0, -3, 0).Format("2006-01-02")})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 100, Status: "owed", DueDate: time.Now().AddDate(0, -2, 0).Format("2006-01-02")})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 100, Status: "owed", DueDate: time.Now().AddDate(0, -1, 0).Format("2006-01-02")})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 100, Status: "owed", DueDate: time.Now().Format("2006-01-02")})
	// 3 consecutive months of unapproved absence fines
	store.AddFine(Fine{MemberID: 1, Amount: 50, Status: "outstanding", Reason: "Unapproved absence", FineDate: time.Now().AddDate(0, -2, 0).Format("2006-01-02")})
	store.AddFine(Fine{MemberID: 1, Amount: 50, Status: "outstanding", Reason: "Unapproved absence", FineDate: time.Now().AddDate(0, -1, 0).Format("2006-01-02")})
	store.AddFine(Fine{MemberID: 1, Amount: 50, Status: "outstanding", Reason: "Unapproved absence", FineDate: time.Now().Format("2006-01-02")})

	count := store.AtRiskMembersCount()
	if count != 1 {
		t.Fatalf("expected 1 at-risk member, got %d", count)
	}
}

func TestMemberDashboardSummaryDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.RecordAttendance(1, "2026-07-01", "present", "")
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})

	summary := store.MemberDashboardSummary(1, 30)
	if summary.AttendanceTotal != 1 {
		t.Fatalf("expected 1 attendance, got %d", summary.AttendanceTotal)
	}
}

func TestMemberBalanceDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1500, Status: "owed", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	balance := store.MemberBalance(1)
	if balance != 1500 {
		t.Fatalf("expected 1500 (2000 paid - 500 fine outstanding), got %.2f", balance)
	}
}

func TestMarkFinePaidDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	err := store.MarkFinePaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected paid, got %s", store.Fines[0].Status)
	}
}

func TestMarkDuesPaidDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-07-01"})

	err := store.MarkDuesPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "paid" {
		t.Fatalf("expected paid, got %s", store.Dues[0].Status)
	}
}

func TestDeductFineFromDuesDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected fine paid, got %s", store.Fines[0].Status)
	}
}

func TestDeductAllFinesFromDuesDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 300, Status: "outstanding", Reason: "Test2", FineDate: "2026-07-02"})

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range store.Fines {
		if f.Status != "paid" {
			t.Fatalf("expected fine paid, got %s", f.Status)
		}
	}
}

func TestAddDuesDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-07-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Dues) != 1 {
		t.Fatalf("expected 1 dues record, got %d", len(store.Dues))
	}
}

func TestAddFineDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Fines) != 1 {
		t.Fatalf("expected 1 fine, got %d", len(store.Fines))
	}
}

func TestPromoteToActiveDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "probation", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.PromoteToActive(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Members[0].Status != "active" {
		t.Fatalf("expected active, got %s", store.Members[0].Status)
	}
}

func TestExtendProbationDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "probation", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.ExtendProbation(1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProbationReviewDueDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "probation", JoinedAt: time.Now().AddDate(0, -6, 0).Format(time.RFC3339)})

	due := store.ProbationReviewDue()
	if len(due) != 1 {
		t.Fatalf("expected 1 member, got %d", len(due))
	}
}

func TestMarkContributionPaidDBMultipleDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 1000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	// Add multiple paid dues
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-05-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-06-01"})

	// Large contribution that covers multiple dues
	store.db.Exec(`UPDATE contributions SET status = 'pending', amount = 2500 WHERE event_id = 1`)
	store.Contributions[0].Status = "pending"
	store.Contributions[0].Amount = 2500

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgingBucketDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	// Dues with different ages
	store.AddDues(DuesRecord{MemberID: 1, Amount: 100, Status: "owed", DueDate: time.Now().AddDate(0, 0, -15).Format("2006-01-02")})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 200, Status: "owed", DueDate: time.Now().AddDate(0, 0, -45).Format("2006-01-02")})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 300, Status: "owed", DueDate: time.Now().AddDate(0, 0, -75).Format("2006-01-02")})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 400, Status: "owed", DueDate: time.Now().AddDate(0, 0, -100).Format("2006-01-02")})

	rows, err := store.ArrearsByMember()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].TotalOwed != 1000 {
		t.Fatalf("expected total 1000, got %.2f", rows[0].TotalOwed)
	}
}

func TestDeductFineFromDuesInMemoryPartialDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "partially_paid", Deducted: 500}}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 300, Status: "outstanding"}}

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected fine paid, got %s", store.Fines[0].Status)
	}
}

func TestDeductFineFromDuesInMemoryFullDuesCoverage(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Dues = []DuesRecord{
		{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"},
		{ID: 2, MemberID: 1, Amount: 2000, Status: "paid"},
	}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 3000, Status: "outstanding"}}

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fine 3000: first dues 2000 → pending (remaining=1000), second dues 2000 → partially_paid (remaining=0)
	if store.Dues[0].Status != "pending" {
		t.Fatalf("expected first dues pending, got %s", store.Dues[0].Status)
	}
	if store.Dues[1].Status != "partially_paid" {
		t.Fatalf("expected second dues partially_paid, got %s", store.Dues[1].Status)
	}
}

func TestDeductAllFinesFromDuesInMemoryPartialDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "partially_paid", Deducted: 500}}
	store.Fines = []Fine{
		{ID: 1, MemberID: 1, Amount: 300, Status: "outstanding"},
		{ID: 2, MemberID: 1, Amount: 200, Status: "outstanding"},
	}

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range store.Fines {
		if f.Status != "paid" {
			t.Fatalf("expected fine paid, got %s", f.Status)
		}
	}
}

func TestDeductAllFinesFromDuesInMemoryFullDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Dues = []DuesRecord{
		{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"},
	}
	store.Fines = []Fine{
		{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"},
	}

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "partially_paid" {
		t.Fatalf("expected partially_paid, got %s", store.Dues[0].Status)
	}
}

func TestDeductAllFinesFromDuesInMemoryNoDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Fines = []Fine{
		{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"},
	}

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundOwed := false
	for _, d := range store.Dues {
		if d.Status == "owed" {
			foundOwed = true
		}
	}
	if !foundOwed {
		t.Fatal("expected owed dues for remainder")
	}
}

func TestDeductAllFinesFromDuesDBPartialDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 2000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "partially_paid", DueDate: "2026-06-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 300, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 200, Status: "outstanding", Reason: "Test2", FineDate: "2026-07-02"})

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductFineFromDuesDBPartialDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 2000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "partially_paid", DueDate: "2026-06-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 300, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductFineFromDuesDBFullCoverage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.Settings.DuesAmount = 1000
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-06-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "paid", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 2500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	err := store.DeductFineFromDues(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeductAllFinesFromDuesDBNoDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	err := store.DeductAllFinesFromDues(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddContributionDBPendingUpdate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	// Set contribution to pending
	store.db.Exec(`UPDATE contributions SET status = 'pending', amount = 100 WHERE event_id = 1`)
	store.Contributions[0].Status = "pending"
	store.Contributions[0].Amount = 100

	// Update with amount >= minimum
	err := store.AddContribution(Contribution{EventID: 1, MemberID: 1, Amount: 200, Status: "paid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddEventDBWithMembers(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.CreateMember(Member{Name: "Bob", Status: "ex-member", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Active member should get a contribution, ex-member should not
	aliceCount := 0
	bobCount := 0
	for _, c := range store.Contributions {
		if c.MemberID == 1 {
			aliceCount++
		}
		if c.MemberID == 2 {
			bobCount++
		}
	}
	if aliceCount != 1 {
		t.Fatalf("expected 1 contribution for Alice, got %d", aliceCount)
	}
	if bobCount != 0 {
		t.Fatalf("expected 0 contributions for Bob (ex-member), got %d", bobCount)
	}
}

func TestAddDuesDBMultipleDues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-07-01"})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1500, Status: "pending", DueDate: "2026-08-01"})

	if len(store.Dues) != 2 {
		t.Fatalf("expected 2 dues records, got %d", len(store.Dues))
	}
}

func TestAddFineDBMultipleFines(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Lateness", FineDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 300, Status: "outstanding", Reason: "Absence", FineDate: "2026-07-02"})

	if len(store.Fines) != 2 {
		t.Fatalf("expected 2 fines, got %d", len(store.Fines))
	}
}

func TestMarkEventSettledDBWithPaidContributions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.db.Exec(`UPDATE contributions SET status = 'paid' WHERE event_id = 1`)
	store.Contributions[0].Status = "paid"

	err := store.MarkEventSettled(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Events[0].Status != "settled" {
		t.Fatalf("expected settled, got %s", store.Events[0].Status)
	}
	if store.Contributions[0].Status != "settled" {
		t.Fatalf("expected contribution settled, got %s", store.Contributions[0].Status)
	}
}

func TestRecordAttendanceDBPresent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})

	err := store.RecordAttendance(1, "2026-07-01", "present", "Good attendance")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Attendance) != 1 {
		t.Fatalf("expected 1 attendance, got %d", len(store.Attendance))
	}
	if store.Attendance[0].Note != "Good attendance" {
		t.Fatalf("expected note, got %s", store.Attendance[0].Note)
	}
}

func TestArrearsByMemberDBWithMultipleMembers(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.CreateMember(Member{Name: "Bob", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddDues(DuesRecord{MemberID: 1, Amount: 1000, Status: "owed", DueDate: time.Now().AddDate(0, 0, -15).Format("2006-01-02")})
	store.AddDues(DuesRecord{MemberID: 2, Amount: 500, Status: "owed", DueDate: time.Now().AddDate(0, -2, 0).Format("2006-01-02")})
	store.AddFine(Fine{MemberID: 1, Amount: 300, Status: "outstanding", Reason: "Test", FineDate: time.Now().AddDate(0, 0, -10).Format("2006-01-02")})

	rows, err := store.ArrearsByMember()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestMarkContributionPaidInMemoryPartiallyPaidDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Dues = []DuesRecord{
		{ID: 1, MemberID: 1, Amount: 2000, Status: "partially_paid", Deducted: 500},
	}

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkContributionPaidInMemoryFullDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Dues = []DuesRecord{
		{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"},
	}

	// Contribution is 100 (minAmountExpected), dues is 2000
	// 100 < 2000, so dues becomes partially_paid
	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "partially_paid" {
		t.Fatalf("expected partially_paid, got %s", store.Dues[0].Status)
	}
}

func TestMarkContributionPaidInMemoryFullDuesCoverage(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 1000
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	// Contribution is 500 (MinAmountExpected on event), dues is 200
	// 500 > 200, so dues resets to pending and remaining goes to next or owed
	store.Contributions[0].Amount = 500
	store.Dues = []DuesRecord{
		{ID: 1, MemberID: 1, Amount: 200, Status: "paid"},
	}

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Dues[0].Status != "pending" {
		t.Fatalf("expected pending, got %s", store.Dues[0].Status)
	}
}

func TestMarkContributionPaidInMemoryMultipleDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Dues = []DuesRecord{
		{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"},
		{ID: 2, MemberID: 1, Amount: 2000, Status: "paid"},
	}

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkContributionPaidInMemoryNoDues(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	err := store.MarkContributionPaid(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundOwed := false
	for _, d := range store.Dues {
		if d.Status == "owed" {
			foundOwed = true
		}
	}
	if !foundOwed {
		t.Fatal("expected owed dues for remainder")
	}
}

func TestMemberDashboardSummaryAllBranches(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.Attendance = []AttendanceRecord{
		{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"},
		{MemberID: 1, MeetingDate: "2026-07-08", Status: "absent_without_permission"},
		{MemberID: 1, MeetingDate: "", Status: "present"},
		{MemberID: 2, MeetingDate: "2026-07-01", Status: "present"},
	}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "paid"},
		{MemberID: 1, Amount: 1000, Status: "partially_paid", Deducted: 500},
		{MemberID: 1, Amount: 500, Status: "pending"},
		{MemberID: 1, Amount: 300, Status: "owed"},
	}
	store.Fines = []Fine{
		{MemberID: 1, Amount: 500, Status: "outstanding"},
		{MemberID: 1, Amount: 200, Status: "paid"},
	}
	store.Contributions = []Contribution{
		{MemberID: 1, EventID: 1, Amount: 100, Status: "paid"},
		{MemberID: 1, EventID: 2, Amount: 200, Status: "partially_paid"},
		{MemberID: 1, EventID: 3, Amount: 300, Status: "settled"},
		{MemberID: 1, EventID: 4, Amount: 400, Status: "pending"},
	}

	summary := store.MemberDashboardSummary(1, 30)
	if summary.AttendanceTotal != 2 {
		t.Fatalf("expected 2 attendance (empty date skipped with days=30), got %d", summary.AttendanceTotal)
	}
	if summary.AttendancePresent != 1 {
		t.Fatalf("expected 1 present, got %d", summary.AttendancePresent)
	}
	if summary.AttendanceAbsent != 1 {
		t.Fatalf("expected 1 absent, got %d", summary.AttendanceAbsent)
	}
	if summary.DuesPaid != 3000 {
		t.Fatalf("expected dues paid 3000, got %.2f", summary.DuesPaid)
	}
	if summary.DuesOwed != 1300 {
		t.Fatalf("expected dues owed 1300 (500 pending + 300 owed + 500 deducted), got %.2f", summary.DuesOwed)
	}
	if summary.FinesOwed != 500 {
		t.Fatalf("expected fines owed 500, got %.2f", summary.FinesOwed)
	}
	if summary.FinesPaid != 200 {
		t.Fatalf("expected fines paid 200, got %.2f", summary.FinesPaid)
	}
	if summary.ContributionsPaid != 600 {
		t.Fatalf("expected contributions paid 600, got %.2f", summary.ContributionsPaid)
	}
	if summary.ContributionsOwed != 400 {
		t.Fatalf("expected contributions owed 400, got %.2f", summary.ContributionsOwed)
	}
}

func TestMemberDashboardSummaryAllBranchesZeroDays(t *testing.T) {
	store := NewStore()
	store.Attendance = []AttendanceRecord{
		{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"},
		{MemberID: 1, MeetingDate: "", Status: "present"},
	}
	summary := store.MemberDashboardSummary(1, 0)
	if summary.AttendanceTotal != 2 {
		t.Fatalf("expected 2 attendance with days=0, got %d", summary.AttendanceTotal)
	}
}

func TestMemberDashboardSummaryOtherMember(t *testing.T) {
	store := NewStore()
	store.Attendance = []AttendanceRecord{
		{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"},
	}
	summary := store.MemberDashboardSummary(2, 30)
	if summary.AttendanceTotal != 0 {
		t.Fatalf("expected 0 for other member, got %d", summary.AttendanceTotal)
	}
}

func TestMemberFinancialSummariesCoversExMemberAndOutOfRange(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Alice", Status: "active"},
	}
	// Fine for non-existent member → ex-member fallback (line 329-336)
	store.Fines = []Fine{
		{MemberID: 1, Amount: 500, Status: "outstanding", FineDate: "2026-07-01"},
		{MemberID: 99, Amount: 300, Status: "outstanding", FineDate: "2026-07-01"},
		{MemberID: 1, Amount: 200, Status: "outstanding", FineDate: "2026-01-01"},
	}
	// Contribution for non-existent member → ex-member fallback (line 353-360)
	store.Contributions = []Contribution{
		{EventID: 1, MemberID: 99, Amount: 100, Status: "paid"},
		{EventID: 1, MemberID: 1, Amount: 200, Status: "paid"},
	}
	store.Events = []Event{{ID: 1, Title: "E", Date: "2026-07-01"}}
	// Dues for non-existent member → ex-member fallback (line 309-315)
	store.Dues = []DuesRecord{
		{MemberID: 99, Amount: 1000, Status: "pending", DueDate: "2026-07-01"},
		{MemberID: 1, Amount: 500, Status: "pending", DueDate: "2026-07-01"},
	}

	summaries, err := store.MemberFinancialSummaries("2026-06-01", "2026-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
}
