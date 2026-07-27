package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	wd, _ := os.Getwd()
	repoRoot := filepath.Join(wd, "..", "..")
	if err := os.Chdir(repoRoot); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func newRequest(method, url string, body string) *http.Request {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, url, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return r
}

func newHTMXRequest(method, url string, body string) *http.Request {
	r := newRequest(method, url, body)
	r.Header.Set("HX-Request", "true")
	return r
}

func recCode(rec *httptest.ResponseRecorder) int {
	return rec.Code
}

func recBody(rec *httptest.ResponseRecorder) string {
	return rec.Body.String()
}

func recHeader(rec *httptest.ResponseRecorder, key string) string {
	return rec.Header().Get(key)
}

// === HandleHealth ===

func TestHandleHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/health", "")
	HandleHealth(rec, r)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
	if !strings.Contains(recBody(rec), "ok") {
		t.Fatalf("expected 'ok', got %q", recBody(rec))
	}
}

// === HandleIndex ===

func TestHandleIndexGET(t *testing.T) {
	store := NewStore()
	store.Fines = append(store.Fines, Fine{Amount: 500, Status: "paid", Reason: "Test"})
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/", "")
	HandleIndex(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
	if recHeader(rec, "Content-Type") == "" {
		t.Fatal("expected Content-Type header")
	}
}

func TestHandleIndexMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/", "")
	HandleIndex(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleIndexHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/", "")
	HandleIndex(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === HandleMembers ===

func TestHandleMembersPOST(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/members", "name=Alice&email=alice@test.com&phone=123&status=active")
	HandleMembers(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302 redirect, got %d", recCode(rec))
	}
	if len(store.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(store.Members))
	}
}

func TestHandleMembersPOSTHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/members", "name=Alice&status=active")
	HandleMembers(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
	if len(store.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(store.Members))
	}
}

func TestHandleMembersGET(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/members", "")
	HandleMembers(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleMembersGETHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/members", "")
	HandleMembers(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleMembersEmptyName(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/members", "name=")
	HandleMembers(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMembersEmptyNameHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/members", "name=")
	HandleMembers(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

// === HandleAttendance ===

func TestHandleAttendanceMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/attendance", "")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleAttendanceMethodNotAllowedHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/attendance", "")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAttendanceEmptyDate(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance", "meeting_date=")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAttendanceEmptyDateHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAttendanceSingleMember(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance", "meeting_date=2026-07-01&member_id=1&status=present")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
}

func TestHandleAttendanceSingleMemberInvalidID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance", "meeting_date=2026-07-01&member_id=abc&status=present")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAttendanceSingleMemberInvalidIDHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=2026-07-01&member_id=abc&status=present")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAttendanceSingleMemberEmptyStatus(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance", "meeting_date=2026-07-01&member_id=1&status=")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAttendanceSingleMemberEmptyStatusHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=2026-07-01&member_id=1&status=")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAttendanceBatch(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Alice", Status: "active"},
		{ID: 2, Name: "Bob", Status: "active"},
	}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance", "meeting_date=2026-07-01&present_1=1&dues_1=1&absenteeism_2=1&late_2=1")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if len(store.Attendance) != 2 {
		t.Fatalf("expected 2 attendance records, got %d", len(store.Attendance))
	}
}

func TestHandleAttendanceBatchHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=2026-07-01&present_1=1&dues_1=1")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAttendanceBatchExMemberSkipped(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Alice", Status: "active"},
		{ID: 2, Name: "Bob", Status: "ex-member"},
	}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance", "meeting_date=2026-07-01&present_1=1")
	HandleAttendance(rec, r, store)
	if len(store.Attendance) != 1 {
		t.Fatalf("expected 1 attendance (ex-member skipped), got %d", len(store.Attendance))
	}
}

// === HandleDues ===

func TestHandleDuesMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/dues", "")
	HandleDues(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleDuesInvalidMemberID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/dues", "member_id=abc&amount=100")
	HandleDues(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDuesInvalidMemberIDZero(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/dues", "member_id=0&amount=100")
	HandleDues(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDuesInvalidAmount(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/dues", "member_id=1&amount=abc")
	HandleDues(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDuesSuccess(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/dues", "member_id=1&amount=2000&due_date=2026-07-01")
	HandleDues(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if len(store.Dues) != 1 {
		t.Fatalf("expected 1 dues record, got %d", len(store.Dues))
	}
}

// === HandleAddFine ===

func TestHandleAddFineMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/fines", "")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleAddFineMethodNotAllowedHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/fines", "")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAddFineInvalidMemberID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/fines", "member_id=abc&amount=500&reason=Late")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAddFineInvalidMemberIDHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/fines", "member_id=abc&amount=500&reason=Late")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAddFineInvalidAmount(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/fines", "member_id=1&amount=abc&reason=Late")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAddFineInvalidAmountHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/fines", "member_id=1&amount=abc&reason=Late")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAddFineZeroAmount(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/fines", "member_id=1&amount=0&reason=Late")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAddFineEmptyReason(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/fines", "member_id=1&amount=500&reason=")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAddFineEmptyReasonHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/fines", "member_id=1&amount=500&reason=")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleAddFineSuccess(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/fines", "member_id=1&amount=500&reason=Late&fine_date=2026-07-01")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if len(store.Fines) != 1 {
		t.Fatalf("expected 1 fine, got %d", len(store.Fines))
	}
}

func TestHandleAddFineSuccessHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/fines", "member_id=1&amount=500&reason=Late&fine_date=2026-07-01")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

// === HandleEvents ===

func TestHandleEventsMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/events", "")
	HandleEvents(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleEventsMethodNotAllowedHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/events", "")
	HandleEvents(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleEventsInvalidAmount(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/events", "title=Event&min_amount_expected=abc")
	HandleEvents(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleEventsInvalidAmountHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/events", "title=Event&min_amount_expected=abc")
	HandleEvents(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleEventsSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/events", "title=Fundraiser&description=Test&date=2026-07-01&min_amount_expected=100")
	HandleEvents(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if len(store.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(store.Events))
	}
}

func TestHandleEventsSuccessHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/events", "title=Fundraiser&date=2026-07-01&min_amount_expected=100")
	HandleEvents(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

// === HandleContribution ===

func TestHandleContributionMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/contributions", "")
	HandleContribution(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleContributionMethodNotAllowedHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/contributions", "")
	HandleContribution(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleContributionInvalidEventID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/contributions", "event_id=abc&member_id=1&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleContributionInvalidEventIDHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contributions", "event_id=abc&member_id=1&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 405/400, got %d", recCode(rec))
	}
}

func TestHandleContributionInvalidMemberID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/contributions", "event_id=1&member_id=abc&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleContributionInvalidMemberIDHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contributions", "event_id=1&member_id=abc&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleContributionInvalidAmount(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/contributions", "event_id=1&member_id=1&amount=abc")
	HandleContribution(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleContributionInvalidAmountHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contributions", "event_id=1&member_id=1&amount=abc")
	HandleContribution(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleContributionSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/contributions", "event_id=1&member_id=1&amount=100&status=paid")
	HandleContribution(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
}

func TestHandleContributionSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contributions", "event_id=1&member_id=1&amount=100&status=paid")
	HandleContribution(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleContributionDefaultStatus(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 50, Status: "open"})
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/contributions", "event_id=1&member_id=1&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
}

func TestHandleContributionErrorHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contributions", "event_id=999&member_id=1&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 (no event found, redirect), got %d", recCode(rec))
	}
}

func TestHandleContributionErrorHTMXViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contributions", "event_id=999&member_id=1&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect (no event found), got %d", recCode(rec))
	}
}

// === HandleMemberDetail ===

func TestHandleMemberDetailGET(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/member-detail?member_id=1", "")
	HandleMemberDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMemberDetailHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/member-detail?member_id=1", "")
	HandleMemberDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMemberDetailNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/member-detail?member_id=999", "")
	HandleMemberDetail(rec, r, store)
	if recCode(rec) != 404 {
		t.Fatalf("expected 404, got %d", recCode(rec))
	}
}

func TestHandleMemberDetailInvalidID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/member-detail?member_id=abc", "")
	HandleMemberDetail(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMemberDetailMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/member-detail?member_id=1", "")
	HandleMemberDetail(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

// === HandleEventDetail ===

func TestHandleEventDetailGET(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/event-detail?event_id=1", "")
	HandleEventDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleEventDetailHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/event-detail?event_id=1", "")
	HandleEventDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleEventDetailNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/event-detail?event_id=999", "")
	HandleEventDetail(rec, r, store)
	if recCode(rec) != 404 {
		t.Fatalf("expected 404, got %d", recCode(rec))
	}
}

func TestHandleEventDetailInvalidID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/event-detail?event_id=abc", "")
	HandleEventDetail(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleEventDetailMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/event-detail?event_id=1", "")
	HandleEventDetail(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleEventDetailWithFilter(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/event-detail?event_id=1&filter=paid", "")
	HandleEventDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === HandleSettleEvent ===

func TestHandleSettleEventMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/settle-event", "")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleSettleEventInvalidEventID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/settle-event", "event_id=abc")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleSettleEventInvalidEventIDHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/settle-event", "event_id=abc")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleSettleEventSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	for i := range store.Contributions {
		store.Contributions[i].Status = "paid"
	}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/settle-event", "event_id=1")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if store.Events[0].Status != "settled" {
		t.Fatalf("expected 'settled', got %q", store.Events[0].Status)
	}
}

func TestHandleSettleEventSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	for i := range store.Contributions {
		store.Contributions[i].Status = "paid"
	}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/settle-event", "event_id=1")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleSettleEventErrorHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/settle-event", "event_id=999")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleSettleEventNotOpenHTMX(t *testing.T) {
	store := NewStore()
	store.Events = []Event{{ID: 1, Title: "Event", Status: "settled"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/settle-event", "event_id=1")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200 (re-renders with error toast), got %d", recCode(rec))
	}
}

// === HandlePromoteToActive ===

func TestHandlePromoteToActiveMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/promote-to-active", "")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveInvalidID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/promote-to-active", "member_id=abc")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveInvalidIDHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/promote-to-active", "member_id=abc")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveInvalidIDHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/promote-to-active", "member_id=abc&source=index")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/promote-to-active", "member_id=1")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if store.Members[0].Status != "active" {
		t.Fatalf("expected 'active', got %q", store.Members[0].Status)
	}
}

func TestHandlePromoteToActiveSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/promote-to-active", "member_id=1")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveSuccessHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/promote-to-active", "member_id=1&source=index")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveErrorNotOnProbation(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/promote-to-active", "member_id=1")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveErrorHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/promote-to-active", "member_id=1")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveErrorHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/promote-to-active", "member_id=1&source=index")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === HandleExtendProbation ===

func TestHandleExtendProbationMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/extend-probation", "")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationInvalidID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/extend-probation", "member_id=abc&months=1")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationInvalidIDHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=abc&months=1")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationInvalidIDHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=abc&months=1&source=index")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationInvalidMonths(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/extend-probation", "member_id=1&months=abc")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationInvalidMonthsHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=abc")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationMonthsTooHigh(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/extend-probation", "member_id=1&months=5")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationMonthsTooHighHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=5")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationMonthsTooHighHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=5&source=index")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation", ProbationEnds: "2026-10-01"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/extend-probation", "member_id=1&months=2")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation", ProbationEnds: "2026-10-01"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=1")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationSuccessHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation", ProbationEnds: "2026-10-01"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=1&source=index")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationErrorNotOnProbation(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/extend-probation", "member_id=1&months=1")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationErrorHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=1")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationErrorHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=1&source=index")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === HandleMarkFinePaid ===

func TestHandleMarkFinePaidMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/mark-fine-paid", "")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidInvalidMemberID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-fine-paid", "member_id=abc&fine_id=1")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidInvalidFineID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-fine-paid", "member_id=1&fine_id=abc")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidSuccess(t *testing.T) {
	store := NewStore()
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-fine-paid", "member_id=1&fine_id=1")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if store.Fines[0].Status != "paid" {
		t.Fatalf("expected 'paid', got %q", store.Fines[0].Status)
	}
}

func TestHandleMarkFinePaidSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-fine-paid", "member_id=1&fine_id=1")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-fine-paid", "member_id=1&fine_id=999")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidNotFoundHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-fine-paid", "member_id=1&fine_id=999")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidNotFoundHTMXViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-fine-paid", "member_id=999&fine_id=999")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

// === HandleMarkDuesPaid ===

func TestHandleMarkDuesPaidMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/mark-dues-paid", "")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidInvalidMemberID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-dues-paid", "member_id=abc&dues_id=1")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidInvalidDuesID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-dues-paid", "member_id=1&dues_id=abc")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidSuccess(t *testing.T) {
	store := NewStore()
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "pending"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-dues-paid", "member_id=1&dues_id=1")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
	if store.Dues[0].Status != "paid" {
		t.Fatalf("expected 'paid', got %q", store.Dues[0].Status)
	}
}

func TestHandleMarkDuesPaidSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "pending"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-dues-paid", "member_id=1&dues_id=1")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-dues-paid", "member_id=1&dues_id=999")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidNotFoundHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-dues-paid", "member_id=1&dues_id=999")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidNotFoundHTMXViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-dues-paid", "member_id=999&dues_id=999")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

// === HandleMarkContributionPaid ===

func TestHandleMarkContributionPaidMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/mark-contribution-paid", "")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidInvalidMemberID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-contribution-paid", "member_id=abc&event_id=1")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidInvalidEventID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-contribution-paid", "member_id=1&event_id=abc")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-contribution-paid", "member_id=1&event_id=1")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-contribution-paid", "member_id=1&event_id=1")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-contribution-paid", "member_id=1&event_id=999")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidNotFoundHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-contribution-paid", "member_id=1&event_id=999")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidNotFoundHTMXViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-contribution-paid", "member_id=999&event_id=999")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

// === HandleDeductFine ===

func TestHandleDeductFineMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/deduct-fine", "")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleDeductFineInvalidMemberID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/deduct-fine", "member_id=abc&fine_id=1")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDeductFineInvalidFineID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/deduct-fine", "member_id=1&fine_id=abc")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDeductFineSuccess(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/deduct-fine", "member_id=1&fine_id=1")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 302, got %d", recCode(rec))
	}
}

func TestHandleDeductFineSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=1&fine_id=1")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleDeductFineAllFinesSuccess(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/deduct-fine", "member_id=1")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
}

func TestHandleDeductFineAllFinesSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Settings.DuesAmount = 2000
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Fines = []Fine{{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Dues = []DuesRecord{{ID: 1, MemberID: 1, Amount: 2000, Status: "paid"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=1")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleDeductFineAllFinesHTMXViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=999")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect, got %d", recCode(rec))
	}
}

func TestHandleDeductFineNotFoundHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=1&fine_id=999")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200 (re-renders detail with error toast), got %d", recCode(rec))
	}
}

func TestHandleDeductFineAllFinesNotFoundHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=999")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect, got %d", recCode(rec))
	}
}

func TestHandleDeductFineNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/deduct-fine", "member_id=1&fine_id=999")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDeductFineAllFinesNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/deduct-fine", "member_id=999")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect, got %d", recCode(rec))
	}
}

// === HandleDeleteMember ===

func TestHandleDeleteMemberMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/delete-member", "")
	HandleDeleteMember(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleDeleteMemberInvalidID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/delete-member", "member_id=abc")
	HandleDeleteMember(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDeleteMemberSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/delete-member", "member_id=1")
	HandleDeleteMember(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
	if recHeader(rec, "HX-Redirect") != "/" {
		t.Fatalf("expected HX-Redirect '/', got %q", recHeader(rec, "HX-Redirect"))
	}
	if store.Members[0].Status != "ex-member" {
		t.Fatalf("expected 'ex-member', got %q", store.Members[0].Status)
	}
}

// === HandleAttendanceDetail ===

func TestHandleAttendanceDetailGET(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.Attendance = []AttendanceRecord{{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"}}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/attendance-detail?date=2026-07-01", "")
	HandleAttendanceDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleAttendanceDetailHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/attendance-detail?date=2026-07-01", "")
	HandleAttendanceDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleAttendanceDetailMissingDate(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/attendance-detail", "")
	HandleAttendanceDetail(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleAttendanceDetailMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance-detail?date=2026-07-01", "")
	HandleAttendanceDetail(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleAttendanceDetailWithFilter(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.Attendance = []AttendanceRecord{
		{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"},
		{MemberID: 2, MeetingDate: "2026-07-01", Status: "absent_without_permission"},
	}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/attendance-detail?date=2026-07-01&filter=present", "")
	HandleAttendanceDetail(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === HandleExportAttendancePDF ===

func TestHandleExportAttendancePDFGET(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.Attendance = []AttendanceRecord{{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"}}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-attendance", "")
	HandleExportAttendancePDF(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
	if recHeader(rec, "Content-Type") != "application/pdf" {
		t.Fatalf("expected PDF content type, got %q", recHeader(rec, "Content-Type"))
	}
}

func TestHandleExportAttendancePDFMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/export-attendance", "")
	HandleExportAttendancePDF(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleExportAttendancePDFWithDateRange(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.Attendance = []AttendanceRecord{
		{MemberID: 1, MeetingDate: "2026-05-01", Status: "present"},
		{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"},
		{MemberID: 1, MeetingDate: "2026-09-01", Status: "present"},
	}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-attendance?start_date=2026-07-01&end_date=2026-07-31", "")
	HandleExportAttendancePDF(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === HandleExportContributionsPDF ===

func TestHandleExportContributionsPDFGET(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-contributions?event_id=1", "")
	HandleExportContributionsPDF(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExportContributionsPDFMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/export-contributions", "")
	HandleExportContributionsPDF(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleExportContributionsPDFInvalidID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-contributions?event_id=abc", "")
	HandleExportContributionsPDF(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleExportContributionsPDFNotFound(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-contributions?event_id=999", "")
	HandleExportContributionsPDF(rec, r, store)
	if recCode(rec) != 404 {
		t.Fatalf("expected 404, got %d", recCode(rec))
	}
}

func TestHandleExportContributionsPDFPendingStatus(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-contributions?event_id=1", "")
	HandleExportContributionsPDF(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === HandleCommitteeReport ===

func TestHandleCommitteeReportGET(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/reports/committee", "")
	HandleCommitteeReport(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleCommitteeReportMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/reports/committee", "")
	HandleCommitteeReport(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleCommitteeReportWithDates(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/reports/committee?from=2026-01-01&to=2026-12-31", "")
	HandleCommitteeReport(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleCommitteeReportPartialDates(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/reports/committee?from=2026-01-01", "")
	HandleCommitteeReport(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleCommitteeReportFromAfterTo(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/reports/committee?from=2026-12-31&to=2026-01-01", "")
	HandleCommitteeReport(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

// === HandleArrearsReport ===

func TestHandleArrearsReportGET(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/reports/arrears", "")
	HandleArrearsReport(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleArrearsReportMethodNotAllowed(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/reports/arrears", "")
	HandleArrearsReport(rec, r, store)
	if recCode(rec) != 405 {
		t.Fatalf("expected 405, got %d", recCode(rec))
	}
}

func TestHandleArrearsReportWithData(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Alice", Status: "active"},
		{ID: 2, Name: "Bob", Status: "active"},
		{ID: 3, Name: "Carol", Status: "active"},
	}
	store.Dues = []DuesRecord{
		{MemberID: 1, Amount: 2000, Status: "pending", DueDate: "2026-01-01"},
		{MemberID: 2, Amount: 1000, Status: "owed", DueDate: time.Now().AddDate(0, 0, -45).Format("2006-01-02")},
		{MemberID: 3, Amount: 1500, Status: "pending", DueDate: time.Now().AddDate(0, 0, -75).Format("2006-01-02")},
	}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/reports/arrears", "")
	HandleArrearsReport(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === Render functions ===

func TestRenderMembersNew(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/members/new", "")
	RenderMembersNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderMembersNewHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/members/new", "")
	RenderMembersNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderEventsNew(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/events/new", "")
	RenderEventsNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderEventsNewHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/events/new", "")
	RenderEventsNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderFinesNew(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob", Status: "ex-member"}}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/fines/new", "")
	RenderFinesNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderFinesNewHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/fines/new", "")
	RenderFinesNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderAttendanceNew(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob", Status: "ex-member"}}
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/attendance/new", "")
	RenderAttendanceNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderAttendanceNewHTMX(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/attendance/new", "")
	RenderAttendanceNew(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === buildPageData ===

func TestBuildPageData(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Alice", Status: "active"},
		{ID: 2, Name: "Bob", Status: "probation"},
	}
	store.Events = []Event{{ID: 1, Title: "Event", Status: "open", MinAmountExpected: 100}}
	store.Dues = []DuesRecord{{MemberID: 1, Amount: 2000, Status: "paid"}}
	store.Fines = []Fine{{MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Contributions = []Contribution{{EventID: 1, MemberID: 1, Amount: 100, Status: "paid"}}

	data := buildPageData(store)
	if data.Stats.TotalMembers != 2 {
		t.Fatalf("expected 2 total members, got %d", data.Stats.TotalMembers)
	}
	if data.Stats.ProbationMembers != 1 {
		t.Fatalf("expected 1 probation member, got %d", data.Stats.ProbationMembers)
	}
	if data.Stats.OpenEvents != 1 {
		t.Fatalf("expected 1 open event, got %d", data.Stats.OpenEvents)
	}
	if data.Stats.OutstandingFines != 1 {
		t.Fatalf("expected 1 outstanding fine, got %d", data.Stats.OutstandingFines)
	}
}

func TestBuildPageDataExMembersFiltered(t *testing.T) {
	store := NewStore()
	store.Members = []Member{
		{ID: 1, Name: "Alice", Status: "active"},
		{ID: 2, Name: "Bob", Status: "ex-member"},
	}
	data := buildPageData(store)
	if len(data.Members) != 1 {
		t.Fatalf("expected 1 member (ex-member filtered), got %d", len(data.Members))
	}
}

func TestBuildPageDataEventFundingProgress(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.Events = []Event{{ID: 1, Title: "Event", Status: "open", MinAmountExpected: 100}}
	store.Contributions = []Contribution{{EventID: 1, MemberID: 1, Amount: 50, Status: "paid"}}

	data := buildPageData(store)
	if len(data.Stats.EventFundingProgress) != 1 {
		t.Fatalf("expected 1 event funding progress, got %d", len(data.Stats.EventFundingProgress))
	}
}

func TestBuildPageDataFundingPercentageOver100(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.Events = []Event{{ID: 1, Title: "Event", Status: "open", MinAmountExpected: 100}}
	store.Contributions = []Contribution{{EventID: 1, MemberID: 1, Amount: 200, Status: "paid"}}

	data := buildPageData(store)
	if data.Stats.EventFundingProgress[0].Percentage != 100 {
		t.Fatalf("expected percentage capped at 100, got %.2f", data.Stats.EventFundingProgress[0].Percentage)
	}
}

func TestBuildPageDataNoSettledEvents(t *testing.T) {
	store := NewStore()
	store.Events = []Event{{ID: 1, Title: "Event", Status: "settled", MinAmountExpected: 100}}

	data := buildPageData(store)
	if len(data.Stats.EventFundingProgress) != 0 {
		t.Fatalf("expected 0 funding progress for settled events, got %d", len(data.Stats.EventFundingProgress))
	}
}

func TestBuildPageDataUnknownMemberName(t *testing.T) {
	store := NewStore()
	store.Attendance = []AttendanceRecord{{MemberID: 999, MeetingDate: "2026-07-01", Status: "present"}}

	data := buildPageData(store)
	if data.Attendance[0].MemberName != "Unknown" {
		t.Fatalf("expected 'Unknown' member name, got %q", data.Attendance[0].MemberName)
	}
}

// === buildMemberDashboardView ===

func TestBuildMemberDashboardView(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Attendance = []AttendanceRecord{{MemberID: 1, MeetingDate: "2026-07-01", Status: "present"}}
	store.Dues = []DuesRecord{{MemberID: 1, Amount: 2000, Status: "paid"}}
	store.Fines = []Fine{{MemberID: 1, Amount: 500, Status: "outstanding"}}
	store.Contributions = []Contribution{{EventID: 1, MemberID: 1, Amount: 100, Status: "paid"}}

	view := buildMemberDashboardView(store, 1)
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if view.Member.Name != "Alice" {
		t.Fatalf("expected 'Alice', got %q", view.Member.Name)
	}
	if len(view.Attendance) != 1 {
		t.Fatalf("expected 1 attendance, got %d", len(view.Attendance))
	}
}

func TestBuildMemberDashboardViewNotFound(t *testing.T) {
	store := NewStore()
	view := buildMemberDashboardView(store, 999)
	if view != nil {
		t.Fatal("expected nil view for nonexistent member")
	}
}

// === buildEventDashboardView ===

func TestBuildEventDashboardView(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"

	view := buildEventDashboardView(store, 1, "")
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if view.Event.Title != "Event" {
		t.Fatalf("expected 'Event', got %q", view.Event.Title)
	}
	if view.TotalCollected != 100 {
		t.Fatalf("expected collected 100, got %.2f", view.TotalCollected)
	}
}

func TestBuildEventDashboardViewNotFound(t *testing.T) {
	store := NewStore()
	view := buildEventDashboardView(store, 999, "")
	if view != nil {
		t.Fatal("expected nil view for nonexistent event")
	}
}

func TestBuildEventDashboardViewFilterPaid(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"
	store.Contributions[1].Status = "pending"

	view := buildEventDashboardView(store, 1, "paid")
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if len(view.Members) != 1 {
		t.Fatalf("expected 1 member (paid filter), got %d", len(view.Members))
	}
}

func TestBuildEventDashboardViewFilterNotPaid(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"

	view := buildEventDashboardView(store, 1, "not_paid")
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if len(view.Members) != 1 {
		t.Fatalf("expected 1 member (not_paid filter), got %d", len(view.Members))
	}
}

func TestBuildEventDashboardViewFilterOther(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"

	view := buildEventDashboardView(store, 1, "pending")
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if len(view.Members) != 0 {
		t.Fatalf("expected 0 members (pending filter, none match), got %d", len(view.Members))
	}
}

func TestBuildEventDashboardViewCanSettle(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"

	view := buildEventDashboardView(store, 1, "")
	if !view.CanSettle {
		t.Fatal("expected CanSettle true")
	}
}

func TestBuildEventDashboardViewCannotSettlePending(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	// Leave contribution as pending

	view := buildEventDashboardView(store, 1, "")
	if view.CanSettle {
		t.Fatal("expected CanSettle false (pending contributions)")
	}
}

func TestBuildEventDashboardViewCannotSettleNotOpen(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "settled"})

	view := buildEventDashboardView(store, 1, "")
	if view.CanSettle {
		t.Fatal("expected CanSettle false (not open)")
	}
}

func TestBuildEventDashboardViewMemberNotInContributions(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	// Only Alice has a contribution
	store.Contributions = store.Contributions[:1]

	view := buildEventDashboardView(store, 1, "")
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	// Bob should show as "not_paid"
	found := false
	for _, m := range view.Members {
		if m.Name == "Bob" {
			if m.Status != "not_paid" {
				t.Fatalf("expected Bob status 'not_paid', got %q", m.Status)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("expected Bob in members list")
	}
}

func TestBuildEventDashboardViewFilterAll(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	view := buildEventDashboardView(store, 1, "all")
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if len(view.Members) != 1 {
		t.Fatalf("expected 1 member (all filter), got %d", len(view.Members))
	}
}

// === renderIndexFragment ===

func TestRenderIndexFragmentFromIndex(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/members", "source=index")
	renderIndexFragment(rec, r, store, "Member added", "success")
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderIndexFragmentNotFromIndex(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/members", "")
	renderIndexFragment(rec, r, store, "Member added", "success")
	if recCode(rec) != 303 {
		t.Fatalf("expected 302 redirect, got %d", recCode(rec))
	}
}

// === RenderMemberDetail ===

func TestRenderMemberDetail(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	view := buildMemberDashboardView(store, 1)
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/member-detail?member_id=1", "")
	RenderMemberDetail(rec, r, view, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderMemberDetailHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	view := buildMemberDashboardView(store, 1)
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/member-detail?member_id=1", "")
	RenderMemberDetail(rec, r, view, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === RenderEventDetail ===

func TestRenderEventDetail(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	view := buildEventDashboardView(store, 1, "")
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/event-detail?event_id=1", "")
	RenderEventDetail(rec, r, view)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestRenderEventDetailHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	view := buildEventDashboardView(store, 1, "")
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/event-detail?event_id=1", "")
	RenderEventDetail(rec, r, view)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

// === Additional handler coverage tests ===

func TestHandleAttendanceBatchWithLateFine(t *testing.T) {
	store := NewStore()
	store.Settings.LateFineAmount = 500
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/attendance", "meeting_date=2026-07-01&present_1=on&late_1=on")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303, got %d", recCode(rec))
	}
	fineCount := 0
	for _, f := range store.Fines {
		if f.Reason == "Lateness" {
			fineCount++
		}
	}
	if fineCount != 1 {
		t.Fatalf("expected 1 lateness fine, got %d", fineCount)
	}
}

func TestHandleAttendanceBatchWithLateFineHTMX(t *testing.T) {
	store := NewStore()
	store.Settings.LateFineAmount = 500
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=2026-07-01&present_1=on&late_1=on")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect (not from index), got %d", recCode(rec))
	}
}

func TestHandleAttendanceBatchHTMXFromIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=2026-07-01&present_1=on&source=index")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200 from index, got %d", recCode(rec))
	}
}

func TestHandleAttendanceSingleMemberSuccessHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=2026-07-01&member_id=1&status=present")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect, got %d", recCode(rec))
	}
}

func TestHandleAttendanceSingleMemberSuccessHTMXFromIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/attendance", "meeting_date=2026-07-01&member_id=1&status=present&source=index")
	HandleAttendance(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleContributionHTMXSuccessNilView(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contribution", "event_id=999&member_id=1&amount=100")
	HandleContribution(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect, got %d", recCode(rec))
	}
}

func TestHandleContributionHTMXErrorNilView(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contribution", "event_id=1&member_id=1&amount=50")
	HandleContribution(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect (event doesn't exist, success path falls through), got %d", recCode(rec))
	}
}

func TestHandleEventsHTMXErrorNilView(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/events", "title=Test&min_amount_expected=abc")
	HandleEvents(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect, got %d", recCode(rec))
	}
}

func TestHandleAddFineHTMXSuccessFromIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/fines", "member_id=1&amount=500&reason=Test&fine_date=2026-07-01&source=index")
	HandleAddFine(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200 from index, got %d", recCode(rec))
	}
}

func TestHandleMembersHTMXFromIndex(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/members", "name=Bob&status=active&source=index")
	HandleMembers(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200 from index, got %d", recCode(rec))
	}
}

func TestHandleMembersHTMXErrorFromIndex(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/members", "name=&source=index")
	HandleMembers(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200 from index, got %d", recCode(rec))
	}
}

func TestHandleMembersHTMXCreateErrorFromIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/members", "name=&source=index")
	HandleMembers(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200 from index, got %d", recCode(rec))
	}
}

func TestHandleEventsHTMXGet(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("GET", "/events", "")
	HandleEvents(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 (renderIndexFragment redirect), got %d", recCode(rec))
	}
}

func TestHandleDeleteMemberHTMX(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/delete-member", "member_id=1")
	HandleDeleteMember(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleDeductFineSingleFineHTMXViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=999&fine_id=1")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400 (view nil, falls through to http.Error), got %d", recCode(rec))
	}
}

func TestHandleDeductFineSingleFineInvalidFineID(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/deduct-fine", "member_id=1&fine_id=abc")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidHTMXSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Fines = append(store.Fines, Fine{ID: 1, MemberID: 1, Amount: 500, Status: "outstanding"})
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-fine-paid", "member_id=1&fine_id=1")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidHTMXNotFoundViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-fine-paid", "member_id=999&fine_id=1")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400 (view nil, falls through to http.Error), got %d", recCode(rec))
	}
}

func TestHandleMarkFinePaidNotFoundViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-fine-paid", "member_id=999&fine_id=1")
	HandleMarkFinePaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidHTMXSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.Dues = append(store.Dues, DuesRecord{ID: 1, MemberID: 1, Amount: 2000, Status: "pending"})
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-dues-paid", "member_id=1&dues_id=1")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidHTMXNotFoundViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-dues-paid", "member_id=999&dues_id=1")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400 (view nil, falls through to http.Error), got %d", recCode(rec))
	}
}

func TestHandleMarkDuesPaidNotFoundViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-dues-paid", "member_id=999&dues_id=1")
	HandleMarkDuesPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidHTMXSuccess(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-contribution-paid", "member_id=1&event_id=1")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidHTMXNotFoundViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/mark-contribution-paid", "member_id=999&event_id=1")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400 (view nil, falls through to http.Error), got %d", recCode(rec))
	}
}

func TestHandleMarkContributionPaidNotFoundViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("POST", "/mark-contribution-paid", "member_id=999&event_id=1")
	HandleMarkContributionPaid(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400, got %d", recCode(rec))
	}
}

func TestHandleDeductFineAllFinesHTMXErrorViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=999")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 303 {
		t.Fatalf("expected 303 redirect, got %d", recCode(rec))
	}
}

func TestHandleDeductFineSingleFineHTMXSuccessViewNil(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/deduct-fine", "member_id=999&fine_id=1")
	HandleDeductFine(rec, r, store)
	if recCode(rec) != 400 {
		t.Fatalf("expected 400 (view nil, falls through to http.Error), got %d", recCode(rec))
	}
}

func TestHandleContributionHTMXSuccessWithView(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/contribution", "event_id=1&member_id=1&amount=200")
	HandleContribution(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleArrearsReportNoData(t *testing.T) {
	store := NewStore()
	rec := httptest.NewRecorder()
	r := newRequest("GET", "/arrears-report", "")
	HandleArrearsReport(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExportAttendancePDFWithDBStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.RecordAttendance(1, "2026-07-01", "present", "")

	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-attendance-pdf?from=2026-07-01&to=2026-07-31", "")
	HandleExportAttendancePDF(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExportContributionsPDFWithDBStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})

	rec := httptest.NewRecorder()
	r := newRequest("GET", "/export-contributions-pdf?event_id=1", "")
	HandleExportContributionsPDF(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleCommitteeReportWithDBStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pillars.db")
	store, _ := NewStoreWithPath(dbPath)
	store.CreateMember(Member{Name: "Alice", Status: "active", JoinedAt: time.Now().Format(time.RFC3339)})
	store.RecordAttendance(1, "2026-07-01", "present", "")
	store.AddDues(DuesRecord{MemberID: 1, Amount: 2000, Status: "paid", DueDate: "2026-07-01"})
	store.AddFine(Fine{MemberID: 1, Amount: 500, Status: "outstanding", Reason: "Test", FineDate: "2026-07-01"})

	rec := httptest.NewRecorder()
	r := newRequest("GET", "/committee-report", "")
	HandleCommitteeReport(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleSettleEventHTMXSuccessWithView(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "active"}}
	store.AddEvent(Event{Title: "Event", MinAmountExpected: 100, Status: "open"})
	store.Contributions[0].Status = "paid"

	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/settle-event", "event_id=1")
	HandleSettleEvent(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandlePromoteToActiveHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/promote-to-active", "member_id=1&source=index")
	HandlePromoteToActive(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}

func TestHandleExtendProbationHTMXSourceIndex(t *testing.T) {
	store := NewStore()
	store.Members = []Member{{ID: 1, Name: "Alice", Status: "probation"}}
	rec := httptest.NewRecorder()
	r := newHTMXRequest("POST", "/extend-probation", "member_id=1&months=3&source=index")
	HandleExtendProbation(rec, r, store)
	if recCode(rec) != 200 {
		t.Fatalf("expected 200, got %d", recCode(rec))
	}
}
