package exportfile

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestExportNeutralizesFormulasAndPreservesNumbers(t *testing.T) {
	columns := []Column{{Title: "Tên"}, {Title: "Thu ròng", Number: true}}
	rows := [][]string{{" \t=1+1", "-101"}, {"\u200b@SUM(A1)", "1800000000000001"}, {"+84901234567", "10"}}
	file, err := Table("csv", columns, rows)
	if err != nil {
		t.Fatal(err)
	}
	csvRows, err := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(file.Bytes, []byte{0xef, 0xbb, 0xbf}))).ReadAll()
	if err != nil || csvRows[1][0] != "' \t=1+1" || csvRows[1][1] != "-101" || csvRows[2][0] != "'\u200b@SUM(A1)" || csvRows[3][0] != "'+84901234567" {
		t.Fatalf("CSV cells %v: %v", csvRows, err)
	}
	file, err = Table("xlsx", columns, rows)
	if err != nil {
		t.Fatal(err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(file.Bytes))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	sheet := book.GetSheetName(0)
	if dimension, err := book.GetSheetDimension(sheet); err != nil || dimension != "A1:B4" {
		t.Fatalf("sheet dimension %s: %v", dimension, err)
	}
	for cell, value := range map[string]string{"A2": rows[0][0], "A3": rows[1][0], "A4": rows[2][0], "B2": "-101", "B3": rows[1][1]} {
		got, e := book.GetCellValue(sheet, cell)
		formula, ferr := book.GetCellFormula(sheet, cell)
		if e != nil || ferr != nil || got != value || formula != "" {
			t.Fatalf("%s value=%q formula=%q: %v / %v", cell, got, formula, e, ferr)
		}
	}
	if _, err := Table("xlsx", columns, make([][]string, MaxRows+1)); err != ErrTooLarge {
		t.Fatal("large report accepted", err)
	}
	if dir := os.Getenv("BLUESKY_EXPORT_QA"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "so-thu-demo.xlsx"), file.Bytes, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPDFVietnameseAndPageBreak(t *testing.T) {
	body := HTMLText(`<p>Đón khách ở Thành phố Hồ Chí Minh &amp; đi Ninh Chữ.</p><img src="http://127.0.0.1:65535/private"><p>Làng gốm Bàu Trúc</p>`)
	if strings.Contains(body, "127.0.0.1") || !strings.Contains(body, "Minh & đi Ninh Chữ") {
		t.Fatal(body)
	}
	file, err := PDF("Ninh Chữ · Biển và văn hoá Chăm", "Báo giá đoàn · Phiên bản 2\nTổng tiền: 1.650.000 VND", []Section{
		{Title: "Ngày 1 · Làng gốm Bàu Trúc", Body: strings.Repeat(body+"\n", 35)},
		{Title: "Ngày 2 · Hang Rái", Body: "Vườn nho và đồng muối Phương Cựu."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(file.Bytes, []byte("%PDF-")) || bytes.Count(file.Bytes, []byte("/Type /Page\n")) < 2 {
		t.Fatal("expected a multisection PDF with page breaks")
	}
	if dir := os.Getenv("BLUESKY_EXPORT_QA"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "lich-trinh-demo.pdf"), file.Bytes, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
