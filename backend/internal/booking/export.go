package booking

import (
	"errors"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/exportfile"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
)

func (s *Server) Export(c *gin.Context) {
	format := c.DefaultQuery("format", "xlsx")
	if format != "xlsx" && format != "csv" {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	where, args, ok := bookingFilters(c, true)
	if !ok {
		return
	}
	staff, ok := s.Principal(c, true)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	failed := func() { httpx.Error(c, 503, "EXPORT_UNAVAILABLE") }
	if err := s.Sweep(ctx); err != nil {
		failed()
		return
	}
	rows, err := s.DB.Query(ctx, `SELECT b.code,b.tour_snapshot->>'title',b.status,b.payment_status,b.source,
		b.contact->>'name',b.contact->>'email',b.contact->>'phone',b.total_amount::text,b.paid_amount::text,b.refunded_amount::text,
		to_char(b.created_at AT TIME ZONE 'Asia/Ho_Chi_Minh','YYYY-MM-DD HH24:MI:SS')
		FROM bookings b`+where+" ORDER BY b.created_at DESC,b.id LIMIT 10001", args...)
	if err != nil {
		failed()
		return
	}
	defer rows.Close()
	var values [][]string
	for rows.Next() {
		row := make([]string, 12)
		if err := rows.Scan(&row[0], &row[1], &row[2], &row[3], &row[4], &row[5], &row[6], &row[7],
			&row[8], &row[9], &row[10], &row[11]); err != nil {
			failed()
			return
		}
		values = append(values, row)
	}
	if rows.Err() != nil {
		failed()
		return
	}
	cols := []exportfile.Column{{Title: "Mã đơn"}, {Title: "Tour"}, {Title: "Trạng thái"}, {Title: "Thanh toán"}, {Title: "Nguồn"},
		{Title: "Liên hệ"}, {Title: "Email"}, {Title: "Điện thoại"}, {Title: "Giá trị đơn (VND)", Number: true},
		{Title: "Đã thu (VND)", Number: true}, {Title: "Đã hoàn (VND)", Number: true}, {Title: "Ngày tạo · giờ Việt Nam"}}
	file, err := exportfile.Table(format, cols, values)
	if errors.Is(err, exportfile.ErrTooLarge) {
		httpx.Error(c, 422, "EXPORT_TOO_LARGE")
		return
	}
	if err != nil {
		failed()
		return
	}
	if _, err := s.DB.Exec(ctx, `INSERT INTO audit_logs(actor_staff_id,action,entity_type,request_id,changed_fields)
		VALUES($1,'booking.export','booking',$2,jsonb_build_object('format',$3::text,'rows',$4::int))`,
		staff.ID, c.GetString("request_id"), format, len(values)); err != nil {
		failed()
		return
	}
	file.Download(c, "danh-sach-don."+format)
}
