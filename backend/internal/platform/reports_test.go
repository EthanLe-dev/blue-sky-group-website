package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

func TestM6ReportCalendarLedgerAndPermissions(t *testing.T) {
	a, staff, departure := bookingFixture(t, 5)
	ctx := context.Background()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := a.DB.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	q := quoteFor(t, a, staff, departure, booking.Counts{Adult: 2})
	input := bookingInput(q)
	input.Passengers = append(input.Passengers, input.Passengers[0])
	raw, err := a.BookingServer().Create(ctx, staff, uuid.NewString(), input)
	if err != nil {
		t.Fatal(err)
	}
	var record struct {
		ID    string
		Total string `json:"total_amount"`
	}
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	exec("UPDATE bookings SET status='confirmed',paid_amount=1000,refunded_amount=250,created_at='2026-01-01 16:59:59Z' WHERE id=$1", record.ID)
	exec("UPDATE seat_holds SET status='converted' WHERE booking_id=$1", record.ID)
	exec("UPDATE departures SET held_seats=0,confirmed_seats=2,starts_at='2026-01-01 17:00:00Z',sales_close_at='2026-01-01 16:00:00Z',balance_due_at='2026-01-01 15:00:00Z' WHERE id=$1", departure)
	exec(`WITH p AS (INSERT INTO payments(booking_id,purpose,amount,status,expires_at,paid_at)
		VALUES($1,'deposit',1000,'succeeded','2026-01-01 18:00:00Z','2026-01-01 17:00:00Z') RETURNING id)
		INSERT INTO refunds(payment_id,amount,reason,requested_by_staff_id,completed_at)
		SELECT id,250,'Hoàn demo',$2,'2026-01-02 17:00:00Z' FROM p`, record.ID, staff.ID)
	exec(`INSERT INTO group_requests(code,group_type,organization_name,contact_name,contact_email,contact_phone,participant_count,destination_text,status,consent_version,created_at)
		VALUES('REPORT-WON','school','Demo','Demo','demo@example.com','+84901234567',20,'Ninh Chữ','won','sales-v1','2026-01-01 18:00:00Z'),
		('REPORT-NEW','company','Demo','Demo','demo@example.com','+84901234567',20,'Ninh Chữ','new','sales-v1','2026-01-01 18:00:00Z')`)
	period := reportPeriod{From: "2026-01-01", To: "2026-01-03", Bucket: "day"}
	report, err := a.revenue(ctx, period)
	if err != nil {
		t.Fatal(err)
	}
	if report.Gross != record.Total || report.Count != 1 || report.Received != "1000" || report.Refunded != "250" || report.Net != "750" || len(report.Series) != 3 ||
		report.Series[0].Gross != record.Total || report.Series[1].Received != "1000" || report.Series[2].Net != "-250" {
		t.Fatalf("ledger grouped incorrectly: %+v", report)
	}
	_, sid, err := a.createSession(ctx, staff.ID)
	if err != nil {
		t.Fatal(err)
	}
	token, err := a.access(staff.ID, sid)
	if err != nil {
		t.Fatal(err)
	}
	call := func(path string, status int) []byte {
		t.Helper()
		w := request(a, "GET", "/api/v1"+path, "", token, "", nil)
		if w.Code != status {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
		return w.Body.Bytes()
	}
	const filter = "?from=2026-01-01&to=2026-01-03"
	if body := call("/admin/reports/occupancy"+filter, 200); !strings.Contains(string(body), `"occupancy_percent": 40.00`) &&
		!strings.Contains(string(body), `"occupancy_percent":40.00`) {
		t.Fatal(string(body))
	}
	quotes := call("/admin/reports/group-quotes"+filter+"&group_type=school", 200)
	var cohort struct {
		Data struct {
			Count int     `json:"request_count"`
			Won   float64 `json:"won_rate_percent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(quotes, &cohort); err != nil || cohort.Data.Count != 1 || cohort.Data.Won != 100 {
		t.Fatal(string(quotes), err)
	}
	call("/admin/reports/revenue?from=2026-01-01&to=2027-01-03", 422)
	book, err := excelize.OpenReader(bytes.NewReader(call("/admin/reports/revenue/export"+filter+"&format=xlsx", 200)))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	if value, err := book.GetCellValue(book.GetSheetName(0), "F2"); err != nil || value != "750" {
		t.Fatal(value, err)
	}
	call("/admin/bookings/export?status=confirmed&format=csv", 200)
	var audits int
	if err := a.DB.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE action='booking.export'").Scan(&audits); err != nil || audits != 1 {
		t.Fatal(audits, err)
	}
	call("/admin/tours/export?format=csv", 200)
	var slug string
	if err := a.DB.QueryRow(ctx, "SELECT t.slug FROM tours t JOIN departures d ON t.id=d.tour_id WHERE d.id=$1", departure).Scan(&slug); err != nil {
		t.Fatal(err)
	}
	if body := call("/tours/"+slug+"/itinerary.pdf", 200); !bytes.HasPrefix(body, []byte("%PDF-")) {
		t.Fatal("tour PDF missing")
	}
	var quoteID string
	if err := a.DB.QueryRow(ctx, `INSERT INTO group_quotes(group_request_id,version,title,itinerary,total_amount,currency,valid_until,created_by_staff_id)
		SELECT id,1,'Trường học','Ngày đi',1650000,'VND',current_date+10,$1 FROM group_requests WHERE code='REPORT-NEW' RETURNING id`, staff.ID).Scan(&quoteID); err != nil {
		t.Fatal(err)
	}
	quotePath := fmt.Sprintf("/admin/group-quotes/%s/document.pdf", quoteID)
	if body := call(quotePath, 200); !bytes.HasPrefix(body, []byte("%PDF-")) {
		t.Fatal("quote PDF missing")
	}
	exec("UPDATE tours SET status='archived' WHERE slug=$1", slug)
	call("/tours/"+slug+"/itinerary.pdf", 404)
	exec("DELETE FROM staff_roles WHERE staff_id=$1", staff.ID)
	exec("INSERT INTO staff_roles(staff_id,role_id) SELECT $1,id FROM roles WHERE code='executive'", staff.ID)
	call("/admin/reports/revenue"+filter, 200)
	call("/admin/reports/revenue/export"+filter, 403)
	call("/admin/bookings/export", 403)
	call(quotePath, 403)
}
