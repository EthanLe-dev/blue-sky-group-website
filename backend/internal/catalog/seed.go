package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"image"
	"io/fs"
	"os"
	"time"
)

func DemoID(name string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("blue-sky-demo:"+name)).String()
}
func Seed(ctx context.Context, db *pgxpool.Pool, store *MediaStore, files fs.FS) error {
	if os.Getenv("APP_ENV") != "local" && os.Getenv("APP_ENV") != "demo" && os.Getenv("APP_ENV") != "test" {
		return errors.New("demo seed forbidden")
	}
	var staff string
	if e := db.QueryRow(ctx, "SELECT u.id FROM staff_users u JOIN staff_roles sr ON sr.staff_id=u.id JOIN roles r ON r.id=sr.role_id WHERE r.code='admin' ORDER BY u.created_at LIMIT 1").Scan(&staff); e != nil {
		return errors.New("bootstrap an admin before seeding")
	}
	for _, asset := range []struct{ name, alt, collection, source string }{{"ninh-chu.jpg", "Bãi biển Ninh Chữ", "destination", "Phương Huy — Wikimedia Commons, Public domain"}, {"vinh-hy.jpg", "Vịnh Vĩnh Hy", "destination", "Phương Huy — Wikimedia Commons, Public domain"}, {"ninh-chu-poster.png", "Poster Ninh Chữ 2 ngày 1 đêm", "poster", "Người dùng cung cấp"}} {
		id := DemoID(asset.name)
		var exists bool
		db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1)", id).Scan(&exists)
		if exists {
			continue
		}
		data, e := fs.ReadFile(files, asset.name)
		if e != nil {
			return e
		}
		cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
		if e != nil {
			return e
		}
		mime, kind, ext, _ := detectMedia(data[:min(512, len(data))])
		f, e := os.CreateTemp("", "blue-sky-seed-*")
		if e != nil {
			return e
		}
		f.Write(data)
		f.Seek(0, 0)
		key := id + ext
		e = store.put(ctx, key, "public", mime, f, int64(len(data)))
		f.Close()
		os.Remove(f.Name())
		if e != nil && !os.IsExist(e) {
			return e
		}
		_, e = db.Exec(ctx, "INSERT INTO media_assets(id,object_key,original_name,mime_type,size_bytes,width,height,kind,visibility,status,alt_text,collection,source_credit,uploaded_by_staff_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'public','ready',$9,$10,$11,$12) ON CONFLICT(id) DO NOTHING", id, key, asset.name, mime, len(data), cfg.Width, cfg.Height, kind, asset.alt, asset.collection, asset.source, staff)
		if e != nil {
			return e
		}
	}
	tx, e := db.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var settingsRaw []byte
	var defaults struct {
		Deposit int     `json:"deposit_bps"`
		Balance int     `json:"balance_due_days"`
		Policy  string  `json:"cancellation_policy"`
		Ages    []Price `json:"age_bands"`
	}
	if e = tx.QueryRow(ctx, "SELECT value FROM site_settings WHERE key='booking_defaults'").Scan(&settingsRaw); e != nil {
		return e
	}
	json.Unmarshal(settingsRaw, &defaults)
	_, e = tx.Exec(ctx, "INSERT INTO destinations(id,slug,name,description,latitude,longitude,cover_media_id,published) VALUES($1,'ninh-thuan','Ninh Thuận','Biển Ninh Chữ, Vĩnh Hy và văn hoá Chăm.',11.5901,108.9648,$2,true) ON CONFLICT DO NOTHING", DemoID("destination"), DemoID("ninh-chu.jpg"))
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, "INSERT INTO tour_categories(id,slug,name) VALUES($1,'bien-dao','Biển và khám phá') ON CONFLICT DO NOTHING", DemoID("category"))
	if e != nil {
		return e
	}
	tours := []struct {
		slug, title, code, cover, audience string
		price                              int64
		poster                             bool
	}{{"ninh-chu-2-ngay-1-dem", "Ninh Chữ · Biển và văn hoá Chăm", "BSG-NC02", "ninh-chu.jpg", "general", 1650000, true}, {"vinh-hy-hang-rai", "Vĩnh Hy · Hang Rái cuối tuần", "BSG-VH02", "vinh-hy.jpg", "general", 1850000, false}, {"ninh-chu-hoc-duong", "Ninh Chữ · Ngoại khoá học đường", "BSG-SC02", "ninh-chu.jpg", "school", 1450000, false}, {"vinh-hy-doanh-nghiep", "Vĩnh Hy · Chuyến đi cùng đồng đội", "BSG-CO02", "vinh-hy.jpg", "company", 1950000, false}}
	for _, item := range tours {
		id := DemoID(item.slug)
		var existed bool
		tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tours WHERE id=$1 OR slug=$2)", id, item.slug).Scan(&existed)
		if existed {
			continue
		}
		var poster *string
		if item.poster {
			v := DemoID("ninh-chu-poster.png")
			poster = &v
		}
		_, e = tx.Exec(ctx, `INSERT INTO tours(id,slug,code,title,summary,description,region,audience,duration_days,duration_nights,origin_name,transport,difficulty,hotel_summary,included,excluded,cancellation_policy,preparation_guide,status,featured,cover_media_id,poster_media_id,poster_verified_at,poster_verified_by,seo_title,seo_description,is_demo,published_at) VALUES($1,$2,$3,$4::text,'Hai ngày trải nghiệm biển và điểm đến văn hoá, khởi hành từ TP. Hồ Chí Minh.','<p>Lịch trình demo để bạn trải nghiệm website. Giá và các dịch vụ giả định có thể điều chỉnh trong quản trị.</p>','domestic',$5,2,1,'TP. Hồ Chí Minh','Xe du lịch','easy','Khách sạn theo chương trình demo','Xe, phòng nghỉ và các bữa ăn theo lịch trình demo.','Chi phí cá nhân và dịch vụ ngoài chương trình.',$6,'Mang theo giấy tờ cá nhân, đồ chống nắng và trang phục phù hợp.','published',true,$7,$8,CASE WHEN $8::uuid IS NOT NULL THEN now() END,CASE WHEN $8::uuid IS NOT NULL THEN $9::uuid END,$4::text,'Lịch trình hai ngày, giá minh bạch theo lịch khởi hành.',true,now())`, id, item.slug, item.code, item.title, item.audience, defaults.Policy, DemoID(item.cover), poster, staff)
		if e != nil {
			return e
		}
		for _, d := range []Day{{1, "Từ thành phố đến Ninh Chữ", "<p>Ăn sáng. Miếu Quan Âm sông Mao, Trung tâm trưng bày Văn hoá Chăm, làng gốm Bàu Trúc, Trường Sơn Cổ Tự. Nhận phòng và ăn tối.</p>", "Sáng, trưa, tối", "Khách sạn theo chương trình demo"}, {2, "Biển, vườn nho và trở về", "<p>Mỹ Hoà, Núi Chúa – Hang Rái, vườn nho, đồng muối Phương Cựu, Sắc Tứ Kim Sơn Tự, Linh đài Đức Mẹ Fatima Bình Chính. Ăn trưa và khởi hành về.</p>", "Sáng, trưa", ""}} {
			if _, e = tx.Exec(ctx, "INSERT INTO tour_days(tour_id,day_number,title,body,meals,accommodation) VALUES($1,$2,$3,$4,$5,$6)", id, d.Number, d.Title, d.Body, d.Meals, d.Accommodation); e != nil {
				return e
			}
		}
		for _, point := range []Stop{{1, 1, "Ninh Chữ", 11.599, 109.022, "Điểm đến demo"}, {2, 1, "Vĩnh Hy", 11.7255, 109.1966, "Vị trí tham khảo, xác minh trước chuyến đi"}} {
			if _, e = tx.Exec(ctx, "INSERT INTO tour_stops(tour_id,day_number,sort_order,name,latitude,longitude,description) VALUES($1,$2,$3,$4,$5,$6,$7)", id, point.Day, point.Order, point.Name, point.Latitude, point.Longitude, point.Description); e != nil {
				return e
			}
		}
		if _, e = tx.Exec(ctx, "INSERT INTO tour_destinations(tour_id,destination_id) VALUES($1,$2)", id, DemoID("destination")); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "INSERT INTO tour_category_links(tour_id,category_id) VALUES($1,$2)", id, DemoID("category")); e != nil {
			return e
		}
		for i, name := range []string{item.cover, "vinh-hy.jpg"} {
			if _, e = tx.Exec(ctx, "INSERT INTO tour_media(tour_id,media_id,sort_order) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", id, DemoID(name), i); e != nil {
				return e
			}
		}
		for i := 0; i < 3; i++ {
			depID := DemoID(fmt.Sprintf("%s-departure-%d", item.slug, i))
			start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 30+i*7).Add(22 * time.Hour)
			_, e = tx.Exec(ctx, "INSERT INTO departures(id,tour_id,starts_at,ends_at,sales_close_at,capacity,status,deposit_bps,balance_due_at) VALUES($1,$2,$3,$4,$5,30,'open',$6,$7)", depID, id, start, start.AddDate(0, 0, 1).Add(12*time.Hour), start.Add(-24*time.Hour), defaults.Deposit, start.AddDate(0, 0, -defaults.Balance))
			if e != nil {
				return e
			}
			prices := []Price{{Type: "infant", Amount: 0, Seats: 0, Min: 0, Max: intPointer(2)}, {Type: "child", Amount: Money(item.price * 7 / 10), Seats: 1, Min: 2, Max: intPointer(12)}, {Type: "adult", Amount: Money(item.price), Seats: 1, Min: 12}}
			for _, p := range prices {
				for _, band := range defaults.Ages {
					if band.Type == p.Type {
						p.Min = band.Min
						p.Max = band.Max
						p.Seats = band.Seats
					}
				}
				if _, e = tx.Exec(ctx, "INSERT INTO departure_prices(departure_id,passenger_type,unit_price,seats_per_person,min_age,max_age_exclusive) VALUES($1,$2,$3,$4,$5,$6)", depID, p.Type, int64(p.Amount), p.Seats, p.Min, p.Max); e != nil {
					return e
				}
			}
		}
	}
	for _, page := range []struct{ slug, title, body string }{{"home", "Tour biển cho những ngày nghỉ", "Chọn lịch khởi hành, xem giá theo từng hành khách và lên kế hoạch cho chuyến đi của bạn."}, {"about", "Về Blue Sky Group", "<p>Blue Sky Group — Travel, Event, Media. Website demo giới thiệu tour và nhận yêu cầu đặt chỗ. Thông tin doanh nghiệp được cập nhật trong Admin.</p>"}, {"contact", "Liên hệ Blue Sky Group", "<p>Hotline theo poster tham chiếu: 0989 876 847. Địa chỉ theo poster: 183/61 Nguyễn Văn Khối, phường Thông Tây Hội, TP. Hồ Chí Minh. Vui lòng xác minh trước khi sử dụng thông tin chính thức.</p>"}, {"schools", "Chuyến đi dành cho trường học", "<p>Chương trình ngoại khoá kết hợp biển và văn hoá Chăm. Số lượng, lịch trình, nhân sự hỗ trợ và yêu cầu an toàn được trao đổi trước khi chốt.</p>"}, {"companies", "Chuyến đi dành cho doanh nghiệp", "<p>Lịch trình cho tập thể, hoạt động nhóm và thời gian nghỉ dưỡng. Các dịch vụ được điều chỉnh theo quy mô đoàn.</p>"}, {"privacy", "Chính sách bảo mật", "<p>Dữ liệu liên hệ chỉ dùng để xử lý yêu cầu đặt tour. Đây là nội dung demo, cần công ty duyệt trước vận hành chính thức.</p>"}, {"terms", "Điều khoản đặt tour", "<p>Đặt tour dựa trên lịch khởi hành và giá đã xác nhận. Dữ liệu demo không tạo giao dịch thanh toán thật.</p>"}, {"payment", "Chính sách thanh toán", "<p>Chọn thanh toán đầy đủ hoặc đặt cọc theo tour. Giai đoạn demo không chuyển tiền thật.</p>"}, {"cancellation", "Chính sách huỷ và đổi", "<p>Liên hệ nhân viên để được xem xét theo chính sách lưu cùng đơn. Chưa có chức năng tự huỷ trực tuyến.</p>"}, {"preparation", "Chuẩn bị cho chuyến đi", "<p>Kiểm tra giấy tờ, hành lý, thông tin thời tiết, giờ đón và nhu cầu sức khoẻ với nhân viên trước ngày đi.</p>"}} {
		if _, e = tx.Exec(ctx, "INSERT INTO pages(slug,title,body,status,published_at) VALUES($1,$2,$3,'published',now()) ON CONFLICT(slug) DO NOTHING", page.slug, page.title, page.body); e != nil {
			return e
		}
	}
	for _, f := range []struct{ q, a string }{{"Tôi có cần tài khoản để đặt tour không?", "Bạn có thể đặt tour bằng thông tin liên hệ, không cần tạo tài khoản."}, {"Giá trên poster có áp dụng mọi ngày không?", "Giá chính thức hiển thị theo lịch khởi hành và từng đối tượng hành khách. Vui lòng kiểm tra trước khi đặt."}, {"Đặt cọc được xác nhận như thế nào?", "Số tiền cọc và hạn thanh toán phần còn lại hiển thị khi chọn lịch. Giai đoạn demo không chuyển tiền thật."}} {
		if _, e = tx.Exec(ctx, "INSERT INTO faqs(id,topic,question,answer,published) VALUES($1,'booking',$2,$3,true) ON CONFLICT(id) DO NOTHING", DemoID(f.q), f.q, f.a); e != nil {
			return e
		}
	}
	cat := DemoID("post-category")
	if _, e = tx.Exec(ctx, "INSERT INTO post_categories(id,slug,name) VALUES($1,'cam-nang','Cẩm nang du lịch') ON CONFLICT DO NOTHING", cat); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO posts(id,slug,title,excerpt,body,category_id,cover_media_id,author_id,status,published_at) VALUES($1,'chuan-bi-di-ninh-chu','Chuẩn bị cho hai ngày ở Ninh Chữ','Giấy tờ, hành lý và những điều nên kiểm tra trước ngày đi.','<p>Trao đổi giờ đón, kiểm tra giấy tờ và chuẩn bị trang phục phù hợp với hoạt động biển. Thông báo nhu cầu ăn uống và sức khoẻ cho nhân viên trước chuyến đi.</p>',$2,$3,$4,'published',now()) ON CONFLICT DO NOTHING", DemoID("post"), cat, DemoID("ninh-chu.jpg"), staff); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func intPointer(v int) *int { return &v }
