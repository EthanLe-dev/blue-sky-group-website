package booking

import (
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"testing"
	"time"
)

func TestAgeAtDepartureBirthdayBoundary(t *testing.T) {
	two, twelve := 2, 12
	q := Quote{Counts: Counts{Adult: 1, Child: 1, Infant: 1}, Snapshot: Snapshot{Start: time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC), Prices: []catalog.Price{{Type: "adult", Min: 12}, {Type: "child", Min: 2, Max: &twelve}, {Type: "infant", Min: 0, Max: &two}}}}
	p := []Passenger{{Type: "adult", Name: "Người lớn", Birth: "2018-06-15"}, {Type: "child", Name: "Trẻ em", Birth: "2018-06-16"}, {Type: "infant", Name: "Em bé", Birth: "2028-06-16"}}
	if !validatePassengers(p, q) {
		t.Fatal("birthday ages rejected")
	}
	p[2].Birth = "2028-06-15"
	if validatePassengers(p, q) {
		t.Fatal("second birthday still infant")
	}
}
