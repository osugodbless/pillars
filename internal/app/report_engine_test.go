package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewReportBuilderPortrait(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", Subtitle: "Sub", CoopName: "Coop"})
	if rb == nil {
		t.Fatal("expected non-nil ReportBuilder")
	}
	if rb.config.Title != "Test" {
		t.Fatalf("expected title 'Test', got %q", rb.config.Title)
	}
	if rb.pdf == nil {
		t.Fatal("expected non-nil pdf")
	}
}

func TestNewReportBuilderLandscape(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Orientation: "L"})
	if rb.config.Orientation != "L" {
		t.Fatalf("expected orientation 'L', got %q", rb.config.Orientation)
	}
}

func TestReportBuilderAddSection(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test"})
	rb.AddSection(ReportSection{Title: "Section 1"})
	if len(rb.sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(rb.sections))
	}
}

func TestReportBuilderAddSummary(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test"})
	rb.AddSummary([]SummaryRow{{Label: "Total", Value: "100"}})
	if len(rb.summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(rb.summaries))
	}
}

func TestRenderEmpty(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Empty", CoopName: "Coop"})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatal("expected PDF header")
	}
}

func TestRenderWithSubtitle(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Title", Subtitle: "Subtitle", CoopName: "Coop"})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderWithoutSubtitle(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Title", CoopName: "Coop"})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderSectionWithTitleAndRows(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", CoopName: "Coop"})
	rb.AddSection(ReportSection{
		Title:   "Financial Summary",
		Columns: []TableColumn{{Header: "Name", Width: 50, Align: "L"}, {Header: "Amount", Width: 30, Align: "R"}},
		Rows:    [][]string{{"Alice", "1000"}, {"Bob", "2000"}},
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderSectionEmptyRows(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", CoopName: "Coop"})
	rb.AddSection(ReportSection{
		Title: "Empty Section",
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderSectionNoTitle(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", CoopName: "Coop"})
	rb.AddSection(ReportSection{
		Rows: [][]string{{"data"}},
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderSectionColumnDefaultAlign(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", CoopName: "Coop"})
	rb.AddSection(ReportSection{
		Columns: []TableColumn{{Header: "Col", Width: 50}},
		Rows:    [][]string{{"val"}},
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderSectionWithSummary(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", CoopName: "Coop"})
	rb.AddSection(ReportSection{
		Title: "With Summary",
		Columns: []TableColumn{{Header: "Name", Width: 50}},
		Rows:    [][]string{{"Alice"}},
		Summary: []SummaryRow{
			{Label: "Total", Value: "1000", Bold: true},
			{Label: "Subtotal", Value: "500", Bold: false},
		},
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderSectionNoColumnsNoRows(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", CoopName: "Coop"})
	rb.AddSection(ReportSection{Title: "Bare Section"})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderSummary(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Test", CoopName: "Coop"})
	rb.AddSummary([]SummaryRow{
		{Label: "Treasury", Value: "₦50,000", Bold: true},
		{Label: "Owed", Value: "₦10,000", Bold: false},
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderMultipleSectionsAndSummaries(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Full Report", CoopName: "Pillars"})
	rb.AddSection(ReportSection{
		Title:   "Section 1",
		Columns: []TableColumn{{Header: "A", Width: 50}},
		Rows:    [][]string{{"row1"}},
	})
	rb.AddSection(ReportSection{
		Title: "Section 2",
	})
	rb.AddSummary([]SummaryRow{{Label: "Sum1", Value: "v1"}})
	rb.AddSummary([]SummaryRow{{Label: "Sum2", Value: "v2"}})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderManyRows(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Many Rows", CoopName: "Coop"})
	var rows [][]string
	for i := 0; i < 100; i++ {
		rows = append(rows, []string{"Name", "Amount"})
	}
	rb.AddSection(ReportSection{
		Columns: []TableColumn{{Header: "Name", Width: 50}, {Header: "Amount", Width: 30}},
		Rows:    rows,
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderPageBreak(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Page Break", CoopName: "Coop"})
	var rows [][]string
	for i := 0; i < 50; i++ {
		rows = append(rows, []string{"Long Name Here", "Amount Value"})
	}
	rb.AddSection(ReportSection{
		Columns: []TableColumn{
			{Header: "Member Name", Width: 80},
			{Header: "Status", Width: 40},
			{Header: "Dues Owed", Width: 40},
			{Header: "Fines", Width: 40},
			{Header: "Contrib", Width: 40},
			{Header: "Total", Width: 40},
			{Header: "Date", Width: 40},
			{Header: "Age", Width: 40},
		},
		Rows: rows,
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderLandscape(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Landscape", CoopName: "Coop", Orientation: "L"})
	rb.AddSection(ReportSection{
		Title:   "Landscape Section",
		Columns: []TableColumn{{Header: "Col", Width: 50}},
		Rows:    [][]string{{"data"}},
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}

func TestRenderGeneratedTimestamp(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Timestamped", Subtitle: "Period: 2026-01-01 to 2026-12-31", CoopName: "Coop"})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = pdf
	_ = strings.Contains
}

func TestReportSectionTypes(t *testing.T) {
	rb := NewReportBuilder(ReportConfig{Title: "Types Test", CoopName: "Coop"})
	rb.AddSection(ReportSection{
		Title: "Mixed Content",
		Columns: []TableColumn{
			{Header: "Left", Width: 50, Align: "L"},
			{Header: "Center", Width: 30, Align: "C"},
			{Header: "Right", Width: 30, Align: "R"},
		},
		Rows: [][]string{
			{"A", "B", "C"},
			{"D", "E", "F"},
		},
		Summary: []SummaryRow{
			{Label: "Bold Total", Value: "100", Bold: true},
			{Label: "Normal Total", Value: "200", Bold: false},
		},
	})
	pdf, err := rb.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("expected non-empty PDF")
	}
}
