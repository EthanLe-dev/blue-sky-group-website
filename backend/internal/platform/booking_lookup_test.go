package platform

import "testing"

func TestLookupLimiterIsBoundToBookingCode(t *testing.T) {
	a := testApp(t)
	for i := 0; i < 6; i++ {
		w := request(a, "POST", "/api/v1/booking-lookups", `{"code":"BS-FAKE","contact":"demo@example.com","turnstile_token":"local-test-token"}`, "", "", nil)
		expected := 404
		if i == 5 {
			expected = 429
		}
		if w.Code != expected {
			t.Fatalf("attempt %d got %d", i, w.Code)
		}
	}
	w := request(a, "POST", "/api/v1/booking-lookups", `{"code":"BS-OTHER","contact":"demo@example.com","turnstile_token":"local-test-token"}`, "", "", nil)
	if w.Code != 404 {
		t.Fatalf("different code shares lookup bucket: %d", w.Code)
	}
}
