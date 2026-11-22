package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestCRMOnlyListsVerifiedCustomerLinks(t *testing.T) {
	a, staff, departure := bookingFixture(t, 5)
	ctx := context.Background()
	login := func(email string) string {
		t.Helper()
		w := request(a, "POST", "/api/v1/admin/auth/login", fmt.Sprintf(`{"email":%q,"password":"test-password-123","turnstile_token":"local-test-token"}`, email), "", "", nil)
		if w.Code != 200 {
			t.Fatalf("login failed: %s", w.Body.String())
		}
		return data(t, w)["access_token"].(string)
	}
	admin := login("admin@example.com")
	check := func(w *httptest.ResponseRecorder, status int) *httptest.ResponseRecorder {
		t.Helper()
		if w.Code != status {
			t.Fatalf("status=%d want=%d: %s", w.Code, status, w.Body.String())
		}
		return w
	}
	api := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		return check(request(a, method, "/api/v1/admin/"+path, body, admin, "", nil), status)
	}
	list := func(w *httptest.ResponseRecorder) []map[string]any {
		t.Helper()
		var result struct{ Data []map[string]any }
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result.Data
	}
	create := func(name string) string {
		w := api("POST", "customers", fmt.Sprintf(`{"full_name":%q,"email":"DEMO@example.com","phone":"0901234567"}`, name), 201)
		record := data(t, w)
		if record["email"] != "demo@example.com" || record["phone"] != "+84901234567" {
			t.Fatal(w.Body.String())
		}
		return record["id"].(string)
	}
	first := create("Hồ sơ mẫu A")
	second := create("Hồ sơ mẫu B")
	if first == second {
		t.Fatal("matching contacts were merged")
	}
	if n := len(list(api("GET", "customers?q=0901234567", "", 200))); n != 2 {
		t.Fatalf("search returned %d", n)
	}
	api("POST", "customers", `{"full_name":"Thiếu liên hệ"}`, 422)
	customer := api("PATCH", "customers/"+first, `{"notes":"Ghi chú thử nghiệm","email":null}`, 200)
	if saved := data(t, customer); saved["email"] != nil || saved["notes"] != "Ghi chú thử nghiệm" {
		t.Fatal(customer.Body.String())
	}

	input := bookingInput(quoteFor(t, a, staff, departure, booking.Counts{Adult: 1}))
	raw, err := a.BookingServer().Create(ctx, staff, uuid.NewString(), input)
	if err != nil {
		t.Fatal(err)
	}
	var booked struct {
		ID      string
		Version int
	}
	if err = json.Unmarshal(raw, &booked); err != nil {
		t.Fatal(err)
	}
	// A booking linked before M5 stays hidden until a staff member verifies it.
	if _, err = a.DB.Exec(ctx, "UPDATE bookings SET customer_id=$2 WHERE id=$1", booked.ID, first); err != nil {
		t.Fatal(err)
	}
	history := "customers/" + first + "/bookings"
	if entries := list(api("GET", history, "", 200)); len(entries) != 0 {
		t.Fatal(entries)
	}

	patch := func(customerID string, version int, verified bool) string {
		return fmt.Sprintf(`{"expected_version":%d,"customer_id":%q,"customer_link_verified":%t}`, version, customerID, verified)
	}
	bookingPath := "bookings/" + booked.ID
	api("PATCH", bookingPath, patch(first, booked.Version, false), 422)
	api("PATCH", bookingPath, patch(first, booked.Version+1, true), 409)
	api("PATCH", bookingPath, patch(first, booked.Version, true), 200)
	entries := list(api("GET", history, "", 200))
	if len(entries) != 1 || entries[0]["id"] != booked.ID || entries[0]["verified_at"] == nil {
		t.Fatal(entries)
	}
	api("PATCH", bookingPath, patch(second, booked.Version+1, true), 409)
	if entries := list(api("GET", "customers/"+second+"/bookings", "", 200)); len(entries) != 0 {
		t.Fatal(entries)
	}

	input = bookingInput(quoteFor(t, a, staff, departure, booking.Counts{Adult: 1}))
	input.Customer = &second
	if _, err = a.BookingServer().Create(ctx, staff, uuid.NewString(), input); err == nil {
		t.Fatal("created an unverified link")
	}
	input.CustomerVerified = true
	if _, err = a.BookingServer().Create(ctx, staff, uuid.NewString(), input); err != nil {
		t.Fatal(err)
	}
	if entries := list(api("GET", "customers/"+second+"/bookings", "", 200)); len(entries) != 1 {
		t.Fatal(entries)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("test-password-123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	var accountant string
	if err = a.DB.QueryRow(ctx, `INSERT INTO staff_users(email,password_hash,full_name)
		VALUES('accountant@example.com',$1,'Kế toán mẫu') RETURNING id`, string(hash)).Scan(&accountant); err != nil {
		t.Fatal(err)
	}
	if _, err = a.DB.Exec(ctx, `INSERT INTO staff_roles(staff_id,role_id) SELECT $1,id FROM roles WHERE code='accountant'`, accountant); err != nil {
		t.Fatal(err)
	}
	check(request(a, "GET", "/api/v1/admin/"+history, "", "", "", nil), 401)
	token := login("accountant@example.com")
	check(request(a, "GET", "/api/v1/admin/"+history, "", token, "", nil), 403)
	check(request(a, "GET", "/api/v1/admin/group-requests", "", token, "", nil), 403)
}
