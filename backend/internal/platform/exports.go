package platform

import (
	"errors"
	"strconv"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/exportfile"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
)

func exportReply(c *gin.Context, file exportfile.File, filename string, err error) {
	switch {
	case errors.Is(err, exportfile.ErrTooLarge):
		httpx.Error(c, 422, "EXPORT_TOO_LARGE")
	case err != nil:
		httpx.Error(c, 503, "EXPORT_UNAVAILABLE")
	default:
		file.Download(c, filename)
	}
}

func exportFormat(c *gin.Context) (string, bool) {
	format := c.DefaultQuery("format", "xlsx")
	if format != "xlsx" && format != "csv" {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return "", false
	}
	return format, true
}

func (a *App) exportRevenue(c *gin.Context) {
	period, valid := parseReport(c)
	if !valid {
		return
	}
	format, ok := exportFormat(c)
	if !ok {
		return
	}
	report, err := a.revenue(c.Request.Context(), period)
	if err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	columns := []exportfile.Column{
		{Title: "Kỳ · giờ Việt Nam"}, {Title: "Số đơn", Number: true}, {Title: "Giá trị đơn (VND)", Number: true},
		{Title: "Đã thu (VND)", Number: true}, {Title: "Đã hoàn (VND)", Number: true}, {Title: "Thu ròng (VND)", Number: true},
		{Title: "Môi trường"},
	}
	rows := make([][]string, 0, len(report.Series)+1)
	for _, row := range append([]revenueRow{report.revenueRow}, report.Series...) {
		label := row.Period
		if label == "" {
			label = "Tổng"
		}
		rows = append(rows, []string{label, strconv.FormatInt(row.Count, 10), row.Gross, row.Received, row.Refunded, row.Net, "Giả lập"})
	}
	file, err := exportfile.Table(format, columns, rows)
	exportReply(c, file, "bao-cao-doanh-thu."+format, err)
}

func (a *App) exportTours(c *gin.Context) {
	format, ok := exportFormat(c)
	if !ok {
		return
	}
	status := c.Query("status")
	if status != "" && status != "draft" && status != "published" && status != "archived" {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	rows, err := a.DB.Query(c.Request.Context(), `
		SELECT slug,title,status,duration_days::text,duration_nights::text,origin_name,transport
		FROM tours WHERE ($1='' OR status=$1) ORDER BY title,id LIMIT $2`, status, exportfile.MaxRows+1)
	if err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	defer rows.Close()
	var values [][]string
	for rows.Next() {
		row := make([]string, 7)
		if err = rows.Scan(&row[0], &row[1], &row[2], &row[3], &row[4], &row[5], &row[6]); err != nil {
			exportReply(c, exportfile.File{}, "", err)
			return
		}
		values = append(values, row)
	}
	if err = rows.Err(); err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	cols := []exportfile.Column{{Title: "Slug"}, {Title: "Tên tour"}, {Title: "Trạng thái"}, {Title: "Số ngày", Number: true},
		{Title: "Số đêm", Number: true}, {Title: "Nơi đi"}, {Title: "Phương tiện"}}
	file, err := exportfile.Table(format, cols, values)
	exportReply(c, file, "danh-sach-tour."+format, err)
}
