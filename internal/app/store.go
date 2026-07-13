package app

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

// Store is the cooperative domain store with optional SQLite persistence.
type Store struct {
	Members       []Member
	Attendance    []AttendanceRecord
	Dues          []DuesRecord
	Fines         []Fine
	Events        []Event
	Contributions []Contribution
	Applications  []OnboardingApplication
	Settings      Settings
	db            *sql.DB
}

func NewStore() *Store {
	return &Store{
		Settings: Settings{AbsenceFineAmount: 1000, LateFineAmount: 500, DuesAmount: 2000, ProbationPeriodDays: 90},
	}
}

func NewStoreWithPath(path string) (*Store, error) {
	if path == "" {
		path = "./data/pillars.db"
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db, Settings: Settings{AbsenceFineAmount: 1000, LateFineAmount: 500, DuesAmount: 2000, ProbationPeriodDays: 90}}
	if err := store.initDB(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.loadFromDB(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initDB() error {
	pragmas := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA busy_timeout=5000`,
		`PRAGMA foreign_keys=ON`,
	}
	for _, p := range pragmas {
		if _, err := s.db.Exec(p); err != nil {
			return err
		}
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			phone TEXT,
			status TEXT NOT NULL,
			joined_at TEXT,
			is_bonafide INTEGER,
			probation_ends TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS attendance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			meeting_date TEXT NOT NULL,
			status TEXT NOT NULL,
			note TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS dues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			deducted REAL NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			due_date TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS fines (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			status TEXT NOT NULL,
			reason TEXT,
			fine_date TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			date TEXT,
			goal_amount REAL NOT NULL,
			status TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS contributions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			status TEXT NOT NULL
		);`,
	}
	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := s.db.Exec(`ALTER TABLE fines ADD COLUMN fine_date TEXT`); err != nil && !contains(err.Error(), "duplicate column name") {
		return err
	}
	if _, err := s.db.Exec(`ALTER TABLE dues ADD COLUMN deducted REAL NOT NULL DEFAULT 0`); err != nil && !contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

func (s *Store) loadFromDB() error {
	members, err := s.listMembersFromDB()
	if err != nil {
		return err
	}
	s.Members = members
	attendance, err := s.listAttendanceFromDB()
	if err != nil {
		return err
	}
	s.Attendance = attendance
	dues, err := s.listDuesFromDB()
	if err != nil {
		return err
	}
	s.Dues = dues
	fines, err := s.listFinesFromDB()
	if err != nil {
		return err
	}
	s.Fines = fines
	events, err := s.listEventsFromDB()
	if err != nil {
		return err
	}
	s.Events = events
	contributions, err := s.listContributionsFromDB()
	if err != nil {
		return err
	}
	s.Contributions = contributions
	return nil
}

func (s *Store) MemberBalance(memberID int) float64 {
	balance := 0.0
	for _, due := range s.Dues {
		if due.MemberID == memberID && (due.Status == "paid" || due.Status == "partially_paid") {
			balance += due.Amount
		}
	}
	for _, fine := range s.Fines {
		if fine.MemberID == memberID && fine.Status == "outstanding" {
			balance -= fine.Amount
		}
	}
	return balance
}

func (s *Store) MemberDashboardSummary(memberID int, days int) MemberDashboardSummary {
	summary := MemberDashboardSummary{MemberID: memberID}
	for _, record := range s.Attendance {
		if record.MemberID != memberID {
			continue
		}
		if days > 0 {
			if record.MeetingDate == "" {
				continue
			}
		}
		summary.AttendanceTotal++
		if record.Status == "present" {
			summary.AttendancePresent++
		}
		if record.Status == "absent_without_permission" {
			summary.AttendanceAbsent++
		}
	}
	for _, due := range s.Dues {
		if due.MemberID != memberID {
			continue
		}
		if due.Status == "paid" || due.Status == "partially_paid" {
			summary.DuesPaid += due.Amount
		} else if due.Status == "pending" || due.Status == "owed" {
			summary.DuesOwed += due.Amount
		}
		summary.DuesOwed += due.Deducted
	}
	for _, fine := range s.Fines {
		if fine.MemberID != memberID {
			continue
		}
		if fine.Status == "outstanding" {
			summary.FinesOwed += fine.Amount
		} else {
			summary.FinesPaid += fine.Amount
		}
	}
	for _, contribution := range s.Contributions {
		if contribution.MemberID != memberID {
			continue
		}
		if contribution.Status == "paid" || contribution.Status == "partially_paid" || contribution.Status == "settled" {
			summary.ContributionsPaid += contribution.Amount
		} else {
			summary.ContributionsOwed += contribution.Amount
		}
	}
	return summary
}

func AgingBucket(oldestDebtDate string) string {
	if oldestDebtDate == "" {
		return "0-30"
	}
	parsed, err := time.Parse("2006-01-02", oldestDebtDate)
	if err != nil {
		return "0-30"
	}
	days := int(time.Since(parsed).Hours() / 24)
	if days <= 30 {
		return "0-30"
	} else if days <= 60 {
		return "31-60"
	} else if days <= 90 {
		return "61-90"
	}
	return "90+"
}

func FormatNaira(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}
	rounded := math.Round(amount*100) / 100
	intPart := int64(rounded)
	fracPart := int64(math.Round((rounded - float64(intPart)) * 100))

	s := fmt.Sprintf("%d", intPart)
	n := len(s)
	if n > 3 {
		var result []byte
		for i, c := range s {
			if i > 0 && (n-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		s = string(result)
	}
	formatted := fmt.Sprintf("₦%s.%02d", s, fracPart)
	if negative {
		formatted = "-" + formatted
	}
	return formatted
}

func (s *Store) MemberFinancialSummaries(from, to string) ([]MemberFinancialSummary, error) {
	type memberKey struct {
		id   int
		name string
	}
	memberMap := make(map[int]*MemberFinancialSummary)
	for _, m := range s.Members {
		if m.Status == "ex-member" {
			continue
		}
		memberMap[m.ID] = &MemberFinancialSummary{
			MemberID:   m.ID,
			MemberName: m.Name,
			Status:     m.Status,
		}
	}

	for _, due := range s.Dues {
		if due.DueDate < from || due.DueDate > to {
			continue
		}
		summary, ok := memberMap[due.MemberID]
		if !ok {
			summary = &MemberFinancialSummary{
				MemberID:   due.MemberID,
				MemberName: fmt.Sprintf("Member #%d", due.MemberID),
				Status:     "ex-member",
			}
			memberMap[due.MemberID] = summary
		}
		summary.DuesExpected += due.Amount + due.Deducted
		if due.Status == "paid" || due.Status == "partially_paid" {
			summary.DuesPaid += due.Amount
			summary.DuesDeducted += due.Deducted
		}
	}

	for _, fine := range s.Fines {
		if fine.FineDate < from || fine.FineDate > to {
			continue
		}
		summary, ok := memberMap[fine.MemberID]
		if !ok {
			summary = &MemberFinancialSummary{
				MemberID:   fine.MemberID,
				MemberName: fmt.Sprintf("Member #%d", fine.MemberID),
				Status:     "ex-member",
			}
			memberMap[fine.MemberID] = summary
		}
		summary.FinesLevied += fine.Amount
		if fine.Status == "paid" {
			summary.FinesPaid += fine.Amount
		}
	}

	eventDateMap := make(map[int]string)
	for _, ev := range s.Events {
		eventDateMap[ev.ID] = ev.Date
	}
	for _, contrib := range s.Contributions {
		eventDate, ok := eventDateMap[contrib.EventID]
		if !ok || eventDate < from || eventDate > to {
			continue
		}
		summary, ok2 := memberMap[contrib.MemberID]
		if !ok2 {
			summary = &MemberFinancialSummary{
				MemberID:   contrib.MemberID,
				MemberName: fmt.Sprintf("Member #%d", contrib.MemberID),
				Status:     "ex-member",
			}
			memberMap[contrib.MemberID] = summary
		}
		summary.ContributionsExpected += contrib.Amount
		if contrib.Status == "paid" || contrib.Status == "partially_paid" || contrib.Status == "settled" {
			summary.ContributionsPaid += contrib.Amount
		}
	}

	var result []memberKey
	for id := range memberMap {
		result = append(result, memberKey{id: id, name: memberMap[id].MemberName})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].name < result[j].name
	})
	summaries := make([]MemberFinancialSummary, 0, len(result))
	for _, key := range result {
		summary := memberMap[key.id]
		summary.NetBalance = (summary.DuesPaid + summary.FinesPaid + summary.ContributionsPaid) -
			(summary.DuesExpected + summary.FinesLevied + summary.ContributionsExpected)
		summaries = append(summaries, *summary)
	}
	return summaries, nil
}

func (s *Store) ArrearsByMember() ([]ArrearsRow, error) {
	type memberKey struct {
		id   int
		name string
	}
	memberMap := make(map[int]*ArrearsRow)
	for _, m := range s.Members {
		if m.Status == "ex-member" {
			continue
		}
		memberMap[m.ID] = &ArrearsRow{
			MemberID:   m.ID,
			MemberName: m.Name,
			Status:     m.Status,
		}
	}

	eventDateMap := make(map[int]string)
	for _, ev := range s.Events {
		eventDateMap[ev.ID] = ev.Date
	}

	for _, due := range s.Dues {
		if due.Status != "pending" && due.Status != "owed" {
			continue
		}
		summary, ok := memberMap[due.MemberID]
		if !ok {
			continue
		}
		summary.DuesOwed += due.Amount
		if summary.OldestDebt == "" || due.DueDate < summary.OldestDebt {
			summary.OldestDebt = due.DueDate
		}
	}

	for _, fine := range s.Fines {
		if fine.Status != "outstanding" {
			continue
		}
		summary, ok := memberMap[fine.MemberID]
		if !ok {
			continue
		}
		summary.FinesOwed += fine.Amount
		if summary.OldestDebt == "" || fine.FineDate < summary.OldestDebt {
			summary.OldestDebt = fine.FineDate
		}
	}

	for _, contrib := range s.Contributions {
		if contrib.Status != "pending" && contrib.Status != "not_paid" {
			continue
		}
		summary, ok := memberMap[contrib.MemberID]
		if !ok {
			continue
		}
		summary.ContribOwed += contrib.Amount
		eventDate := eventDateMap[contrib.EventID]
		if summary.OldestDebt == "" || eventDate < summary.OldestDebt {
			summary.OldestDebt = eventDate
		}
	}

	var result []memberKey
	for id := range memberMap {
		result = append(result, memberKey{id: id, name: memberMap[id].MemberName})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].name < result[j].name
	})

	var rows []ArrearsRow
	for _, key := range result {
		row := memberMap[key.id]
		row.TotalOwed = row.DuesOwed + row.FinesOwed + row.ContribOwed
		if row.TotalOwed <= 0 {
			continue
		}
		row.Bucket = AgingBucket(row.OldestDebt)
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].TotalOwed > rows[j].TotalOwed
	})
	return rows, nil
}

func (s *Store) RecordAttendance(memberID int, meetingDate, status, note string) error {
	if memberID <= 0 {
		return fmt.Errorf("member id is required")
	}
	if status == "" {
		return fmt.Errorf("attendance status is required")
	}

	record := AttendanceRecord{MemberID: memberID, MeetingDate: meetingDate, Status: status, Note: note}
	if s.db != nil {
		result, err := s.db.Exec(`INSERT INTO attendance(member_id, meeting_date, status, note) VALUES (?, ?, ?, ?)`, memberID, meetingDate, status, note)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		record.ID = int(id)
		s.Attendance = append(s.Attendance, record)
	} else {
		s.Attendance = append(s.Attendance, record)
	}

	if status == "absent_without_permission" {
		if s.db != nil {
			result, err := s.db.Exec(`INSERT INTO fines(member_id, amount, status, reason, fine_date) VALUES (?, ?, ?, ?, ?)`, memberID, s.Settings.AbsenceFineAmount, "outstanding", "Unapproved absence", meetingDate)
			if err != nil {
				return err
			}
			id, err := result.LastInsertId()
			if err != nil {
				return err
			}
			s.Fines = append(s.Fines, Fine{ID: int(id), MemberID: memberID, Amount: s.Settings.AbsenceFineAmount, Status: "outstanding", Reason: "Unapproved absence", FineDate: meetingDate})
		} else {
			s.Fines = append(s.Fines, Fine{MemberID: memberID, Amount: s.Settings.AbsenceFineAmount, Status: "outstanding", Reason: "Unapproved absence", FineDate: meetingDate})
		}
	}
	return nil
}

func recordAttendanceAndDues(store *Store, memberID int, meetingDate, status, note string, duesPaid bool, duesAmount float64) error {
	if err := store.RecordAttendance(memberID, meetingDate, status, note); err != nil {
		return err
	}
	if duesPaid {
		return store.AddDues(DuesRecord{MemberID: memberID, Amount: duesAmount, Status: "paid", DueDate: meetingDate})
	}
	return store.AddDues(DuesRecord{MemberID: memberID, Amount: duesAmount, Status: "pending", DueDate: meetingDate})
}

func attendanceStatusFromSelection(present, absenteeism bool) string {
	if present {
		return "present"
	}
	if absenteeism {
		return "absent_without_permission"
	}
	return "absent_with_permission"
}

func (s *Store) CreateMember(member Member) error {
	if member.Name == "" {
		return fmt.Errorf("member name is required")
	}
	if member.Status == "probation" && member.ProbationEnds == "" {
		joinedAt, err := time.Parse(time.RFC3339, member.JoinedAt)
		if err == nil {
			member.ProbationEnds = joinedAt.AddDate(0, 0, s.Settings.ProbationPeriodDays).Format("2006-01-02")
		}
	}
	if s.db != nil {
		result, err := s.db.Exec(`INSERT INTO members(name, email, phone, status, joined_at, is_bonafide, probation_ends) VALUES (?, ?, ?, ?, ?, ?, ?)`, member.Name, member.Email, member.Phone, member.Status, member.JoinedAt, boolToInt(member.IsBonafide), member.ProbationEnds)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		member.ID = int(id)
		s.Members = append(s.Members, member)
		return nil
	}
	member.ID = len(s.Members) + 1
	member.JoinedAt = member.JoinedAt
	s.Members = append(s.Members, member)
	return nil
}

func (s *Store) ProbationReviewDue() []Member {
	today := time.Now().Format("2006-01-02")
	var due []Member
	for _, member := range s.Members {
		if member.Status == "probation" && member.ProbationEnds != "" && member.ProbationEnds <= today {
			due = append(due, member)
		}
	}
	return due
}

func (s *Store) PromoteToActive(memberID int) error {
	for i := range s.Members {
		if s.Members[i].ID == memberID {
			if s.Members[i].Status != "probation" {
				return fmt.Errorf("member is not on probation")
			}
			if s.db != nil {
				if _, err := s.db.Exec(`UPDATE members SET status = 'active', is_bonafide = 1 WHERE id = ?`, memberID); err != nil {
					return err
				}
			}
			s.Members[i].Status = "active"
			s.Members[i].IsBonafide = true
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

func (s *Store) ExtendProbation(memberID int, months int) error {
	for i := range s.Members {
		if s.Members[i].ID == memberID {
			if s.Members[i].Status != "probation" {
				return fmt.Errorf("member is not on probation")
			}
			currentEnd := s.Members[i].ProbationEnds
			if currentEnd == "" {
				currentEnd = time.Now().Format("2006-01-02")
			}
			parsed, err := time.Parse("2006-01-02", currentEnd)
			if err != nil {
				return fmt.Errorf("invalid probation end date")
			}
			newEnd := parsed.AddDate(0, months, 0).Format("2006-01-02")
			if s.db != nil {
				if _, err := s.db.Exec(`UPDATE members SET probation_ends = ? WHERE id = ?`, newEnd, memberID); err != nil {
					return err
				}
			}
			s.Members[i].ProbationEnds = newEnd
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

func (s *Store) AddDues(d DuesRecord) error {
	if s.db != nil {
		result, err := s.db.Exec(`INSERT INTO dues(member_id, amount, status, due_date) VALUES (?, ?, ?, ?)`, d.MemberID, d.Amount, d.Status, d.DueDate)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		d.ID = int(id)
		s.Dues = append(s.Dues, d)
		return nil
	}
	d.ID = len(s.Dues) + 1
	s.Dues = append(s.Dues, d)
	return nil
}

func (s *Store) AddFine(f Fine) error {
	if s.db != nil {
		result, err := s.db.Exec(`INSERT INTO fines(member_id, amount, status, reason, fine_date) VALUES (?, ?, ?, ?, ?)`, f.MemberID, f.Amount, f.Status, f.Reason, f.FineDate)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		f.ID = int(id)
		s.Fines = append(s.Fines, f)
		return nil
	}
	f.ID = len(s.Fines) + 1
	s.Fines = append(s.Fines, f)
	return nil
}

func (s *Store) MarkFinePaid(memberID int, fineID int) error {
	for i := range s.Fines {
		if s.Fines[i].ID == fineID && s.Fines[i].MemberID == memberID && s.Fines[i].Status == "outstanding" {
			if s.db != nil {
				if _, err := s.db.Exec(`UPDATE fines SET status = 'paid' WHERE id = ?`, s.Fines[i].ID); err != nil {
					return err
				}
			}
			s.Fines[i].Status = "paid"
			return nil
		}
	}
	return fmt.Errorf("outstanding fine not found")
}

func (s *Store) MarkDuesPaid(memberID int, duesID int) error {
	for i := range s.Dues {
		if s.Dues[i].ID == duesID && s.Dues[i].MemberID == memberID && (s.Dues[i].Status == "pending" || s.Dues[i].Status == "owed" || s.Dues[i].Status == "partially_paid") {
			if s.db != nil {
				if _, err := s.db.Exec(`UPDATE dues SET status = 'paid' WHERE id = ?`, s.Dues[i].ID); err != nil {
					return err
				}
			}
			s.Dues[i].Status = "paid"
			return nil
		}
	}
	return fmt.Errorf("pending, owed, or partially_paid dues record not found")
}

func (s *Store) MarkContributionPaid(memberID int, eventID int) error {
	contributionAmount := 0.0
	contributionIdx := -1
	found := false
	for i := range s.Contributions {
		if s.Contributions[i].EventID == eventID && s.Contributions[i].MemberID == memberID && s.Contributions[i].Status == "pending" {
			contributionAmount = s.Contributions[i].Amount
			contributionIdx = i
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("pending contribution not found for this event")
	}

	var paidDuesIdx []int
	for i := range s.Dues {
		if s.Dues[i].MemberID == memberID && s.Dues[i].Status == "partially_paid" {
			paidDuesIdx = append(paidDuesIdx, i)
		}
	}
	for i := range s.Dues {
		if s.Dues[i].MemberID == memberID && s.Dues[i].Status == "paid" {
			paidDuesIdx = append(paidDuesIdx, i)
		}
	}

	if s.db != nil {
		remaining := contributionAmount
		for _, idx := range paidDuesIdx {
			if remaining <= 0 {
				break
			}
			if s.Dues[idx].Amount <= remaining {
				s.Dues[idx].Deducted += s.Dues[idx].Amount
				remaining -= s.Dues[idx].Amount
				s.Dues[idx].Amount = s.Settings.DuesAmount
				s.Dues[idx].Deducted = 0
				if _, err := s.db.Exec(`UPDATE dues SET amount = ?, status = 'pending', deducted = 0 WHERE id = ?`, s.Dues[idx].Amount, s.Dues[idx].ID); err != nil {
					return err
				}
				s.Dues[idx].Status = "pending"
			} else {
				s.Dues[idx].Deducted += remaining
				s.Dues[idx].Amount -= remaining
				if _, err := s.db.Exec(`UPDATE dues SET amount = ?, status = 'partially_paid', deducted = ? WHERE id = ?`, s.Dues[idx].Amount, s.Dues[idx].Deducted, s.Dues[idx].ID); err != nil {
					return err
				}
				s.Dues[idx].Status = "partially_paid"
				remaining = 0
			}
		}
		if remaining > 0 {
			if err := s.AddDues(DuesRecord{MemberID: memberID, Amount: remaining, Status: "owed"}); err != nil {
				return err
			}
		}
		if _, err := s.db.Exec(`UPDATE contributions SET status = 'paid' WHERE id = ?`, s.Contributions[contributionIdx].ID); err != nil {
			return err
		}
		s.Contributions[contributionIdx].Status = "paid"
		return nil
	}

	remaining := contributionAmount
	for _, idx := range paidDuesIdx {
		if remaining <= 0 {
			break
		}
		if s.Dues[idx].Amount <= remaining {
			s.Dues[idx].Deducted += s.Dues[idx].Amount
			remaining -= s.Dues[idx].Amount
			s.Dues[idx].Amount = s.Settings.DuesAmount
			s.Dues[idx].Deducted = 0
			s.Dues[idx].Status = "pending"
		} else {
			s.Dues[idx].Deducted += remaining
			s.Dues[idx].Amount -= remaining
			s.Dues[idx].Status = "partially_paid"
			remaining = 0
		}
	}
	if remaining > 0 {
		s.Dues = append(s.Dues, DuesRecord{ID: len(s.Dues) + 1, MemberID: memberID, Amount: remaining, Status: "owed"})
	}
	s.Contributions[contributionIdx].Status = "paid"
	return nil
}

func (s *Store) MarkEventSettled(eventID int) error {
	eventIdx := -1
	for i := range s.Events {
		if s.Events[i].ID == eventID {
			eventIdx = i
			break
		}
	}
	if eventIdx == -1 {
		return fmt.Errorf("event not found")
	}
	if s.Events[eventIdx].Status != "open" {
		return fmt.Errorf("event is not open")
	}

	for _, contrib := range s.Contributions {
		if contrib.EventID == eventID && contrib.Status == "pending" {
			return fmt.Errorf("not all members have contributed yet")
		}
	}

	if s.db != nil {
		if _, err := s.db.Exec(`UPDATE events SET status = 'settled' WHERE id = ?`, eventID); err != nil {
			return err
		}
	}
	s.Events[eventIdx].Status = "settled"

	for i := range s.Contributions {
		if s.Contributions[i].EventID == eventID && (s.Contributions[i].Status == "paid" || s.Contributions[i].Status == "partially_paid") {
			if s.db != nil {
				if _, err := s.db.Exec(`UPDATE contributions SET status = 'settled' WHERE id = ?`, s.Contributions[i].ID); err != nil {
					return err
				}
			}
			s.Contributions[i].Status = "settled"
		}
	}
	return nil
}

func (s *Store) DeductFineFromDues(memberID int, fineID int) error {
	var fineIdx int
	var fineAmount float64
	found := false
	for i := range s.Fines {
		if s.Fines[i].ID == fineID && s.Fines[i].MemberID == memberID && s.Fines[i].Status == "outstanding" {
			fineIdx = i
			fineAmount = s.Fines[i].Amount
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("outstanding fine not found")
	}

	var paidDuesIdx []int
	for i := range s.Dues {
		if s.Dues[i].MemberID == memberID && s.Dues[i].Status == "partially_paid" {
			paidDuesIdx = append(paidDuesIdx, i)
		}
	}
	for i := range s.Dues {
		if s.Dues[i].MemberID == memberID && s.Dues[i].Status == "paid" {
			paidDuesIdx = append(paidDuesIdx, i)
		}
	}

	if s.db != nil {
		remaining := fineAmount
		for _, idx := range paidDuesIdx {
			if remaining <= 0 {
				break
			}
			if s.Dues[idx].Amount <= remaining {
				s.Dues[idx].Deducted += s.Dues[idx].Amount
				remaining -= s.Dues[idx].Amount
				s.Dues[idx].Amount = s.Settings.DuesAmount
				s.Dues[idx].Deducted = 0
				if _, err := s.db.Exec(`UPDATE dues SET amount = ?, status = 'pending', deducted = 0 WHERE id = ?`, s.Dues[idx].Amount, s.Dues[idx].ID); err != nil {
					return err
				}
				s.Dues[idx].Status = "pending"
			} else {
				s.Dues[idx].Deducted += remaining
				s.Dues[idx].Amount -= remaining
				if _, err := s.db.Exec(`UPDATE dues SET amount = ?, status = 'partially_paid', deducted = ? WHERE id = ?`, s.Dues[idx].Amount, s.Dues[idx].Deducted, s.Dues[idx].ID); err != nil {
					return err
				}
				s.Dues[idx].Status = "partially_paid"
				remaining = 0
			}
		}
		if remaining > 0 {
			if err := s.AddDues(DuesRecord{MemberID: memberID, Amount: remaining, Status: "owed"}); err != nil {
				return err
			}
		}
		if _, err := s.db.Exec(`UPDATE fines SET status = 'paid' WHERE id = ?`, s.Fines[fineIdx].ID); err != nil {
			return err
		}
		s.Fines[fineIdx].Status = "paid"
		return nil
	}

	remaining := fineAmount
	for _, idx := range paidDuesIdx {
		if remaining <= 0 {
			break
		}
		if s.Dues[idx].Amount <= remaining {
			s.Dues[idx].Deducted += s.Dues[idx].Amount
			remaining -= s.Dues[idx].Amount
			s.Dues[idx].Amount = s.Settings.DuesAmount
			s.Dues[idx].Deducted = 0
			s.Dues[idx].Status = "pending"
		} else {
			s.Dues[idx].Deducted += remaining
			s.Dues[idx].Amount -= remaining
			s.Dues[idx].Status = "partially_paid"
			remaining = 0
		}
	}
	if remaining > 0 {
		s.Dues = append(s.Dues, DuesRecord{ID: len(s.Dues) + 1, MemberID: memberID, Amount: remaining, Status: "owed"})
	}
	s.Fines[fineIdx].Status = "paid"
	return nil
}

func (s *Store) DeductAllFinesFromDues(memberID int) error {
	totalFineAmount := 0.0
	var fineIdxs []int
	for i := range s.Fines {
		if s.Fines[i].MemberID == memberID && s.Fines[i].Status == "outstanding" {
			totalFineAmount += s.Fines[i].Amount
			fineIdxs = append(fineIdxs, i)
		}
	}
	if totalFineAmount == 0 {
		return nil
	}

	var paidDuesIdx []int
	for i := range s.Dues {
		if s.Dues[i].MemberID == memberID && s.Dues[i].Status == "partially_paid" {
			paidDuesIdx = append(paidDuesIdx, i)
		}
	}
	for i := range s.Dues {
		if s.Dues[i].MemberID == memberID && s.Dues[i].Status == "paid" {
			paidDuesIdx = append(paidDuesIdx, i)
		}
	}

	if s.db != nil {
		remaining := totalFineAmount
		for _, idx := range paidDuesIdx {
			if remaining <= 0 {
				break
			}
			if s.Dues[idx].Amount <= remaining {
				s.Dues[idx].Deducted += s.Dues[idx].Amount
				remaining -= s.Dues[idx].Amount
				s.Dues[idx].Amount = s.Settings.DuesAmount
				s.Dues[idx].Deducted = 0
				if _, err := s.db.Exec(`UPDATE dues SET amount = ?, status = 'pending', deducted = 0 WHERE id = ?`, s.Dues[idx].Amount, s.Dues[idx].ID); err != nil {
					return err
				}
				s.Dues[idx].Status = "pending"
			} else {
				s.Dues[idx].Deducted += remaining
				s.Dues[idx].Amount -= remaining
				if _, err := s.db.Exec(`UPDATE dues SET amount = ?, status = 'partially_paid', deducted = ? WHERE id = ?`, s.Dues[idx].Amount, s.Dues[idx].Deducted, s.Dues[idx].ID); err != nil {
					return err
				}
				s.Dues[idx].Status = "partially_paid"
				remaining = 0
			}
		}
		if remaining > 0 {
			if err := s.AddDues(DuesRecord{MemberID: memberID, Amount: remaining, Status: "owed"}); err != nil {
				return err
			}
		}
		for _, idx := range fineIdxs {
			if _, err := s.db.Exec(`UPDATE fines SET status = 'paid' WHERE id = ?`, s.Fines[idx].ID); err != nil {
				return err
			}
			s.Fines[idx].Status = "paid"
		}
		return nil
	}

	remaining := totalFineAmount
	for _, idx := range paidDuesIdx {
		if remaining <= 0 {
			break
		}
		if s.Dues[idx].Amount <= remaining {
			s.Dues[idx].Deducted += s.Dues[idx].Amount
			remaining -= s.Dues[idx].Amount
			s.Dues[idx].Amount = s.Settings.DuesAmount
			s.Dues[idx].Deducted = 0
			s.Dues[idx].Status = "pending"
		} else {
			s.Dues[idx].Deducted += remaining
			s.Dues[idx].Amount -= remaining
			s.Dues[idx].Status = "partially_paid"
			remaining = 0
		}
	}
	if remaining > 0 {
		s.Dues = append(s.Dues, DuesRecord{ID: len(s.Dues) + 1, MemberID: memberID, Amount: remaining, Status: "owed"})
	}
	for _, idx := range fineIdxs {
		s.Fines[idx].Status = "paid"
	}
	return nil
}

func (s *Store) AddEvent(ev Event) error {
	if s.db != nil {
		result, err := s.db.Exec(`INSERT INTO events(title, description, date, goal_amount, status) VALUES (?, ?, ?, ?, ?)`, ev.Title, ev.Description, ev.Date, ev.MinAmountExpected, ev.Status)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		ev.ID = int(id)
		s.Events = append(s.Events, ev)
		for _, member := range s.Members {
			if member.Status == "ex-member" {
				continue
			}
			contributionResult, err := s.db.Exec(`INSERT INTO contributions(event_id, member_id, amount, status) VALUES (?, ?, ?, ?)`, ev.ID, member.ID, ev.MinAmountExpected, "pending")
			if err != nil {
				return err
			}
			contributionID, err := contributionResult.LastInsertId()
			if err != nil {
				return err
			}
			s.Contributions = append(s.Contributions, Contribution{ID: int(contributionID), EventID: ev.ID, MemberID: member.ID, Amount: ev.MinAmountExpected, Status: "pending"})
		}
		return nil
	}
	ev.ID = len(s.Events) + 1
	s.Events = append(s.Events, ev)
	for _, member := range s.Members {
		if member.Status == "ex-member" {
			continue
		}
		s.Contributions = append(s.Contributions, Contribution{ID: len(s.Contributions) + 1, EventID: ev.ID, MemberID: member.ID, Amount: ev.MinAmountExpected, Status: "pending"})
	}
	return nil
}

func (s *Store) DeleteMember(memberID int) error {
	if s.db != nil {
		_, err := s.db.Exec(`UPDATE members SET status = 'ex-member' WHERE id = ?`, memberID)
		if err != nil {
			return err
		}
	}
	for i, member := range s.Members {
		if member.ID == memberID {
			s.Members[i].Status = "ex-member"
			break
		}
	}
	return nil
}

func (s *Store) AddContribution(c Contribution) error {
	minimumAmount := 0.0
	for _, event := range s.Events {
		if event.ID == c.EventID {
			minimumAmount = event.MinAmountExpected
			break
		}
	}
	if minimumAmount == 0 && s.db != nil {
		var event Event
		err := s.db.QueryRow(`SELECT id, title, description, date, goal_amount, status FROM events WHERE id = ?`, c.EventID).Scan(&event.ID, &event.Title, &event.Description, &event.Date, &event.MinAmountExpected, &event.Status)
		if err == nil {
			minimumAmount = event.MinAmountExpected
		}
	}

	if s.db != nil {
		var existingID int
		var existingAmount float64
		var existingStatus string
		err := s.db.QueryRow(`SELECT id, amount, status FROM contributions WHERE event_id = ? AND member_id = ?`, c.EventID, c.MemberID).Scan(&existingID, &existingAmount, &existingStatus)
		if err == nil {
			var newAmount float64
			if existingStatus == "paid" || existingStatus == "partially_paid" {
				newAmount = existingAmount + c.Amount
			} else {
				if c.Amount < minimumAmount {
					return fmt.Errorf("amount must be at least %.2f", minimumAmount)
				}
				newAmount = c.Amount
			}
			if _, err := s.db.Exec(`UPDATE contributions SET amount = ?, status = ? WHERE id = ?`, newAmount, c.Status, existingID); err != nil {
				return err
			}
			for i := range s.Contributions {
				if s.Contributions[i].EventID == c.EventID && s.Contributions[i].MemberID == c.MemberID {
					s.Contributions[i].Amount = newAmount
					s.Contributions[i].Status = c.Status
					break
				}
			}
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}

		if minimumAmount > 0 && c.Amount < minimumAmount {
			return fmt.Errorf("amount must be at least %.2f", minimumAmount)
		}
		result, err := s.db.Exec(`INSERT INTO contributions(event_id, member_id, amount, status) VALUES (?, ?, ?, ?)`, c.EventID, c.MemberID, c.Amount, c.Status)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		c.ID = int(id)
		s.Contributions = append(s.Contributions, c)
		return nil
	}

	for i := range s.Contributions {
		if s.Contributions[i].EventID == c.EventID && s.Contributions[i].MemberID == c.MemberID {
			if s.Contributions[i].Status == "paid" || s.Contributions[i].Status == "partially_paid" {
				s.Contributions[i].Amount += c.Amount
			} else {
				if c.Amount < minimumAmount {
					return fmt.Errorf("amount must be at least %.2f", minimumAmount)
				}
				s.Contributions[i].Amount = c.Amount
			}
			s.Contributions[i].Status = c.Status
			return nil
		}
	}
	if minimumAmount > 0 && c.Amount < minimumAmount {
		return fmt.Errorf("amount must be at least %.2f", minimumAmount)
	}
	c.ID = len(s.Contributions) + 1
	s.Contributions = append(s.Contributions, c)
	return nil
}

func (s *Store) listMembersFromDB() ([]Member, error) {
	rows, err := s.db.Query(`SELECT id, name, email, phone, status, joined_at, is_bonafide, probation_ends FROM members ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var member Member
		var bonafide int
		if err := rows.Scan(&member.ID, &member.Name, &member.Email, &member.Phone, &member.Status, &member.JoinedAt, &bonafide, &member.ProbationEnds); err != nil {
			return nil, err
		}
		member.IsBonafide = bonafide == 1
		members = append(members, member)
	}
	return members, rows.Err()
}

func (s *Store) listAttendanceFromDB() ([]AttendanceRecord, error) {
	rows, err := s.db.Query(`SELECT id, member_id, meeting_date, status, note FROM attendance ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AttendanceRecord
	for rows.Next() {
		var record AttendanceRecord
		if err := rows.Scan(&record.ID, &record.MemberID, &record.MeetingDate, &record.Status, &record.Note); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *Store) TotalTreasuryBalance() float64 {
	balance := 0.0
	for _, due := range s.Dues {
		if due.Status == "paid" || due.Status == "partially_paid" {
			balance += due.Amount
		}
	}
	for _, fine := range s.Fines {
		if fine.Status == "paid" {
			balance += fine.Amount
		}
	}
	for _, contrib := range s.Contributions {
		if contrib.Status == "paid" || contrib.Status == "partially_paid" {
			balance += contrib.Amount
		}
	}
	return balance
}

func (s *Store) TotalOutstandingReceivables() float64 {
	owed := 0.0
	for _, due := range s.Dues {
		if due.Status == "pending" || due.Status == "owed" {
			owed += due.Amount
		}
	}
	for _, fine := range s.Fines {
		if fine.Status == "outstanding" {
			owed += fine.Amount
		}
	}
	for _, contrib := range s.Contributions {
		if contrib.Status == "pending" || contrib.Status == "not_paid" {
			owed += contrib.Amount
		}
	}
	return owed
}

func (s *Store) AtRiskMembersCount() int {
	count := 0
	for _, member := range s.Members {
		duesOwed := 0.0
		for _, due := range s.Dues {
			if due.MemberID == member.ID && (due.Status == "pending" || due.Status == "owed") {
				duesOwed += due.Amount
			}
		}

		has3MonthsDues := duesOwed > 3*s.Settings.DuesAmount

		var fineDates []time.Time
		for _, fine := range s.Fines {
			if fine.MemberID == member.ID && fine.Status == "outstanding" && fine.Reason == "Unapproved absence" {
				t, err := time.Parse("2006-01-02", fine.FineDate)
				if err == nil {
					fineDates = append(fineDates, t)
				}
			}
		}

		has3ConsecutiveMonths := false
		if len(fineDates) >= 3 {
			sort.Slice(fineDates, func(i, j int) bool {
				return fineDates[i].Before(fineDates[j])
			})
			var months []time.Time
			for _, d := range fineDates {
				m := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
				if len(months) == 0 || !months[len(months)-1].Equal(m) {
					months = append(months, m)
				}
			}
			consecutive := 1
			for i := 1; i < len(months); i++ {
				expectedNext := months[i-1].AddDate(0, 1, 0)
				if months[i].Equal(expectedNext) {
					consecutive++
					if consecutive >= 3 {
						has3ConsecutiveMonths = true
						break
					}
				} else {
					consecutive = 1
				}
			}
		}

		if has3MonthsDues && has3ConsecutiveMonths {
			count++
		}
	}
	return count
}

func (s *Store) listDuesFromDB() ([]DuesRecord, error) {
	rows, err := s.db.Query(`SELECT id, member_id, amount, deducted, status, due_date FROM dues ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dues []DuesRecord
	for rows.Next() {
		var record DuesRecord
		if err := rows.Scan(&record.ID, &record.MemberID, &record.Amount, &record.Deducted, &record.Status, &record.DueDate); err != nil {
			return nil, err
		}
		dues = append(dues, record)
	}
	return dues, rows.Err()
}

func (s *Store) listFinesFromDB() ([]Fine, error) {
	rows, err := s.db.Query(`SELECT id, member_id, amount, status, reason, fine_date FROM fines ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fines []Fine
	for rows.Next() {
		var record Fine
		var fineDate sql.NullString
		if err := rows.Scan(&record.ID, &record.MemberID, &record.Amount, &record.Status, &record.Reason, &fineDate); err != nil {
			return nil, err
		}
		if fineDate.Valid {
			record.FineDate = fineDate.String
		}
		fines = append(fines, record)
	}
	return fines, rows.Err()
}

func (s *Store) listEventsFromDB() ([]Event, error) {
	rows, err := s.db.Query(`SELECT id, title, description, date, goal_amount, status FROM events ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var ev Event
		if err := rows.Scan(&ev.ID, &ev.Title, &ev.Description, &ev.Date, &ev.MinAmountExpected, &ev.Status); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (s *Store) listContributionsFromDB() ([]Contribution, error) {
	rows, err := s.db.Query(`SELECT id, event_id, member_id, amount, status FROM contributions ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributions []Contribution
	for rows.Next() {
		var c Contribution
		if err := rows.Scan(&c.ID, &c.EventID, &c.MemberID, &c.Amount, &c.Status); err != nil {
			return nil, err
		}
		contributions = append(contributions, c)
	}
	return contributions, rows.Err()
}

func contains(text, needle string) bool {
	return len(text) >= len(needle) && (text == needle || len(text) > len(needle) && (contains(text[1:], needle) || text[:len(needle)] == needle))
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

type Member struct {
	ID            int
	Name          string
	Email         string
	Phone         string
	Status        string
	JoinedAt      string
	IsBonafide    bool
	ProbationEnds string
}

type AttendanceRecord struct {
	ID          int
	MemberID    int
	MeetingDate string
	Status      string
	Note        string
}

type DuesRecord struct {
	ID        int
	MemberID  int
	Amount    float64
	Deducted  float64
	Status    string
	DueDate   string
}

type Fine struct {
	ID       int
	MemberID int
	Amount   float64
	Status   string
	Reason   string
	FineDate string
}

type Event struct {
	ID                int
	Title             string
	Description       string
	Date              string
	MinAmountExpected float64
	Status            string
}

func (s *Store) EventTitle(eventID int) string {
	for _, ev := range s.Events {
		if ev.ID == eventID {
			return ev.Title
		}
	}
	return fmt.Sprintf("Event #%d", eventID)
}

type Contribution struct {
	ID       int
	EventID  int
	MemberID int
	Amount   float64
	Status   string
}

type OnboardingApplication struct {
	ID         int
	MemberID   int
	Status     string
	AppliedAt  string
	ReviewedAt string
}

type Settings struct {
	AbsenceFineAmount   float64
	LateFineAmount      float64
	DuesAmount          float64
	ProbationPeriodDays int
}

type MemberDashboardSummary struct {
	MemberID          int
	AttendanceTotal   int
	AttendancePresent int
	AttendanceAbsent  int
	DuesPaid          float64
	DuesOwed          float64
	FinesPaid         float64
	FinesOwed         float64
	ContributionsPaid float64
	ContributionsOwed float64
	AbsenceFineAmount float64
}

type AttendanceGroup struct {
	MeetingDate string
	Count       int
	Status      string
}

type MemberFinancialSummary struct {
	MemberID              int
	MemberName            string
	Status                string
	DuesExpected          float64
	DuesPaid              float64
	DuesDeducted          float64
	FinesLevied           float64
	FinesPaid             float64
	ContributionsExpected float64
	ContributionsPaid     float64
	NetBalance            float64
}

type ArrearsRow struct {
	MemberID    int
	MemberName  string
	Status      string
	DuesOwed    float64
	FinesOwed   float64
	ContribOwed float64
	TotalOwed   float64
	OldestDebt  string
	Bucket      string
}

func groupAttendanceByDate(records []AttendanceRecord) []AttendanceGroup {
	groups := make([]AttendanceGroup, 0)
	seen := make(map[string]int)
	for _, record := range records {
		if record.MeetingDate == "" {
			continue
		}
		if index, ok := seen[record.MeetingDate]; ok {
			groups[index].Count++
			continue
		}
		seen[record.MeetingDate] = len(groups)
		groups = append(groups, AttendanceGroup{MeetingDate: record.MeetingDate, Count: 1, Status: record.Status})
	}
	return groups
}
