package catalog

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestR2AdapterUsesSignedS3Request(t *testing.T) {
	var received bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/public/file.png" || !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
			t.Error("unexpected unsigned S3 request")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test-image" {
			t.Error("wrong body")
		}
		received = true
		w.Header().Set("ETag", `"test"`)
		w.WriteHeader(200)
	}))
	defer server.Close()
	t.Setenv("STORAGE_PROVIDER", "r2")
	t.Setenv("R2_ENDPOINT", server.URL)
	t.Setenv("R2_ACCESS_KEY_ID", "test-key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test-secret")
	t.Setenv("R2_PUBLIC_BUCKET", "public")
	store := NewMediaStore("test-signing-key")
	f, e := os.CreateTemp(t.TempDir(), "r2*")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	f.WriteString("test-image")
	f.Seek(0, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e = store.put(ctx, "file.png", "public", "image/png", f, 10); e != nil {
		t.Fatal(e)
	}
	if !received {
		t.Fatal("no S3 request")
	}
}
func TestDepartureValidationChecksSurcharges(t *testing.T) {
	start := time.Now().Add(30 * 24 * time.Hour)
	b := DepartureWrite{Start: start, End: start.Add(24 * time.Hour), Close: start.Add(-time.Hour), Due: start.Add(-7 * 24 * time.Hour), Capacity: 10, Deposit: 3000, Prices: []Price{{Type: "infant", Min: 0, Max: intPointer(2), Seats: 0}, {Type: "child", Min: 2, Max: intPointer(12), Seats: 1, Amount: 500}, {Type: "adult", Min: 12, Seats: 1, Amount: 1000}}, Surcharges: []Surcharge{{Code: "bad", Label: "Bad unit", Amount: 10, Unit: "unknown"}}}
	if validDeparture(b) {
		t.Fatal("invalid surcharge accepted")
	}
	b.Surcharges = nil
	if !validDeparture(b) {
		t.Fatal("valid departure rejected")
	}
}
