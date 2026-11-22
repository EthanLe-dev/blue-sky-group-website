package platform

import (
	"context"
	"testing"
)

func TestCatalogSchemaConstraints(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	var tour string
	e := a.DB.QueryRow(ctx, "INSERT INTO tours(slug,code,title,region,audience,duration_days,duration_nights,origin_name,transport) VALUES('schema-test','SCHEMA','Tour mẫu','domestic','general',2,1,'TP.HCM','Xe') RETURNING id").Scan(&tour)
	if e != nil {
		t.Fatal(e)
	}
	for _, q := range []string{`INSERT INTO departures(tour_id,starts_at,ends_at,sales_close_at,capacity,deposit_bps,balance_due_at) VALUES($1,now()+interval '30 days',now()+interval '31 days',now()+interval '29 days',-1,3000,now()+interval '23 days')`, `INSERT INTO departures(tour_id,starts_at,ends_at,sales_close_at,capacity,deposit_bps,balance_due_at) VALUES($1,now()+interval '30 days',now()+interval '29 days',now()+interval '28 days',30,3000,now()+interval '23 days')`} {
		if _, e = a.DB.Exec(ctx, q, tour); e == nil {
			t.Fatal("invalid inventory/date accepted")
		}
	}
	var d string
	e = a.DB.QueryRow(ctx, `INSERT INTO departures(tour_id,starts_at,ends_at,sales_close_at,capacity,deposit_bps,balance_due_at) VALUES($1,now()+interval '30 days',now()+interval '31 days',now()+interval '29 days',5,3000,now()+interval '23 days') RETURNING id`, tour).Scan(&d)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = a.DB.Exec(ctx, "UPDATE departures SET held_seats=6 WHERE id=$1", d); e == nil {
		t.Fatal("overcapacity accepted")
	}
	if _, e = a.DB.Exec(ctx, "INSERT INTO departure_prices(departure_id,passenger_type,unit_price,seats_per_person,min_age) VALUES($1,'adult',-1,1,12)", d); e == nil {
		t.Fatal("negative money accepted")
	}
}
