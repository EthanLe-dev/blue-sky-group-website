package platform

import (
	"context"
	"encoding/json"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/google/uuid"
	"testing"
)

func TestCapacityEditExpiresHoldsWhileLocked(t *testing.T) {
	a, p, dep := bookingFixture(t, 3)
	ctx := context.Background()
	q := quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	out, e := a.BookingServer().Create(ctx, p, uuid.NewString(), bookingInput(q))
	if e != nil {
		t.Fatal(e)
	}
	var d map[string]any
	json.Unmarshal(out, &d)
	id := d["id"].(string)
	a.DB.Exec(ctx, "UPDATE seat_holds SET expires_at=clock_timestamp()-interval '1 second' WHERE booking_id=$1", id)
	_, sid, _ := a.createSession(ctx, p.ID)
	token, _ := a.access(p.ID, sid)
	w := request(a, "PATCH", "/api/v1/admin/departures/"+dep, `{"capacity":1,"expected_price_version":1}`, token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	var status string
	var held int
	a.DB.QueryRow(ctx, "SELECT b.status,d.held_seats FROM bookings b JOIN departures d ON d.id=b.departure_id WHERE b.id=$1", id).Scan(&status, &held)
	if status != "expired" || held != 0 {
		t.Fatal("capacity kept stale hold", status, held)
	}
}
