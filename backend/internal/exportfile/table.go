package exportfile

import (
	"bytes"
	"encoding/csv"
	"errors"
	"mime"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

const MaxRows = 10000

var ErrTooLarge = errors.New("export exceeds row limit")

type Column struct {
	Title  string
	Number bool
}

type File struct {
	Bytes       []byte
	ContentType string
}

func (f File) Download(c *gin.Context, filename string) {
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	c.Header("Cache-Control", "no-store")
	c.Data(200, f.ContentType, f.Bytes)
}

func safeCell(value string) string {
	trimmed := strings.TrimLeftFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
	})
	if strings.HasPrefix(value, "\t") || strings.HasPrefix(value, "\r") || strings.HasPrefix(value, "\n") ||
		(trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0]))) {
		return "'" + value
	}
	return value
}

func Table(format string, columns []Column, rows [][]string) (File, error) {
	if len(rows) > MaxRows {
		return File{}, ErrTooLarge
	}
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = col.Title
	}
	if format == "csv" {
		var out bytes.Buffer
		out.WriteString("\xef\xbb\xbf")
		writer := csv.NewWriter(&out)
		writer.UseCRLF = true
		if err := writer.Write(names); err != nil {
			return File{}, err
		}
		for _, row := range rows {
			values := make([]string, len(row))
			for i, value := range row {
				values[i] = safeCell(value)
				if columns[i].Number {
					if _, err := strconv.ParseInt(value, 10, 64); err == nil {
						values[i] = value
					}
				}
			}
			if err := writer.Write(values); err != nil {
				return File{}, err
			}
		}
		writer.Flush()
		return File{out.Bytes(), "text/csv; charset=utf-8"}, writer.Error()
	}

	book := excelize.NewFile()
	defer book.Close()
	sheet := book.GetSheetName(0)
	head := make([]any, len(names))
	for i, name := range names {
		head[i] = name
	}
	last, err := excelize.CoordinatesToCellName(len(columns), len(rows)+1)
	if err != nil {
		return File{}, err
	}
	if err = book.SetSheetDimension(sheet, "A1:"+last); err != nil {
		return File{}, err
	}
	writer, err := book.NewStreamWriter(sheet)
	if err != nil {
		return File{}, err
	}
	if err = writer.SetRow("A1", head); err != nil {
		return File{}, err
	}
	for n, row := range rows {
		values := make([]any, len(row))
		for i, value := range row {
			// Excel keeps only 15 significant digits in numeric cells.
			values[i] = value
			if columns[i].Number && len(strings.TrimPrefix(value, "-")) <= 15 {
				if number, err := strconv.ParseInt(value, 10, 64); err == nil {
					values[i] = number
				}
			}
		}
		if err = writer.SetRow("A"+strconv.Itoa(n+2), values); err != nil {
			return File{}, err
		}
	}
	if err = writer.Flush(); err != nil {
		return File{}, err
	}
	out, err := book.WriteToBuffer()
	if err != nil {
		return File{}, err
	}
	return File{out.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}, nil
}
