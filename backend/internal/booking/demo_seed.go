package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/jackc/pgx/v5"
)

// SeedDemoStates adds isolated examples and leaves existing bookings, departures, and prices alone.
func (s *Server) SeedDemoStates(ctx context.Context) error {
	switch os.Getenv("APP_ENV") {
	case "local", "test", "demo":
	default:
		return errors.New("demo seed forbidden")
	}
	guard, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer guard.Rollback(ctx)
	if _, err = guard.Exec(ctx, "SELECT pg_advisory_xact_lock(735297)"); err != nil {
		return err
	}
	var admin string
	if err = s.DB.QueryRow(ctx, `SELECT u.id FROM staff_users u JOIN staff_roles sr ON sr.staff_id=u.id
		JOIN roles r ON r.id=sr.role_id WHERE r.code='admin' AND u.status='active' ORDER BY u.created_at LIMIT 1`).Scan(&admin); err != nil {
		return err
	}
	if err = s.seedInternational(ctx); err != nil {
		return err
	}

	for i, example := range []struct {
		label   string
		purpose string
		seats   int
	}{
		{"Hết chỗ", "full", 1},
		{"Còn một chỗ", "deposit", 2},
		{"Đã hết hạn giữ chỗ", "expired", 3},
	} {
		departure := catalog.DemoID("mvp-release:" + example.purpose)
		start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 60+i*7).Add(22 * time.Hour)
		if _, err = s.DB.Exec(ctx, `WITH new_departure AS (
			INSERT INTO departures(id,tour_id,starts_at,ends_at,sales_close_at,balance_due_at,capacity,status,deposit_bps)
			SELECT $1,t.id,$2,$3,$4,$5,$6,'open',3000 FROM tours t WHERE id=$7
			ON CONFLICT DO NOTHING RETURNING id
		)
		INSERT INTO departure_prices(departure_id,passenger_type,unit_price,seats_per_person,min_age,max_age_exclusive)
		SELECT d.id,p.passenger_type,p.unit_price,p.seats_per_person,p.min_age,p.max_age_exclusive
		FROM new_departure d CROSS JOIN departure_prices p WHERE p.departure_id=$8`,
			departure, start, start.Add(36*time.Hour), start.Add(-24*time.Hour), start.AddDate(0, 0, -7),
			example.seats, catalog.DemoID("vinh-hy-doanh-nghiep"), catalog.DemoID("vinh-hy-doanh-nghiep-departure-0")); err != nil {
			return err
		}
		marker := "Dữ liệu demo · " + example.label
		var id string
		err = s.DB.QueryRow(ctx, "SELECT id FROM bookings WHERE departure_id=$1 AND contact->>'email'=$2",
			departure, example.purpose+"@blue-sky.example").Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			principal := Principal{ID: admin, Staff: true}
			quote, e := s.Quote(ctx, principal, QuoteInput{Departure: departure, Counts: Counts{Adult: 1}, Codes: []string{}})
			if e != nil {
				return e
			}
			option := example.purpose
			if option == "expired" {
				option = "full"
			}
			out, e := s.Create(ctx, principal, catalog.DemoID("mvp-booking:"+example.purpose), CreateInput{
				Quote: quote.ID, Source: "counter", Option: option, Requests: marker,
				Contact:    Contact{Name: "Khách minh họa", Email: example.purpose + "@blue-sky.example", Phone: "+84900000000"},
				Passengers: []Passenger{{Type: "adult", Name: "Khách minh họa", Birth: "1990-01-01"}},
				Consent:    Consent{Accepted: true, Version: "booking-v1"},
			})
			if e != nil {
				return e
			}
			var b struct{ ID string }
			if e = json.Unmarshal(out, &b); e != nil {
				return e
			}
			id = b.ID
		} else if err != nil {
			return err
		}
		if err = s.finishDemoBooking(ctx, id, admin, example.purpose, marker); err != nil {
			return err
		}
	}

	_, err = s.DB.Exec(ctx, `INSERT INTO group_requests(id,code,group_type,organization_name,contact_name,contact_email,
		contact_phone,participant_count,destination_text,requirements,consent_version)
		VALUES($1,'DEMO-DOAN-TRUONG','school','Trường mẫu Blue Sky','Người liên hệ minh họa',
		'truong@blue-sky.example','+84900000000',32,'Ninh Chữ','Yêu cầu mẫu cho buổi nghiệm thu','sales-v1'),
		($2,'DEMO-DOAN-CONGTY','company','Doanh nghiệp mẫu Blue Sky','Người liên hệ minh họa',
		'cong-ty@blue-sky.example','+84900000000',20,'Vĩnh Hy','Yêu cầu mẫu cho buổi nghiệm thu','sales-v1')
		ON CONFLICT DO NOTHING`, catalog.DemoID("mvp-school-lead"), catalog.DemoID("mvp-company-lead"))
	return err
}

func (s *Server) finishDemoBooking(ctx context.Context, id, staff, purpose, marker string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	departure, err := s.lockBooking(ctx, tx, id)
	if err != nil {
		return err
	}
	var untouched bool
	if err = tx.QueryRow(ctx, `SELECT status='pending' AND version=1 AND special_requests=$2
		FROM bookings WHERE id=$1`, id, marker).Scan(&untouched); err != nil {
		return err
	}
	if !untouched {
		return tx.Commit(ctx)
	}
	if purpose == "expired" {
		if _, err = tx.Exec(ctx, `WITH shortened AS (
			UPDATE bookings SET expires_at=clock_timestamp()-interval '1 second' WHERE id=$1 RETURNING expires_at
		) UPDATE seat_holds SET expires_at=(SELECT expires_at FROM shortened) WHERE booking_id=$1`, id); err != nil {
			return err
		}
		err = ExpireLocked(ctx, tx, departure)
	} else {
		var payment struct{ ID string }
		raw, e := createPayment(ctx, tx, id, purpose, staff)
		if e != nil {
			return e
		}
		if e = json.Unmarshal(raw, &payment); e != nil {
			return e
		}
		_, err = confirmPayment(ctx, tx, id, departure, payment.ID, staff, marker, "seed-demo-fixtures")
	}
	if err != nil {
		return fmt.Errorf("complete demo state: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Server) seedInternational(ctx context.Context) error {
	start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 45).Add(2 * time.Hour)
	_, err := s.DB.Exec(ctx, `WITH new_tour AS (
		INSERT INTO tours(id,slug,code,title,summary,description,region,audience,duration_days,duration_nights,
			origin_name,transport,status,is_demo,published_at)
		VALUES($1,'singapore-minh-hoa','DEMO-SG','Singapore · Hành trình minh họa',
			'Chương trình mẫu ba ngày để thử bộ lọc tour quốc tế.',
			'<p>Lịch trình và giá giả định cho bản demo.</p>','international','general',3,2,
			'TP. Hồ Chí Minh','Máy bay','published',true,clock_timestamp())
		ON CONFLICT DO NOTHING RETURNING id
	), itinerary AS (
		INSERT INTO tour_days(tour_id,day_number,title,body)
		SELECT id,day,title,body FROM new_tour CROSS JOIN (VALUES
			(1,'Đến Singapore','<p>Đón đoàn và nhận phòng.</p>'),
			(2,'Khám phá thành phố','<p>Tham quan khu Marina Bay.</p>'),
			(3,'Kết thúc hành trình','<p>Trả phòng và trở về.</p>')) AS t(day,title,body)
	), new_departure AS (
		INSERT INTO departures(id,tour_id,starts_at,ends_at,sales_close_at,balance_due_at,capacity,status,deposit_bps)
		SELECT $2,id,$3,$4,$5,$6,12,'open',3000 FROM new_tour RETURNING id
	)
	INSERT INTO departure_prices(departure_id,passenger_type,unit_price,seats_per_person,min_age,max_age_exclusive)
	SELECT id,passenger_type,price,seats,min_age,max_age FROM new_departure CROSS JOIN (VALUES
		('adult',8900000,1,12,NULL::int),('child',6900000,1,2,12),('infant',0,0,0,2)) AS t(passenger_type,price,seats,min_age,max_age)`,
		catalog.DemoID("mvp-international"), catalog.DemoID("mvp-intl-departure"), start, start.Add(60*time.Hour), start.AddDate(0, 0, -2), start.AddDate(0, 0, -7))
	return err
}
