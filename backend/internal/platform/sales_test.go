package platform

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestSalesIntakeAndQuoteWorkflow(t *testing.T) {
	a := testApp(t)
	guest := request(a, "POST", "/api/v1/guest-sessions", "{}", "", "", nil)
	csrf := data(t, guest)["csrf_token"].(string)
	cookie := guest.Result().Cookies()[0]
	login := request(a, "POST", "/api/v1/admin/auth/login", `{"email":"admin@example.com","password":"test-password-123","turnstile_token":"local-test-token"}`, "", "", nil)
	bearer := data(t, login)["access_token"].(string)

	check := func(w *httptest.ResponseRecorder, want int) {
		t.Helper()
		if w.Code != want {
			t.Fatalf("status=%d want=%d: %s", w.Code, want, w.Body.String())
		}
	}
	public := func(path, body, key, csrfToken string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/v1/"+path, strings.NewReader(body))
		r.Header.Set("Origin", "http://localhost:5173")
		r.Header.Set("X-CSRF-Token", csrfToken)
		r.Header.Set("Idempotency-Key", key)
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		a.Router().ServeHTTP(w, r)
		return w
	}
	key := uuid.NewString()
	input := `{"group_type":"school","organization_name":"Trường Demo","contact_name":"Khách Demo","contact_email":"demo@example.com","contact_phone":"0901234567","participant_count":30,"destination_text":"Ninh Chữ","consent":{"accepted":true,"policy_version":"sales-v1"},"turnstile_token":"local-test-token"}`
	badCaptcha := strings.Replace(input, "local-test-token", "consumed-token", 1)
	path := "group-requests"
	check(public(path, input, key, ""), 403)
	check(public(path, badCaptcha, key, csrf), 403)
	check(public(path, strings.Replace(input, `"accepted":true`, `"accepted":false`, 1), key, csrf), 422)
	check(public(path, strings.Replace(input, `"participant_count":30`, `"participant_count":30,"preferred_start_date":"2000-01-01"`, 1), key, csrf), 422)
	var responses [2]*httptest.ResponseRecorder
	var wg sync.WaitGroup
	for i := range responses {
		wg.Go(func() { responses[i] = public(path, input, key, csrf) })
	}
	wg.Wait()
	check(responses[0], 201)
	check(responses[1], 201)
	receipt := data(t, responses[0])
	if receipt["code"] != data(t, responses[1])["code"] || receipt["response_sla_message"] == "" || receipt["contact_email"] != nil {
		t.Fatalf("invalid receipt: %v", receipt)
	}
	check(public(path, badCaptcha, key, csrf), 201)
	check(public(path, strings.Replace(input, "Ninh Chữ", "Huế", 1), key, csrf), 409)
	callback := `{"phone":"0901234567","preferred_time":"Buổi chiều","consent":{"accepted":true,"policy_version":"sales-v1"},"turnstile_token":"local-test-token"}`
	check(public("callback-requests", callback, uuid.NewString(), csrf), 201)

	ctx := context.Background()
	var groupID, callbackID, staff string
	for sql, out := range map[string]*string{
		"SELECT id FROM group_requests":                    &groupID,
		"SELECT id FROM callback_requests":                 &callbackID,
		"SELECT id FROM staff_users WHERE status='active'": &staff,
	} {
		if err := a.DB.QueryRow(ctx, sql).Scan(out); err != nil {
			t.Fatal(err)
		}
	}
	admin := func(method, path, body string) *httptest.ResponseRecorder {
		return request(a, method, "/api/v1/admin/"+path, body, bearer, "", nil)
	}
	group := path + "/" + groupID
	cb := "callback-requests/" + callbackID
	check(request(a, "GET", "/api/v1/admin/"+group, "", "", csrf, cookie), 401)
	check(admin("GET", path+"?status=new", ""), 200)
	check(admin("PATCH", cb, `{"status":"closed"}`), 409)
	check(admin("PATCH", cb, `{"status":"contacted"}`), 200)
	check(admin("PATCH", cb, `{"status":"closed"}`), 200)
	check(admin("POST", cb+"/notes", `{"body":"Đã gọi lại"}`), 201)
	check(admin("PATCH", group, `{"status":"won"}`), 409)
	if _, err := a.DB.Exec(ctx, `INSERT INTO staff_users(email,password_hash,full_name,status) VALUES('disabled@example.com','invalid','Disabled','disabled')`); err != nil {
		t.Fatal(err)
	}
	var disabled string
	if err := a.DB.QueryRow(ctx, `SELECT id FROM staff_users WHERE status='disabled'`).Scan(&disabled); err != nil {
		t.Fatal(err)
	}
	check(admin("PATCH", group, `{"assigned_to":"`+disabled+`"}`), 422)
	check(admin("PATCH", group, `{"status":"preparing","assigned_to":"`+staff+`"}`), 200)
	q := `{"title":"Tour Demo","itinerary":"Hai ngày","inclusions":"Xe","exclusions":"","total_amount":"123000","currency":"VND","valid_until":"9999-01-01"}`
	draft := admin("POST", group+"/quotes", q)
	check(draft, 201)
	quote := "group-quotes/" + data(t, draft)["id"].(string)
	check(admin("PATCH", quote, `{"title":"Tour Demo sửa"}`), 200)
	check(admin("POST", quote+"/issue", "{}"), 200)
	check(admin("PATCH", quote, `{"total_amount":"1"}`), 409)
	revision := admin("POST", group+"/quotes", q)
	check(revision, 201)
	if data(t, revision)["version"] != float64(2) {
		t.Fatal(revision.Body.String())
	}
	check(admin("POST", quote+"/decision", `{"status":"accepted","reason":"Xác nhận qua điện thoại"}`), 200)
	detail := admin("GET", group, "")
	check(detail, 200)
	if data(t, detail)["status"] != "won" {
		t.Fatal(detail.Body.String())
	}
	check(admin("PATCH", group, `{"status":"lost"}`), 409)
}
