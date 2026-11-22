package platform

import (
	"context"
	"testing"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
)

func TestReleaseSeedCanBeRerun(t *testing.T) {
	a, _, _ := bookingFixture(t, 5)
	ctx := context.Background()
	server := a.BookingServer()
	for range 2 {
		if err := server.SeedDemoStates(ctx); err != nil {
			t.Fatal(err)
		}
	}
	var international, leads, paid, deposit, expired int
	err := a.DB.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM tours WHERE id=$1 AND region='international'),
		(SELECT count(*) FROM group_requests WHERE code LIKE 'DEMO-DOAN-%'),
		(SELECT count(*) FROM bookings b JOIN departures d ON d.id=b.departure_id WHERE d.id=$2 AND b.payment_status='paid' AND d.confirmed_seats=d.capacity),
		(SELECT count(*) FROM bookings b JOIN departures d ON d.id=b.departure_id WHERE d.id=$3 AND b.payment_status='partially_paid' AND d.capacity-d.confirmed_seats=1),
		(SELECT count(*) FROM bookings b JOIN departures d ON d.id=b.departure_id WHERE d.id=$4 AND b.status='expired' AND d.held_seats=0)`,
		catalog.DemoID("mvp-international"), catalog.DemoID("mvp-release:full"),
		catalog.DemoID("mvp-release:deposit"), catalog.DemoID("mvp-release:expired")).
		Scan(&international, &leads, &paid, &deposit, &expired)
	if err != nil || international != 1 || leads != 2 || paid != 1 || deposit != 1 || expired != 1 {
		t.Fatalf("fixtures: %d/%d/%d/%d/%d: %v", international, leads, paid, deposit, expired, err)
	}
	if _, err = a.DB.Exec(ctx, "UPDATE departures SET capacity=5 WHERE id=$1", catalog.DemoID("mvp-release:full")); err != nil {
		t.Fatal(err)
	}
	if err = server.SeedDemoStates(ctx); err != nil {
		t.Fatal(err)
	}
	var capacity int
	if err = a.DB.QueryRow(ctx, "SELECT capacity FROM departures WHERE id=$1", catalog.DemoID("mvp-release:full")).Scan(&capacity); err != nil || capacity != 5 {
		t.Fatalf("seed overwrote admin edits: %d %v", capacity, err)
	}
	t.Setenv("APP_ENV", "production")
	if err = server.SeedDemoStates(ctx); err == nil {
		t.Fatal("production accepted demo data")
	}
}
