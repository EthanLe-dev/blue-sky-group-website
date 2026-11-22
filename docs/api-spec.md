# Hợp đồng API — Blue Sky Group

Phiên bản 0.1 · 22/11/2026 · REST `/api/v1`, UTF-8 JSON, MVP. Đây là contract thiết kế cho frontend/admin/backend, chưa có endpoint được triển khai. [Database](database-schema.md) định nghĩa persistence; [PLAN](../agent/PLAN.md) ghi giả định nghiệp vụ. M1 sẽ chuyển thành OpenAPI và test schema, không tự suy ra contract từ ORM.

## 1. Quy ước chung

- Public local: `http://localhost:3000`; Admin: `http://localhost:5173`; API trực tiếp: `http://localhost:8080/api/v1`. Demo URL lấy env, không hardcode hostname.
- Path dùng kebab-case, JSON dùng `snake_case`, ID UUID; slug lowercase không dấu. Thời gian RFC3339 UTC, ngày `YYYY-MM-DD`, timezone hiển thị Việt Nam. Tiền là **chuỗi nguyên đồng VND**, ví dụ `"3500000"`, không dấu phân cách; frontend không dùng float để tính tổng chính thức.
- Body JSON giới hạn 256 KiB (upload riêng); unknown fields trả 422. Cấm client gửi status/counters/paid_amount/role ngoài các action được định nghĩa. Chuỗi trim, field length theo schema; mô tả rich text tối đa 100 KiB, note/message tối đa 2.000 ký tự, q tối đa 100.
- Thành công: `{ "success": true, "data": ..., "error": null, "message": null }`. List có thêm `meta: {"page":1,"page_size":20,"total":45,"total_pages":3}`. Mặc định page=1, page_size=20, max=100; sort chỉ allowlist và tie-break id để ổn định.
- Lỗi: `{ "success": false, "data": null, "error": {"code":"...","fields":{}}, "message":"Thông báo tiếng Việt" }`. `fields` map đường dẫn field sang mảng thông báo; không stack trace/SQL/PII.
- `X-Request-ID` server sinh nếu không có ID hợp lệ và trả mọi response. HTTP 200 đọc/action/update, 201 tạo, 202 nhận xử lý async nếu endpoint ghi rõ, 400 JSON sai, 401 chưa xác thực, 403 thiếu quyền/CSRF, 404 không có/không được thấy object, 409 conflict, 413 quá lớn, 415 MIME sai, 422 validation, 429 rate limit + `Retry-After`, 503 phụ thuộc chưa sẵn sàng.
- DELETE CRUD nội dung trả 200 envelope `{id,deleted:true}` hoặc `{id,archived:true}`; không 204 với JSON. File stream và health là ngoại lệ envelope được ghi riêng.
- Public catalog có thể cache; booking/quote/payment/auth/lookup/admin/notification đặt `Cache-Control: no-store`. Frontend chỉ cache catalog, không cache dữ liệu khách trong service worker/localStorage.
- `Idempotency-Key` random UUID bắt buộc trên create booking/payment intent/confirm/refund/admin booking và public lead forms. Scope principal + method + route, giữ tối thiểu 24h. Cùng key/payload trả lại status/body gốc; payload khác 409 `IDEMPOTENCY_CONFLICT`; đang xử lý 409 `REQUEST_IN_PROGRESS`. Không lưu raw token trong response replay. Lead scope từ guest session. Network retry giữ nguyên key. Hash payload nghiệp vụ chuẩn hoá, loại Turnstile token khỏi hash; sau kiểm tra principal/CSRF, replay kết quả đã commit không xác minh lại token Turnstile một lần. Request mới vẫn phải verify Turnstile trước ghi nghiệp vụ.

```json
{
  "success": false,
  "data": null,
  "error": {"code":"VALIDATION_ERROR","fields":{"passengers.0.date_of_birth":["Ngày sinh không phù hợp nhóm tuổi đã chọn."]}},
  "message": "Vui lòng kiểm tra thông tin hành khách."
}
```

## 2. Xác thực, quyền và chống spam

### Session và deployment

- `POST /guest-sessions` tạo cookie random HttpOnly và trả `{csrf_token,expires_at}`; TTL 24h. Nếu cookie hiện tại hợp lệ thì tái sử dụng session và cấp lại CSRF. Cookie chỉ lưu raw secret, DB chỉ hash. Guest cookie gắn các quote/booking do chính session tạo.
- Admin login trả access JWT TTL 15 phút trong response, frontend giữ **memory**; refresh token HttpOnly TTL 7 ngày, rotate mỗi refresh, phát hiện reuse revoke family. Logout/disable account revoke session; middleware kiểm tra session active để access JWT cũ không dùng tiếp.
- Mọi mutation dùng cookie (guest, refresh, logout) kiểm tra Origin allowlist + `X-CSRF-Token` gắn session. Login kiểm tra Origin và rate limit. Bearer-only admin action không tự dựa cookie để cấp quyền.
- Cookie production/demo `Secure; HttpOnly; SameSite=Lax`, host-only, Path `/`; local HTTP cho Secure=false. Không đặt cookie Domain dùng chung toàn hệ thống.
- Public Next.js route proxy `/api/v1/*` tới API; admin Vercel rewrite `/api/v1/*` tương tự, dev Vite proxy. Proxy phải chuyển Cookie/Set-Cookie, Authorization, CSRF và request ID; không log body nhạy cảm. Backend dùng trust-proxy allowlist, không tin forwarded headers từ mọi nguồn. Cách này giữ cookie same-origin trên mỗi frontend, tránh phụ thuộc third-party cookie giữa `vercel.app` và `onrender.com`.
- Nếu gọi API trực tiếp từ origin khác phải cấu hình cụ thể `credentials: include`, CORS credentials + allowlist và chiến lược cookie thích hợp; không wildcard. Proxy là baseline MVP cần kiểm thử M1/M7.
- Public form/lookup/reset password/login dùng Turnstile khi công khai; server verify token/action/hostname. Local dùng adapter mock. Rate limit baseline: login 5/phút/IP+email; lookup 5/15 phút/IP+mã; lead/checkout 10/10 phút/session+IP; session creation 30/10 phút/IP. Backend persistent counter; rate-limit key hash.

### RBAC (kiểm tra ở server)

| Vai trò | Quyền mặc định |
| --- | --- |
| `admin` | Tất cả quyền; quản lý staff/role/settings/audit |
| `product_manager` | `catalog.write`, `content.write`, `media.write`; đọc catalog/content, không đọc PII booking |
| `sales` | `booking.read`, `booking.write`, `customer.write`, `lead.write`, `notification.read`, `media.private.write`; xử lý đơn/lead, không tự refund hay thu tiền thật |
| `accountant` | `booking.read`, `payment.write`, `report.read`, `report.export`; chỉ PII cần đối soát; không sửa giá/catalog |
| `executive` | `report.read`; chỉ số tổng hợp, không export danh sách khách |

Endpoint admin đọc tài nguyên yêu cầu quyền write tương ứng nếu không có quyền read riêng; đọc booking dùng booking.read. Export có PII cần **cả** report.export và booking.read; export tour cần catalog.write, không report.export. Private lead attachment cần lead.write; media.write không tự cấp truy cập file khách. Các route tổng quát phải lọc field theo quyền.

## 3. Mô hình response dùng chung

| DTO | Fields |
| --- | --- |
| `Money` | Chuỗi regex `^[0-9]+$`, currency luôn `VND`; tổng tối đa 1.000.000.000.000 VND/đơn, server kiểm tra overflow |
| `Media` | `id, kind, url, alt_text, caption, width, height, is_demo?`; url public chỉ với asset public/ready |
| `TourCard` | `id,slug,title,region,audience,duration_days,duration_nights,origin_name,transport,difficulty,cover,price_from,currency,next_departure_at,rating,review_count`; price_from=null nếu không có departure open còn bán |
| `TourDetail` | TourCard + `summary,description,destinations[],categories[],gallery[],days[{day_number,title,body,meals,accommodation}],stops[{day_number,name,latitude,longitude}],included,excluded,cancellation_policy,preparation_guide,age_policy,hotel_summary,faqs[],seo{title,description},related_tours[]` |
| `Departure` | `id,tour_id,starts_at,ends_at,sales_close_at,status,capacity,remaining_seats,availability(available/low/sold_out/closed),prices[{passenger_type,unit_price,seats_per_person,min_age,max_age_exclusive}],surcharges[{code,label,amount,unit,required}],deposit_bps,balance_due_at,price_version,server_time`; low = remaining 1..5, closed ưu tiên status/cutoff |
| `PriceBreakdown` | `items[{kind,code,label,quantity,unit_amount,line_amount}],subtotal_amount,surcharge_amount,discount_amount,total_amount,deposit_amount,balance_after_deposit,currency` |
| `Booking` | `id,code,status,payment_status,tour_snapshot,departure{starts_at,ends_at},counts{adult,child,infant},seat_count,price_breakdown,paid_amount,refunded_amount,balance_amount,payment_option,balance_due_at,expires_at,server_time,version,contact{name,email,phone},passengers[],special_requests` |
| `BookingSummary` | `code,status,payment_status,tour_title,starts_at,ends_at,traveller_count,total_amount,paid_amount,balance_amount,currency,contact_email_masked,contact_phone_masked`; không ID nội bộ/passenger/notes/special_requests |
| `Payment` | `id,booking_id,provider,purpose,amount,currency,status,is_mock,expires_at,paid_at,demo_qr_url?`; QR là hình minh hoạ không có tài khoản nhận tiền thật |
| `NotificationPreview` | `id,channel,subject,body,status,is_mock,created_at`; recipient che trên list, nội dung chỉ creator hoặc notification.read |
| `Post` | `id,slug,title,excerpt,body,category,tags[],cover,published_at,seo`; list bỏ body |
| `LeadReceipt` | `code,status,received_at,response_sla_message,is_demo`; không public tracking link trong MVP |

Rating chỉ tính approved; nếu là dữ liệu seed, frontend hiển thị nhãn demo. `remaining_seats` là snapshot tại `server_time`, không bảo đảm thành công booking tiếp theo. Enum status theo database; không dịch giá trị enum trên wire.

## 4. API công khai: catalog và nội dung

Không cần auth, chỉ published/approved/ready, không trả internal note. Empty list 200, slug không tồn tại 404.

| Method/path | Query/body | `data` |
| --- | --- | --- |
| GET `/site-settings` | — | `hotline,email,zalo_url,address,office_coordinates,social_links,company_documents[],response_sla` allowlist, không secret |
| GET `/home` | — | `banners[],featured_tours[],destinations[],testimonials[],latest_posts[],service_commitments[]` từ CMS; promotions chỉ nội dung marketing, không coupon logic |
| GET `/tours` | `q,destination(slug),category(slug),region,audience,depart_from,depart_to,budget_min,budget_max,duration_min,duration_max,travellers,transport,difficulty,min_rating,sort,page,page_size` | TourCard[] + meta |
| GET `/tours/suggestions` | `q` tối thiểu 2, `limit` mặc định 5/max10 | `{tours:[{id,slug,title}],destinations:[{slug,name}]}` |
| GET `/tours/{slug}` | — | TourDetail |
| GET `/tours/{slug}/departures` | `from,to,page,page_size` | Departure[] + meta |
| GET `/tours/{slug}/reviews` | `page,page_size` | `{id,display_name,organization_name,rating,body,is_demo,created_at}`[] |
| GET `/tours/{slug}/itinerary.pdf` | — | Stream PDF công khai chỉ published tour |
| GET `/destinations` | `q,page,page_size` | `{id,slug,name,description,latitude,longitude,cover,seo}`[] |
| GET `/destinations/{slug}` | — | Destination + `featured_tours[]` |
| GET `/tour-categories` | — | `{id,slug,name,description}`[] |
| GET `/posts` | `q,category,tag,page,page_size` | Post[] + meta |
| GET `/posts/{slug}` | — | Post + related_posts[] |
| GET `/post-categories`, `/tags` | — | `{id,slug,name}`[] |
| GET `/pages/{slug}` | — | `{slug,title,body,seo,updated_at}` |
| GET `/faqs` | `topic,tour_id?` | `{id,topic,question,answer}`[] |
| GET `/gallery` | `kind,collection,page,page_size` | Media[] + meta |
| GET `/testimonials` | `group_type=school|company,page,page_size` | Approved review[] |

Tour filters kết hợp AND; ngày đi inclusive theo ngày VN; budget là giá adult một người của **cùng departure** đáp ứng date/availability, không trộn giá rẻ ngày khác. `travellers` là chỗ yêu cầu (adult+child theo baseline), dương ≤20. Sort `recommended` (featured trước, published_at giảm dần), `price_asc`, `price_desc`, `duration_asc`, `departure_asc`, `rating_desc`. Budget thấp hơn 0 hoặc min>max =>422; tour không có departure hợp lệ không vào kết quả khi lọc giá/ngày/chỗ. Search không phân biệt hoa/thường; hỗ trợ tiếng Việt có dấu, không cam kết bỏ dấu trước khi có index phù hợp.

Sitemap/robots/canonical/structured data do Next.js xuất từ dữ liệu published; pagination không đưa draft ra SEO.

## 5. Quote → booking → payment

### 5.1 Tính giá (guest session + CSRF)

`POST /booking-quotes` →201, body:

```json
{
  "departure_id":"7f831b00-3991-44c7-b2ba-01e849393c11",
  "counts":{"adult":2,"child":1,"infant":0},
  "surcharge_codes":[]
}
```

`data = {id,departure_id,price_version,counts,seat_count,price_breakdown,allowed_payment_options:["full","deposit"],balance_due_at,expires_at,server_time}`. Quote TTL 5 phút, **không giữ chỗ**. Mandatory surcharge server tự thêm, optional chỉ từ allowlist. Không coupon logic MVP; field coupon_code bị 422 `FEATURE_NOT_AVAILABLE`. Đã quá balance_due_at thì options chỉ full.

Ví dụ 2 adult×2.000.000 + 1 child×1.000.000 = `total_amount:"5000000"`, cọc 30% `deposit_amount:"1500000"`, `balance_after_deposit:"3500000"`; `discount_amount:"0"`. Làm tròn cọc lên một đồng, không làm tròn mỗi dòng phần trăm riêng.

### 5.2 Tạo đơn và hold (guest + CSRF + Turnstile + Idempotency-Key)

`POST /bookings` →201:

```json
{
  "quote_id":"e49cc1c6-ece8-4088-9aaf-76489f531d60",
  "contact":{"name":"Nguyễn Văn Mẫu","email":"demo@example.com","phone":"+84901234567"},
  "passengers":[
    {"passenger_type":"adult","full_name":"Nguyễn Văn Mẫu","date_of_birth":"1990-01-01"},
    {"passenger_type":"adult","full_name":"Trần Thị Mẫu","date_of_birth":"1992-05-10"},
    {"passenger_type":"child","full_name":"Nguyễn Bé Mẫu","date_of_birth":"2020-06-01"}
  ],
  "payment_option":"deposit",
  "special_requests":"",
  "consent":{"accepted":true,"policy_version":"booking-v1"},
  "turnstile_token":"token-from-widget"
}
```

Trả Booking. Contact name 2..120, email hợp lệ ≤254, phone chuẩn hoá E.164; cần cả email+phone cho demo notification. Passenger 1..20, counts/nhóm tuổi phải khớp quote. Policy version phải là phiên bản đang chấp nhận. Hold 20 phút theo DB. Hết quota →409 `SOLD_OUT`; quote hết hạn →409 `QUOTE_EXPIRED`; version giá khác →409 `PRICE_CHANGED`, frontend lấy quote mới và cho khách xem lại, không tự submit giá khác. Session khác quote →404.

`GET /bookings/{id}` →200 Booking, chỉ creator session. `GET /bookings/{id}/notifications` →200 NotificationPreview[] chỉ creator; không dùng lookup token. Public không có PATCH/cancel ở MVP. Không trả guest token trong URL/JSON/analytics.

### 5.3 Tạo và xác nhận thanh toán mock

`POST /bookings/{id}/payments` (creator + CSRF + Idempotency-Key) body `{ "purpose":"deposit" }` →201 Payment.

- purpose full chỉ cho pending với payment_option full; deposit cho pending/deposit; balance cho confirmed còn nợ. Server tính amount từ snapshot/ledger, client không gửi amount.
- Mỗi booking một pending intent; gọi với key khác khi đã có pending =>409 `PAYMENT_IN_PROGRESS`, lấy intent qua `GET /bookings/{id}/payments`.
- `expires_at` intent ban đầu không vượt hold expiry; balance intent TTL 20 phút, không gia hạn booking/balance_due_at. Khoản balance quá hạn vẫn cho trả nếu booking chưa cancelled/completed, hiển thị overdue; sales xử lý theo policy.
- `POST /payments/{id}/mock-confirm` body `{}` (creator + CSRF + Idempotency-Key) →200 `{payment,booking}`. Chỉ local/demo và PAYMENT_PROVIDER=mock, production không mount route (404).
- Pending booking hết hạn =>409 `HOLD_EXPIRED`; hết hạn payment intent =>409 `PAYMENT_EXPIRED`. Confirm thành công lặp lại trả kết quả đã ghi nhận, không nhân tiền/event/chỗ. Concurrent intent không bao giờ paid>total.
- Booking confirmed khi nhận đủ cọc yêu cầu, payment_status partially_paid; trả balance thì paid. Không frontend tự đổi trạng thái khi bấm nút, phải lấy kết quả server.
- Notification chạy qua outbox sau commit; UI có thể poll mỗi 2 giây tối đa 30 giây khi tab active, sau đó hiển thị “đang tạo bản xem trước”, không coi thiếu preview là thanh toán thất bại.

### 5.4 Tra cứu đơn

`POST /booking-lookups` body `{code,contact,turnstile_token}` →200 `{booking:BookingSummary,lookup_token,expires_at}`; contact là email hoặc phone chuẩn hoá. Không đặt contact vào query URL. Không tìm thấy/sai contact cùng 404 `BOOKING_NOT_FOUND`, không tiết lộ field nào đúng.

Lookup token random TTL 15 phút, giữ memory; `GET /booking-lookups/current` với `Authorization: Bearer <lookup_token>` →BookingSummary. Token **chỉ** gọi route này, không xem hành khách, notification, không confirm payment. Nếu mất creator cookie thì liên hệ sales để thanh toán tiếp; OTP nâng quyền là giai đoạn sau.

## 6. Form khách đoàn và gọi lại

Guest session + CSRF + Turnstile + Idempotency-Key. 201 LeadReceipt, xác nhận đã nhận nhưng không nói đã gửi email thật.

| Endpoint | Body bắt buộc | Body tuỳ chọn |
| --- | --- | --- |
| POST `/group-requests` | `group_type(school/company),organization_name,contact_name,contact_email,contact_phone,participant_count(1..10000),destination_text,consent{accepted:true,policy_version},turnstile_token` | `preferred_start_date,preferred_end_date,budget_per_person,requirements` |
| POST `/callback-requests` | `phone,consent{accepted:true,policy_version},turnstile_token` | `name,preferred_time,tour_id,message` |

Date không quá khứ, end>=start; budget nếu có là Money. SLA thông báo lấy site_settings do công ty duyệt, không hardcode cam kết chưa xác nhận. Không public upload danh sách đoàn, GET trạng thái báo giá hoặc newsletter submit ở MVP.

## 7. Auth/admin và CRUD chuẩn

| Endpoint | Request | Response |
| --- | --- | --- |
| POST `/admin/auth/login` | `{email,password,turnstile_token}` | 200 `{access_token,token_type:"Bearer",expires_in:900,csrf_token,user:{id,email,full_name,roles,permissions}}` + refresh cookie |
| POST `/admin/auth/refresh` | `{}` + refresh cookie/CSRF | 200 access token mới + rotated cookie + csrf_token |
| POST `/admin/auth/logout` | `{}` + refresh cookie/CSRF | 200 `{logged_out:true}`, revoke session |
| GET `/admin/auth/me` | Bearer | User + permissions |
| POST `/admin/auth/forgot-password` | `{email,turnstile_token}` | 202 `{accepted:true}` dù email có/không tồn tại; mock reset delivery chỉ admin khác thấy, không public trả token |
| POST `/admin/auth/reset-password` | `{token,password}` | 200 `{reset:true}`; token một lần, revoke sessions; password 12..128 |

Admin endpoint khác cần Bearer. List có page/page_size/q; lọc cụ thể theo tài nguyên, unknown sort/filter trả 422. Create trả201 DTO; GET/PATCH/action trả200; PATCH chỉ field allowlist gửi lên, null chỉ cho nullable field. Bookings/departures dùng optimistic version: body `expected_version` cho booking; `expected_price_version` khi sửa giá departure, stale =>409 `VERSION_CONFLICT`. Inventory vẫn khoá DB dù version khớp.

### 7.1 Catalog và nội dung

| Resource dưới `/admin` | Operations | Write DTO / điều kiện |
| --- | --- | --- |
| `/tours` | GET list, POST; GET/PATCH/DELETE `/{id}` | `title,slug,code,summary,description,region,audience,duration_days,duration_nights,origin_name,transport,difficulty,hotel_summary?,included,excluded,cancellation_policy,preparation_guide?,age_policy,featured,cover_media_id?,destination_ids[],category_ids[],days[],stops[],media_ids[],seo_title?,seo_description?`; POST draft, PATCH partial; status chỉ action |
| `/tours/{id}/publish`, `/archive` | POST `{}` | Publish yêu cầu đầy đủ itinerary/policy/media và ít nhất một departure có đủ giá; archive không xoá booking |
| `/tours/{id}/clone` | POST `{slug,code,title}` | Draft mới; copy nội dung/media link, không copy departure/booking/review |
| `/tours/{id}/departures` | GET, POST | `starts_at,ends_at,sales_close_at,capacity,deposit_bps,balance_due_at,prices[],surcharges[]`; draft mặc định |
| `/tours/{id}/departures/bulk` | POST `{departures:[DepartureWrite]}` | 1..100, atomic toàn bộ, trùng ngày 409, không thành công một phần |
| `/departures/{id}` | GET,PATCH | Partial DepartureWrite + expected_price_version; tăng price_version khi sửa giá/policy; đổi thời gian có booking active =>409 |
| `/departures/{id}/sales-status` | POST `{status:"open"|"closed"}` | Check đủ giá, cutoff; không huỷ chuyến bằng endpoint này |
| `/destinations`, `/tour-categories` | CRUD | Fields schema, không id/timestamps; không xoá tài nguyên đang referenced |
| `/posts`, `/pages`, `/banners`, `/faqs`, `/post-categories`, `/tags` | CRUD | Fields schema, FK dưới dạng `_id`/`_ids`; post `tag_ids[]`; status/published cho draft/publish, sanitize body; author từ staff hiện tại |
| `/reviews` | CRUD + POST `/{id}/moderation` | Create `tour_id?,group_type?,display_name,organization_name?,rating,body,media_id?,source,is_demo`; moderation `{status:approved|rejected,reason?}`, reviewer từ JWT |
| `/site-settings/{key}` | GET, PUT `{value}` | chỉ admin; key allowlist, shape theo settings DTO, không arbitrary HTML/secret |

`catalog.write` cho tours/departures/destinations/categories; `content.write` cho posts/pages/banners/faqs/reviews/tags. CRUD nghĩa GET collection, POST collection, GET/PATCH/DELETE `/{id}`; DELETE archive khi có history, conflict khi vẫn tham chiếu cần giữ. `days[]/stops[]/media_ids[]` PATCH thay toàn bộ collection trong một transaction; không thay từng phần ngầm. DTO list admin gồm id/status/timestamps và fields nội dung, không tự expose toàn DB.

### 7.2 Booking/payment/khách hàng

| Endpoint | Body/query | Quyền / trả về |
| --- | --- | --- |
| GET `/admin/bookings` | `status,payment_status,tour_id,departure_id,source,from,to,q,page,page_size` | booking.read; Booking list bỏ passengers/special_requests, contact theo quyền |
| GET `/admin/bookings/{id}` | — | booking.read; Booking + events/notes theo quyền sales/admin |
| POST `/admin/booking-quotes` | Cùng quote public | booking.write; quote gắn staff principal |
| POST `/admin/bookings` | Cùng booking public, bỏ Turnstile, thêm `source:phone|counter,customer_id?` | booking.write + Idempotency-Key; cùng inventory service, không bỏ hold/chỗ |
| PATCH `/admin/bookings/{id}` | `expected_version,contact?,passengers?,special_requests?,customer_id?` | booking.write; không đổi counts/departure/giá; thay passengers phải khớp snapshot/nhóm tuổi |
| POST `/admin/bookings/{id}/notes` | `{body}` | booking.write; note nội bộ |
| POST `/admin/bookings/{id}/cancel` | `{expected_version,reason}` | booking.write + Idempotency-Key; CancelResult `{booking,refund_required_amount}`, không tự refund |
| POST `/admin/bookings/{id}/complete` | `{expected_version}` | booking.write; sau ends_at, trả Booking |
| GET `/admin/bookings/{id}/payments` | — | booking.read; Payment[] + refunds[] |
| POST `/admin/bookings/{id}/payments` | `{purpose}` | payment.write + Idempotency-Key; cho đơn thủ công, trả Payment |
| POST `/admin/payments/{id}/mock-confirm` | `{reason}` | payment.write + Idempotency-Key; chỉ demo, trả payment+booking |
| POST `/admin/payments/{id}/refunds` | `{amount,reason}` | payment.write + Idempotency-Key; MVP refund giả lập succeeded, không quá paid còn lại; trả Refund `{id,payment_id,amount,status,is_mock,completed_at}` |
| GET `/admin/customers` | `q,page,page_size` | customer.write; Customer[] |
| POST `/admin/customers` | `{full_name,email?,phone?,company_name?,notes?}` | customer.write; Customer |
| GET/PATCH `/admin/customers/{id}` | PATCH partial CustomerWrite | customer.write; Customer |
| GET `/admin/customers/{id}/bookings` | `page,page_size` | customer.write + booking.read; lịch sử chỉ booking đã linked xác minh |
| GET `/admin/notifications` | `booking_id,channel,status,page,page_size` | notification.read; NotificationPreview[] |

Huỷ có phí: MVP nhân viên xem policy_snapshot và nhập số tiền refund có lý do, không tự tính mức phạt chưa được BA chốt. Refund không tự cancel booking; muốn trả chỗ phải action cancel trước. Chuyến cancelled thực hiện workflow huỷ từng đơn rồi đóng departure khi không còn active booking (MVP xử lý qua admin, không bulk update bỏ transaction).

### 7.3 Lead, quản trị người dùng và audit

| Endpoint | Request/response |
| --- | --- |
| GET `/admin/group-requests`, GET `/{id}` | lead.write; list filters status/assigned_to/from/to; detail gồm fields nhóm, notes, quote versions và attachments |
| PATCH `/admin/group-requests/{id}` | `{status?,assigned_to?,follow_up_due_at?}`; new→preparing→quoted→won/lost; lost có thể reopen preparing với reason; won→khác bị409; audit mọi chuyển |
| POST `/admin/group-requests/{id}/notes` | `{body}` →Note |
| POST `/admin/group-requests/{id}/quotes` | `{title,itinerary,inclusions,exclusions,total_amount,currency:"VND",valid_until}` →draft quote, version server tăng |
| PATCH `/admin/group-quotes/{id}` | Partial QuoteWrite, chỉ draft |
| POST `/admin/group-quotes/{id}/issue` | `{}` →immutable issued quote, cập nhật request quoted; không tự gửi email thật |
| POST `/admin/group-quotes/{id}/decision` | `{status:"accepted"|"rejected",reason}`; ghi nhận kết quả offline; accepted cập nhật request won |
| GET `/admin/group-quotes/{id}/document.pdf` | lead.write; private PDF |
| POST `/admin/group-requests/{id}/attachments` | `{media_id}` private ready →attachment |
| GET `/admin/callback-requests`, GET `/{id}`, PATCH `/{id}` | lead.write; filters status/assigned_to; PATCH `{status?,assigned_to?,follow_up_due_at?}`; new→contacted→closed |
| POST `/admin/callback-requests/{id}/notes` | lead.write; `{body}` |
| GET/POST `/admin/staff`, GET/PATCH `/admin/staff/{id}` | admin; create `{email,full_name,password,role_codes[]}`, patch `{full_name?,status?,role_codes[]}`; không trả hash/password; không disable/xoá quyền admin cuối cùng |
| GET `/admin/roles`, GET `/admin/permissions` | admin; danh sách quyền seed; không role editor tuỳ ý MVP |
| GET `/admin/audit-logs` | admin; filters actor_staff_id/entity_type/entity_id/from/to/page; chỉ redacted diff, không API xoá/sửa |

Assignment chỉ staff active có lead.write. Response_sla dựa settings; `follow_up_due_at` cho danh sách quá hạn, chưa có scheduler nhắn thật.

## 8. Upload, export và báo cáo

- `POST /admin/media` multipart field `file`, `visibility`, `alt_text?`, `collection?`, `source_credit?`. 201 MediaAdmin `{id,object_key,mime_type,size_bytes,kind,visibility,status,url?}`. Cho JPEG/PNG/WebP tối đa10MiB, MP4/WebM tối đa50MiB, PDF tối đa20MiB; kiểm tra magic bytes/size, server sinh key, private là mặc định. Catalog upload cần media.write; private attachment cần media.private.write + lead.write. Chặn SVG/HTML/executable/zip; frontend không tự cấp public visibility cho file khách.
- `GET /admin/media` list metadata theo quyền; `PATCH /admin/media/{id}` cho alt_text/caption/collection; `DELETE` archive chỉ khi không referenced. Private URL phải lấy qua `POST /admin/media/{id}/download-url` →200 `{url,expires_at}` TTL 5 phút, kiểm tra quyền đối tượng cha.
- Export stream headers: `Content-Type: application/pdf`, `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` hoặc `text/csv; charset=utf-8`; `Content-Disposition: attachment; filename*=UTF-8''...`; `Cache-Control:no-store` cho private. Lỗi trước stream dùng JSON error envelope.
- Export tối đa 10.000 rows/lần, vượt trả422 `EXPORT_TOO_LARGE` yêu cầu thu hẹp ngày; không truncate âm thầm. Timeout render trả503 `EXPORT_UNAVAILABLE`; file tạm xoá sau response. Excel/CSV escape formula bắt đầu =,+,-,@ và control prefix; PDF embed font Việt, escape user content, chặn fetch URL nội bộ khi render HTML.

| GET endpoint | Query | Quyền / dữ liệu |
| --- | --- | --- |
| `/admin/reports/revenue` | `from,to,tour_id?,source?,group_by=day|month` | report.read; `{currency,timezone,gross_booking_amount,cash_received,refund_amount,net_received,booking_count,series[],is_demo}` |
| `/admin/reports/occupancy` | `from,to,tour_id?` | report.read; `[{departure_id,tour_title,starts_at,capacity,held_seats,confirmed_seats,occupancy_percent}]` |
| `/admin/reports/group-quotes` | `from,to,group_type?` | report.read; `{request_count,by_status,won_count,won_rate_percent}` theo cohort created_at |
| `/admin/reports/revenue/export` | Như report + `format=xlsx|csv` | report.export; stream |
| `/admin/bookings/export` | Như list booking + `format=xlsx|csv` | report.export + booking.read; stream, audit action export |
| `/admin/tours/export` | `status?,format=xlsx|csv` | catalog.write; stream |

`from/to` bắt buộc, inclusive ngày VN, from<=to, max366 ngày; decimal percentages số thập phân 0..100, không Money. Revenue dùng paid_at và completed_at refunds; gross_booking_amount và count dùng created_at của đơn confirmed/completed. `is_demo` phân biệt dữ liệu thử, không trộn báo cáo tài chính thật.

## 9. Health, lỗi và contract tests

`GET /healthz` ngoài `/api/v1`: 200 `{"status":"ok"}` nếu process sống, không DB. `GET /readyz`: 200 `{"status":"ready"}` khi DB/migrations sẵn sàng; 503 `{"status":"not_ready"}` khi lỗi, không leak connection string. Host dùng health path phù hợp; không dùng health endpoint để thay test checkout.

| Code | HTTP | Frontend xử lý |
| --- | --- | --- |
| `VALIDATION_ERROR`, `FEATURE_NOT_AVAILABLE` | 422 | Lỗi field hoặc nhãn tính năng chưa mở, giữ form |
| `AUTH_REQUIRED`, `TOKEN_EXPIRED` | 401 | Admin refresh tối đa một lần; guest lập session mới, không tự gắn đơn cũ |
| `FORBIDDEN`, `CSRF_INVALID`, `CAPTCHA_INVALID` | 403 | Không retry tự động, hướng dẫn thao tác lại |
| `NOT_FOUND`, `BOOKING_NOT_FOUND` | 404 | Trang không tìm thấy, không leak tồn tại |
| `SOLD_OUT`, `SALES_CLOSED` | 409 | Reload departure, gợi ý lịch khác |
| `QUOTE_EXPIRED`, `PRICE_CHANGED` | 409 | Quote mới, khách xem/xác nhận lại |
| `HOLD_EXPIRED`, `PAYMENT_EXPIRED` | 409 | Khoá CTA thanh toán cũ, tạo đơn/intent mới theo trạng thái |
| `VERSION_CONFLICT`, `INVALID_STATE`, `PAYMENT_IN_PROGRESS` | 409 | Fetch mới, không overwrite dữ liệu |
| `IDEMPOTENCY_CONFLICT`, `REQUEST_IN_PROGRESS` | 409 | Không đổi key mù quáng, lấy trạng thái hoặc retry key cũ sau backoff |
| `RATE_LIMITED` | 429 | Đợi Retry-After |
| `INTERNAL_ERROR`, `SERVICE_UNAVAILABLE` | 500/503 | Hiển thị request ID; GET retry có backoff, POST chỉ retry cùng key |

Contract tests bắt buộc ở M1–M7: envelope/status; money/date/enum; pagination/filter cùng departure; draft không lộ; RBAC từng vai trò; object access/session isolation; lookup không trả PII; CSRF/CORS/Turnstile lỗi; giá client không được tin; quote stale; concurrent booking; payment retry/expiry/cancel/refund; mock route không có production; file type/font/sums/permissions.

## 10. Sau MVP — chưa mount endpoint

Customer signup/login/OTP/OAuth, coupon validation, newsletter subscribe, saved passengers, review submission, self-service cancel/change, provider webhooks, electronic invoice, roster upload và contract signing sẽ có contract riêng khi triển khai. Webhook thật phải xác minh chữ ký/raw body + amount/currency/reference, dedupe event ID, xử lý muộn bằng reconciliation. Không sử dụng `/mock-confirm` làm cổng thanh toán thật.

## Bổ sung hợp đồng Admin settings và poster (22/11/2026)

Các giá trị giữ chỗ20 phút/cọc30%/giới hạn20 khách ở trên là seed demo. Runtime đọc cấu hình/snapshot theo database-schema.

- `GET /admin/site-settings/{key}` trả `{key,value,version,updated_at}`. `PUT /admin/site-settings/{key}` body `{value,expected_version}` →200 cùng DTO, stale409 VERSION_CONFLICT. Chỉ admin; allowlist/schema/validation theo database-schema, ghi audit. Thêm key `booking_defaults`, `company_profile`; `response_sla` gồm business_days/message. Không nhận secret qua settings.
- Public `GET /site-settings` chỉ trả dữ liệu công ty/liên hệ/SLA đã cấu hình và `is_demo`; không trả nội bộ/audit. Trường liên hệ trống =>null. Admin form có preview và nhãn đang dùng dữ liệu mẫu.
- `GET /booking-defaults` public trả `{max_travellers,hold_minutes,is_demo}` để dựng form; giá và tuổi chính thức lấy departure/quote. Quote trả thêm `settings_version,hold_minutes,max_travellers`; booking giữ snapshot dù Admin đổi default giữa checkout.
- Giá, cọc, hạn trả, tuổi/chỗ theo departure vẫn do catalog.write sửa; settings global chỉ admin. Policy huỷ chỉnh qua tour editor. Refund tiếp tục nhân viên xét duyệt theo snapshot, chưa có self-service.
- TourWrite thêm `poster_media_id?`; TourDetail thêm `poster: Media|null,poster_revision`. `POST /admin/tours/{id}/poster-verification` body `{expected_poster_revision,confirmed:true}` cần catalog.write, ghi người/thời gian kiểm. Server kiểm revision và cập nhật xác nhận trong lock tour; mutation giá/lịch phải invalidate xác nhận trong cùng thứ tự lock đã quy định. Không gửi poster chưa xác nhận ra public.
- Admin được cảnh báo “Kiểm tra poster khi thay đổi giá/lịch”; phải xem preview trước xác nhận. Dữ liệu có cấu trúc là nguồn tính tiền; không OCR tự ghi giá từ poster trong MVP. File giới hạn/validation dùng media upload hiện có, không endpoint upload riêng.

## Trạng thái triển khai M1 (22/11/2026)

OpenAPI máy đọc cho endpoint nền tảng ở `backend/internal/platform/openapi.json`, phục vụ qua GET `/openapi.json`; các module chưa triển khai tiếp tục theo đặc tả Markdown, sẽ bổ sung schema cùng code. GET `/admin/auth/session-csrf` có kiểm Origin, đọc cookie refresh và trả csrf_token để phục hồi phiên qua refresh sau reload; không trả raw refresh token. Backend hiện dùng TTL mặc định15 phút/7 ngày/24 giờ/15 phút reset; biến TTL trong env là định hướng cho các phiên bản sau, không giả vờ đã cấu hình được.

Company profile hiện trả `{company_profile,response_sla}` từ GET `/site-settings`, DTO canonical để frontend tiêu thụ; media document IDs chưa thể thêm trước T215 nên validation yêu cầu mảng rỗng. Email/password dùng bcrypt, password12..72 bytes (giới hạn bcrypt), thay cho128 ký tự trong bản đề xuất trước. Reset preview mock mã hoá trong DB, endpoint `/admin/auth/mock-reset-deliveries` chỉ staff.write và non-production, đọc có audit. Logout/reuse/reset thu hồi phiên. GET `/admin/audit-logs` nền tảng giới hạn100 bản ghi; pagination ở module quản trị sau.

Các limit hiện tại bảo vệ theo IP, lưu counter DB; chưa phối hợp limiter tại reverse proxy. API không tin X-Forwarded-For từ client; topology proxy hiện chỉ local một hop, demo phải bổ sung trust-proxy allowlist và test theo host trước triển khai công khai.

## Hợp đồng thực thi M2/M3 — 22/11/2026

OpenAPI `backend/internal/platform/openapi.json` mô tả endpoint đã mount M1–M3; API M4+ trong tài liệu này vẫn là kế hoạch.
- Catalog trả dữ liệu phẳng (`title`, `summary`, `included`, `excluded`, `cancellation_policy`, `seo_title`, `seo_description`); `cover`/`poster` là Media hoặc null. `days`, `stops`, `gallery`, `destinations`, `categories`, `faqs` chỉ có trong detail. Poster chỉ public sau khi xác minh revision hiện tại.
- Admin TourWrite dùng `days:[{day_number,title,body,meals,accommodation}]`, `stops:[{day_number,sort_order,name,latitude,longitude,description}]`, `destination_ids`, `category_ids`, `media_ids`. Tạo mặc định draft. Clone `{slug,code,title}` trả draft mới; không copy departures/booking/review.
- Xác minh poster POST `/admin/tours/{id}/poster-verification` cần `{expected_poster_revision,confirmed:true}`. Chưa đối chiếu thì không gửi confirmed. Sửa tour/giá/lịch xoá xác minh.
- Departure DTO gồm `availability:available|low|sold_out|closed`, `remaining_seats`, `server_time`, prices và surcharges. Phụ thu unit=`per_booking|per_person`. Tiền request/response luôn chuỗi số nguyên. Admin create có thể bỏ tuổi/chỗ của giá để lấy defaults; PATCH gửi đủ bảng giá với `expected_price_version`.
- Upload POST `/admin/media?visibility=public|private` dạng multipart file,alt_text,collection,source_credit; field visibility trong form nếu có phải trùng query. Image10MiB, document20MiB, video50MiB. GET `/media/{id}` chỉ public ready hoặc chữ ký HMAC còn hạn; private download URL TTL5 phút và kiểm tra uploader/quyền. Local storage giữ trong named volume media_data; R2 S3 adapter ký AWS4, không đưa secret frontend.
- Public GET `/site-settings` trả trực tiếp `{company_profile,response_sla}`. Company document IDs phải public/ready image hoặc PDF, tối đa20; tệp đang được công ty tham chiếu không được archive.
- Quote DTO giữ breakdown `{items,age_bands,total_amount,deposit_amount,balance_after_deposit,discount_amount,cancellation_policy,tour,starts_at,balance_due_at,allowed_payment_options}`, các field còn lại theo mục5.1. Quote không giữ chỗ.
- Booking DTO có `tour_snapshot`, `price_snapshot`, `policy_snapshot`, `settings_snapshot`, `balance_amount`, `server_time`, `version`, cùng các trường nghiệp vụ; không trả guest/staff principal IDs. Creator detail có passengers/items, Admin detail thêm events/notes. List bỏ special_requests/consent/snapshots giá-policy-settings và không trả passengers.
- Lookup Summary không có booking UUID, contact name, passengers, ngày sinh hoặc special_requests; chỉ code/status/payment_status/payment_option/amounts, tour, contact_masked và giờ/hạn. Token chỉ lưu trong memory client; hash SHA256 trong DB, TTL15 phút.
- Tạo đơn và cancel có Idempotency-Key UUID; cùng principal/route/key/payload trả DTO gốc. Tính hash bỏ Turnstile; retry đã commit không xác minh token một lần lần nữa. Conflict payload409, lock timeout409 REQUEST_IN_PROGRESS. Expiry sweep mỗi30 giây cộng kiểm tra trực tiếp khi đọc/ghi; không cần Redis.
- Payment/thông báo M4 chưa có endpoint trong runtime. UI ghi rõ chưa thanh toán và không gửi email/SMS thật.
