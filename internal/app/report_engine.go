package app

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

//go:embed NotoSans-Regular.ttf
var notoSansFont []byte

type ReportBuilder struct {
	pdf      *gofpdf.Fpdf
	config   ReportConfig
	sections []ReportSection
	summaries [][]SummaryRow
}

type ReportConfig struct {
	Title      string
	Subtitle   string
	CoopName   string
	Orientation string // "P" for portrait, "L" for landscape
}

type TableColumn struct {
	Header string
	Width  float64
	Align  string
}

type ReportSection struct {
	Title    string
	Columns  []TableColumn
	Rows     [][]string
	Summary  []SummaryRow
	Widths   []float64
}

type SummaryRow struct {
	Label string
	Value string
	Bold  bool
}

func NewReportBuilder(config ReportConfig) *ReportBuilder {
	orientation := config.Orientation
	if orientation == "" {
		orientation = "P"
	}
	pdf := gofpdf.New(orientation, "mm", "A4", "")

	pdf.AddUTF8FontFromBytes("notosans", "", notoSansFont)
	pdf.AddUTF8FontFromBytes("notosans", "B", notoSansFont)
	pdf.AddUTF8FontFromBytes("notosans", "I", notoSansFont)
	pdf.AddUTF8FontFromBytes("notosans", "BI", notoSansFont)
	pdf.SetFont("notosans", "", 10)

	return &ReportBuilder{
		pdf:    pdf,
		config: config,
	}
}

func (rb *ReportBuilder) AddSection(section ReportSection) {
	rb.sections = append(rb.sections, section)
}

func (rb *ReportBuilder) AddSummary(rows []SummaryRow) {
	rb.summaries = append(rb.summaries, rows)
}

func (rb *ReportBuilder) Render() ([]byte, error) {
	rb.pdf.AddPage()

	rb.pdf.SetFont("notosans", "B", 16)
	rb.pdf.Cell(0, 10, rb.config.CoopName)
	rb.pdf.Ln(8)

	rb.pdf.SetFont("notosans", "B", 14)
	rb.pdf.Cell(0, 8, rb.config.Title)
	rb.pdf.Ln(8)

	if rb.config.Subtitle != "" {
		rb.pdf.SetFont("notosans", "", 11)
		rb.pdf.Cell(0, 6, rb.config.Subtitle)
		rb.pdf.Ln(8)
	}

	rb.pdf.Ln(4)

	for _, section := range rb.sections {
		rb.renderSection(section)
		rb.pdf.Ln(6)
	}

	for _, summary := range rb.summaries {
		rb.renderSummary(summary)
		rb.pdf.Ln(4)
	}

	rb.pdf.SetFont("notosans", "I", 8)
	rb.pdf.Cell(0, 5, fmt.Sprintf("Generated: %s", rb.config.Subtitle))
	rb.pdf.Ln(5)

	var buf bytes.Buffer
	if err := rb.pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (rb *ReportBuilder) renderSection(section ReportSection) {
	if section.Title != "" {
		rb.pdf.SetFont("notosans", "B", 12)
		rb.pdf.Cell(0, 8, section.Title)
		rb.pdf.Ln(8)
	}

	if len(section.Rows) == 0 {
		rb.pdf.SetFont("notosans", "I", 10)
		rb.pdf.Cell(0, 6, "No data available.")
		rb.pdf.Ln(8)
		return
	}

	if len(section.Columns) > 0 {
		rb.pdf.SetFont("notosans", "B", 8)
		for _, col := range section.Columns {
			align := col.Align
			if align == "" {
				align = "L"
			}
			rb.pdf.CellFormat(col.Width, 7, col.Header, "1", 0, align, false, 0, "")
		}
		rb.pdf.Ln(-1)

		rb.pdf.SetFont("notosans", "", 8)
		for _, row := range section.Rows {
			if rb.pdf.GetY()+7 > 270 {
				rb.pdf.AddPage()
				rb.pdf.SetFont("notosans", "B", 8)
				for _, col := range section.Columns {
					align := col.Align
					if align == "" {
						align = "L"
					}
					rb.pdf.CellFormat(col.Width, 7, col.Header, "1", 0, align, false, 0, "")
				}
				rb.pdf.Ln(-1)
				rb.pdf.SetFont("notosans", "", 8)
			}
			for i, cell := range row {
				align := "L"
				if i < len(section.Columns) {
					align = section.Columns[i].Align
					if align == "" {
						align = "L"
					}
				}
				rb.pdf.CellFormat(section.Columns[i].Width, 7, cell, "1", 0, align, false, 0, "")
			}
			rb.pdf.Ln(-1)
		}
	}

	if len(section.Summary) > 0 {
		rb.pdf.Ln(2)
		for _, row := range section.Summary {
			if row.Bold {
				rb.pdf.SetFont("notosans", "B", 9)
			} else {
				rb.pdf.SetFont("notosans", "", 9)
			}
			rb.pdf.Cell(80, 6, row.Label)
			rb.pdf.Cell(0, 6, row.Value)
			rb.pdf.Ln(6)
		}
	}
}

func (rb *ReportBuilder) renderSummary(rows []SummaryRow) {
	rb.pdf.SetFont("notosans", "B", 11)
	rb.pdf.Cell(0, 8, "Summary")
	rb.pdf.Ln(8)

	for _, row := range rows {
		if row.Bold {
			rb.pdf.SetFont("notosans", "B", 10)
		} else {
			rb.pdf.SetFont("notosans", "", 10)
		}
		rb.pdf.Cell(100, 6, row.Label)
		rb.pdf.Cell(0, 6, row.Value)
		rb.pdf.Ln(6)
	}
}
