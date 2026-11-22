package platform

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestOpenAPIResponses(t *testing.T) {
	doc, e := openapi3.NewLoader().LoadFromFile("openapi.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = doc.Validate(context.Background()); e != nil {
		t.Fatal(e)
	}
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("spec valid; TEST_DATABASE_URL required for HTTP responses")
	}
	a := testApp(t)
	router, e := legacy.NewRouter(doc)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct{ method, path, body string }{{"GET", "/api/v1/tours", ""}, {"GET", "/api/v1/tours/not-existing", ""}, {"GET", "/api/v1/home", ""}, {"GET", "/api/v1/gallery", ""}, {"GET", "/healthz", ""}, {"GET", "/readyz", ""}, {"GET", "/api/v1/booking-defaults", ""}, {"POST", "/api/v1/guest-sessions", "{}"}, {"POST", "/api/v1/admin/auth/login", `{"email":"admin@example.com","password":"test-password-123","turnstile_token":"local-test-token"}`}} {
		req := httptest.NewRequest(tc.method, "http://localhost:8080"+tc.path, strings.NewReader(tc.body))
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		a.Router().ServeHTTP(w, req)
		route, params, e := router.FindRoute(req)
		if e != nil {
			t.Fatal(e)
		}
		input := &openapi3filter.RequestValidationInput{Request: req, PathParams: params, Route: route}
		out := &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: w.Code, Header: w.Header(), Body: io.NopCloser(strings.NewReader(w.Body.String()))}
		if e = openapi3filter.ValidateResponse(context.Background(), out); e != nil {
			t.Fatalf("%s: %v", tc.path, e)
		}
	}
}
