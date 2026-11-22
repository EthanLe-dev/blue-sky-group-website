package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	demoassets "github.com/EthanLe-dev/blue-sky-group-website/backend/demo-assets"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"
)

func catalogFixture(t *testing.T) (*App, string, string) {
	t.Helper()
	a := testApp(t)
	var staff string
	a.DB.QueryRow(context.Background(), "SELECT id FROM staff_users LIMIT 1").Scan(&staff)
	_, sid, _ := a.createSession(context.Background(), staff)
	token, _ := a.access(staff, sid)
	var b bytes.Buffer
	m := multipart.NewWriter(&b)
	f, _ := m.CreateFormFile("file", "photo.png")
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 4, 4)))
	m.Close()
	req := httptest.NewRequest("POST", "/api/v1/admin/media?visibility=public", &b)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", m.FormDataContentType())
	w := httptest.NewRecorder()
	t.Setenv("LOCAL_STORAGE_PATH", t.TempDir())
	a.Router().ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	media := data(t, w)["id"].(string)
	input := fmt.Sprintf(`{"slug":"ninh-chu-test","code":"NC-TEST","title":"Ninh Chữ mẫu","summary":"Hành trình demo","region":"domestic","audience":"general","duration_days":2,"duration_nights":1,"origin_name":"TP.HCM","transport":"Xe","difficulty":"easy","included":"Xe, khách sạn","excluded":"Chi phí cá nhân","cancellation_policy":"Nhân viên xét duyệt","cover_media_id":"%s","days":[{"day_number":1,"title":"Ngày 1","body":"Biển"},{"day_number":2,"title":"Ngày 2","body":"Trở về"}]}`, media)
	w = request(a, "POST", "/api/v1/admin/tours", input, token, "", nil)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	id := data(t, w)["id"].(string)
	return a, token, id
}
func TestCatalogCRUDFilterAndAtomicBulk(t *testing.T) {
	a, token, tour := catalogFixture(t)
	w := request(a, "POST", "/api/v1/admin/tours/"+tour+"/publish", `{}`, token, "", nil)
	if w.Code != 422 {
		t.Fatal("published without prices")
	}
	start := time.Now().AddDate(0, 0, 30).UTC()
	body := func(start time.Time, price string) string {
		return fmt.Sprintf(`{"starts_at":%q,"ends_at":%q,"capacity":5,"prices":[{"passenger_type":"adult","unit_price":%q},{"passenger_type":"child","unit_price":"1000000"},{"passenger_type":"infant","unit_price":"0"}]}`, start.Format(time.RFC3339), start.AddDate(0, 0, 1).Format(time.RFC3339), price)
	}
	w = request(a, "POST", "/api/v1/admin/tours/"+tour+"/departures", body(start, "1650000"), token, "", nil)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	dep := data(t, w)["id"].(string)
	w = request(a, "POST", "/api/v1/admin/departures/"+dep+"/sales-status", `{"status":"open"}`, token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	w = request(a, "POST", "/api/v1/admin/tours/"+tour+"/publish", `{}`, token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	w = request(a, "GET", "/api/v1/tours?budget_max=1700000", "", "", "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	var list struct {
		Data []any `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list.Data) != 1 {
		t.Fatal("missing filtered tour")
	}
	w = request(a, "GET", "/api/v1/tours?budget_max=100&depart_from="+start.Format("2006-01-02"), "", "", "", nil)
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list.Data) != 0 {
		t.Fatal("wrong departure price filter")
	}
	w = request(a, "POST", "/api/v1/admin/tours/"+tour+"/departures/bulk", `{"departures":[`+body(start.AddDate(0, 0, 1), "1000000")+`,`+body(start, "1000000")+`]}`, token, "", nil)
	if w.Code != 409 {
		t.Fatal(w.Body)
	}
	var count int
	a.DB.QueryRow(context.Background(), "SELECT count(*) FROM departures WHERE tour_id=$1", tour).Scan(&count)
	if count != 1 {
		t.Fatal("bulk partially committed")
	}
	w = request(a, "PATCH", "/api/v1/admin/departures/"+dep, `{"capacity":8,"expected_price_version":99}`, token, "", nil)
	if w.Code != 409 {
		t.Fatal("stale version accepted")
	}
	w = request(a, "POST", "/api/v1/admin/tours/"+tour+"/clone", `{"slug":"copy","code":"COPY","title":"Bản sao"}`, token, "", nil)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	var state string
	a.DB.QueryRow(context.Background(), "SELECT status FROM tours WHERE slug='copy'").Scan(&state)
	if state != "draft" {
		t.Fatal("clone published")
	}
}
func TestContentSanitizationAndPrivateMedia(t *testing.T) {
	a, token, _ := catalogFixture(t)
	w := request(a, "POST", "/api/v1/admin/pages", `{"slug":"test-page","title":"Trang mẫu","body":"<p>Xin chào</p><script>alert(1)</script>","status":"published"}`, token, "", nil)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	w = request(a, "GET", "/api/v1/pages/test-page", "", "", "", nil)
	if w.Code != 200 || bytes.Contains(w.Body.Bytes(), []byte("<script>")) {
		t.Fatal(w.Body)
	}
	w = request(a, "POST", "/api/v1/admin/pages", `{"slug":"draft-page","title":"Nháp"}`, token, "", nil)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	w = request(a, "GET", "/api/v1/pages/draft-page", "", "", "", nil)
	if w.Code != 404 {
		t.Fatal("draft leaked")
	}
}

func TestMediaPrivacyMagicBytesAndPosterRevision(t *testing.T) {
	a, token, tour := catalogFixture(t)
	upload := func(name, visibility string, content []byte) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, _ := writer.CreateFormFile("file", name)
		part.Write(content)
		writer.Close()
		req := httptest.NewRequest("POST", "/api/v1/admin/media?visibility="+visibility, &buf)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		out := httptest.NewRecorder()
		a.Router().ServeHTTP(out, req)
		return out
	}
	if w := upload("fake.png", "public", []byte("<script>bad</script>")); w.Code != 415 {
		t.Fatal("fake image accepted")
	}
	var img bytes.Buffer
	png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 8, 8)))
	w := upload("private.png", "private", img.Bytes())
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	id := data(t, w)["id"].(string)
	if w = request(a, "GET", "/api/v1/media/"+id, "", "", "", nil); w.Code != 404 {
		t.Fatal("private file exposed")
	}
	w = request(a, "POST", "/api/v1/admin/media/"+id+"/download-url", `{}`, token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	link := data(t, w)["url"].(string)
	if w = request(a, "GET", link, "", "", "", nil); w.Code != 200 {
		t.Fatal("signed URL failed")
	}
	if w = request(a, "GET", link+"bad", "", "", "", nil); w.Code != 404 {
		t.Fatal("signature accepted")
	}
	w = upload("poster.png", "public", img.Bytes())
	id = data(t, w)["id"].(string)
	w = request(a, "PATCH", "/api/v1/admin/tours/"+tour, fmt.Sprintf(`{"poster_media_id":%q}`, id), token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	revision := int(data(t, w)["poster_revision"].(float64))
	w = request(a, "POST", "/api/v1/admin/tours/"+tour+"/poster-verification", fmt.Sprintf(`{"expected_poster_revision":%d,"confirmed":true}`, revision), token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	request(a, "PATCH", "/api/v1/admin/tours/"+tour, `{"summary":"Đổi nội dung"}`, token, "", nil)
	w = request(a, "POST", "/api/v1/admin/tours/"+tour+"/poster-verification", fmt.Sprintf(`{"expected_poster_revision":%d,"confirmed":true}`, revision), token, "", nil)
	if w.Code != 409 {
		t.Fatal("old poster revision accepted")
	}
}

func TestDemoSeedIdempotentPreservesAdminChanges(t *testing.T) {
	a := testApp(t)
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOCAL_STORAGE_PATH", t.TempDir())
	ctx := context.Background()
	store := catalog.NewMediaStore(a.Config.CSRFSecret)
	if e := catalog.Seed(ctx, a.DB, store, demoassets.Files); e != nil {
		t.Fatal(e)
	}
	a.DB.Exec(ctx, "UPDATE tours SET title='Đã sửa trong Admin' WHERE slug='ninh-chu-2-ngay-1-dem'")
	if e := catalog.Seed(ctx, a.DB, store, demoassets.Files); e != nil {
		t.Fatal(e)
	}
	var count int
	var title string
	a.DB.QueryRow(ctx, "SELECT count(*) FROM tours").Scan(&count)
	a.DB.QueryRow(ctx, "SELECT title FROM tours WHERE slug='ninh-chu-2-ngay-1-dem'").Scan(&title)
	if count != 4 || title != "Đã sửa trong Admin" {
		t.Fatal("seed overwrote data or duplicated tours")
	}
}

func TestRelatedPostsHideDraftsAndKeepTags(t *testing.T) {
	a, token, _ := catalogFixture(t)
	create := func(path, body string) string {
		w := request(a, "POST", "/api/v1/admin/"+path, body, token, "", nil)
		if w.Code != 201 {
			t.Fatal(w.Body)
		}
		return data(t, w)["id"].(string)
	}
	tag := create("tags", `{"slug":"bien","name":"Biển"}`)
	for i, state := range []string{"published", "published", "draft"} {
		create("posts", fmt.Sprintf(`{"slug":"cam-nang-%d","title":"Cẩm nang %d","body":"Bài viết demo","tag_ids":[%q],"status":%q}`, i, i, tag, state))
	}
	w := request(a, "GET", "/api/v1/posts/cam-nang-0", "", "", "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	d := data(t, w)
	if len(d["related_posts"].([]any)) != 1 || len(d["tags"].([]any)) != 1 {
		t.Fatal("related draft or tags wrong", w.Body)
	}
	w = request(a, "GET", "/api/v1/posts?tag=bien", "", "", "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	var envelope struct {
		Data []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &envelope)
	if len(envelope.Data) != 2 {
		t.Fatal("tag leaked draft")
	}
}

func TestContentCoverRequiresPublicReadyImage(t *testing.T) {
	a, token, tour := catalogFixture(t)
	ctx := context.Background()
	for i, asset := range []struct{ kind, visibility, mime string }{{"image", "private", "image/png"}, {"document", "public", "application/pdf"}} {
		var id string
		e := a.DB.QueryRow(ctx, "INSERT INTO media_assets(object_key,original_name,mime_type,size_bytes,kind,visibility,status) VALUES($1,'cover',$2,12,$3,$4,'ready') RETURNING id", fmt.Sprintf("cover-%d", i), asset.mime, asset.kind, asset.visibility).Scan(&id)
		if e != nil {
			t.Fatal(e)
		}
		w := request(a, "POST", "/api/v1/admin/posts", fmt.Sprintf(`{"slug":"bad-%d","title":"Nội dung demo","cover_media_id":%q}`, i, id), token, "", nil)
		if w.Code != 422 {
			t.Fatal("nonpublic or nonimage cover accepted", w.Body)
		}
	}
	var cover string
	a.DB.QueryRow(ctx, "SELECT cover_media_id FROM tours WHERE id=$1", tour).Scan(&cover)
	w := request(a, "POST", "/api/v1/admin/posts", fmt.Sprintf(`{"slug":"valid-cover","title":"Nội dung demo","cover_media_id":%q}`, cover), token, "", nil)
	if w.Code != 201 {
		t.Fatal("public image cover rejected", w.Body)
	}
}
