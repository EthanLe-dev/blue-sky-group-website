package exportfile

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"golang.org/x/net/html"
)

//go:embed fonts/BeVietnamPro-Regular.ttf
var vietnameseFont []byte

type Section struct{ Title, Body string }

func PDF(title, metadata string, sections []Section) (File, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("Vietnam", "", vietnameseFont)
	pdf.SetTitle(title, true)
	pdf.SetAuthor("Blue Sky Group", true)
	pdf.SetMargins(19, 23, 19)
	pdf.SetAutoPageBreak(true, 21)
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Vietnam", "", 10)
		pdf.SetTextColor(27, 69, 118)
		pdf.SetXY(19, 9)
		pdf.Cell(0, 7, "BLUE SKY GROUP")
		pdf.SetDrawColor(29, 112, 182)
		pdf.Line(19, 18, 191, 18)
		pdf.SetXY(19, 23)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Vietnam", "", 9)
		pdf.SetTextColor(68, 82, 103)
		pdf.CellFormat(0, 6, fmt.Sprintf("Trang %d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AddPage()
	pdf.SetTextColor(16, 39, 69)
	pdf.SetFont("Vietnam", "", 22)
	pdf.MultiCell(0, 10, title, "", "L", false)
	pdf.SetFont("Vietnam", "", 11)
	pdf.MultiCell(0, 6, metadata, "", "L", false)
	pdf.Ln(3)
	for _, section := range sections {
		if strings.TrimSpace(section.Body) == "" {
			continue
		}
		if pdf.GetY() > 257 {
			pdf.AddPage()
		}
		pdf.Ln(4)
		pdf.SetFont("Vietnam", "", 13)
		pdf.SetTextColor(27, 69, 118)
		pdf.MultiCell(0, 7, section.Title, "", "L", false)
		pdf.SetTextColor(24, 38, 54)
		pdf.SetFont("Vietnam", "", 11)
		pdf.MultiCell(0, 6, section.Body, "", "L", false)
	}
	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return File{}, err
	}
	return File{out.Bytes(), "application/pdf"}, nil
}

// HTMLText extracts visible text. The renderer never interprets markup or fetches URLs.
func HTMLText(input string) string {
	z := html.NewTokenizer(strings.NewReader(input))
	var out strings.Builder
	for {
		switch z.Next() {
		case html.ErrorToken:
			return strings.TrimSpace(out.String())
		case html.TextToken:
			out.Write(z.Text())
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "p", "div", "br", "li", "ul", "ol", "h2", "h3", "h4", "tr":
				out.WriteByte('\n')
			}
		}
	}
}
