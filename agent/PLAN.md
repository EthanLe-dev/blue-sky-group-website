# Kế hoạch triển khai Blue Sky Group

Ngày: 22/11/2026. Trạng thái: **thiết kế trước triển khai**. Phạm vi hiện tại là MVP; chưa có mã ứng dụng, migration hay Dockerfile. Git được khởi tạo để đẩy tài liệu và cấu hình. Checkbox chỉ đánh dấu khi có kết quả kiểm chứng.

Quy trình Git: `main` mặc định và chỉ nhận bản đã nghiệm thu; `develop` là nhánh tích hợp code chính. Mỗi module lớn có nhánh `feature/<module>` từ `develop`; task DONE được commit rõ ID và push ngay, module qua CI thì PR vào `develop`, đợt phát hành qua UAT mới PR vào `main`. CI cùng cấu hình cho push/PR hai nhánh; CD demo chỉ chạy trên push `main`. Xem [AGENTS](../AGENTS.md).

## 1. Nguồn và quyết định thiết kế

Nguồn: [MVP](../docs/mvp-requirements.md), [Final](../docs/final-requirements.md), [Architecture](../docs/architecture.md), [Database](../docs/database-schema.md), [API](../docs/api-spec.md), [AGENTS](../AGENTS.md).

| Chủ đề | Quyết định dùng để triển khai / điều kiện chốt |
| --- | --- |
| Phạm vi | MVP làm các mục 3.1 và phụ lục MVP; phần final hoãn giữ trong backlog. FR-82 không kéo coupon vào MVP; FR-83 dùng refund giả lập; FR-88 có CRM cơ bản, không có đăng nhập khách. |
| Runtime | Next.js App Router + React SPA (đề xuất Vite) + Go/Gin + PostgreSQL; phiên bản framework/Go chốt khi scaffold và ghi lockfile/go.mod. Node 22, PostgreSQL 17 là baseline cấu hình ban đầu, rà hỗ trợ trước triển khai. |
| Deploy miễn phí | Đề xuất **Render Free (Go API) + Neon Free (PostgreSQL)**; public/admin theo hướng Vercel trong yêu cầu, cần kiểm tra điều kiện gói phù hợp website doanh nghiệp trước deploy. R2 cho object storage, Turnstile cho form. Chưa tạo tài khoản hay deploy. |
| Sai lệch nguồn | D1 dùng SQLite, không tương thích thay thế PostgreSQL cho thiết kế khoá này. Cloudflare Containers yêu cầu Workers Paid. Giả định “Go + PostgreSQL trên Cloudflare free” trong architecture/MVP cần được BA cập nhật trước nghiệm thu deploy; không tự đổi backend sang Workers JS. |
| Giới hạn demo | Render Free ngủ sau 15 phút vắng request, khởi động lại khoảng 1 phút; filesystem tạm. Render PostgreSQL Free hết hạn 30 ngày nên đề xuất Neon. Dữ liệu/file phải ở DB/R2. Mỗi lần triển khai kiểm tra lại quota và điều kiện miễn phí. |
| Hết hạn giữ chỗ | Demo không dựa vào timer luôn chạy. Worker dọn khi process hoạt động và khi khởi động; giao dịch đọc/ghi kho kiểm tra và giải phóng hold quá hạn trước xử lý. Thời hạn vẫn theo DB dù app ngủ. |
| Thời gian/tiền | UTC ở DB, Asia/Ho_Chi_Minh khi hiển thị; VND nguyên đồng. Hold 20 phút (trong khung 15–30 phút nguồn). Cọc mặc định đề xuất 30%, làm tròn lên; hạn phần còn lại trước khởi hành 7 ngày, tour cận ngày chỉ full. BA cần xác nhận trước seed tour thật. |
| Quy tắc chỗ/tuổi | Đề xuất adult ≥12, child 2–<12, infant <2 tính theo ngày đi; adult/child mỗi người 1 chỗ, infant 0. Lưu cấu hình theo departure, cần BA xác nhận phù hợp xe/chuyến bay. Tối thiểu 1 adult, tối đa 20 khách/đơn lẻ; vượt dùng báo giá đoàn. |
| Phân biệt trạng thái | Booking `pending_payment/confirmed/expired/cancelled/completed`; payment summary `unpaid/partially_paid/paid/partially_refunded/refunded`. Thu đủ cọc => confirmed + partially_paid, không đánh dấu paid. |
| Tra cứu guest | Mã ngẫu nhiên + email/SĐT chuẩn hoá + Turnstile + rate limit; chỉ trả tóm tắt đã che thông tin, token lookup read-only. Preview thông báo/passenger chỉ creator session hoặc nhân viên được cấp quyền. |
| Nội dung | Demo gắn nhãn seed/mẫu; không bịa review thật, giấy phép, hotline hay SLA tư vấn. Hồ sơ doanh nghiệp, địa chỉ, ảnh, điều khoản và SLA phản hồi do công ty cung cấp. |
| PDF/chứng từ | PDF tour/báo giá trong MVP. Chứng từ demo nếu có ghi “không phải hoá đơn điện tử”; không tích hợp phát hành hoá đơn. |
| Newsletter | Chỉ UI chưa kích hoạt gửi/thu lead newsletter công khai; FR-88 quản lý contact hiện có. Không hiện toast đăng ký thành công giả. |

Nguồn đối chiếu hạ tầng (22/11/2026): [D1](https://developers.cloudflare.com/d1/), [Containers](https://developers.cloudflare.com/containers/), [Render Free](https://render.com/docs/free), [Gin trên Render](https://render.com/docs/deploy-go-gin), [Neon Free](https://neon.com/pricing). Chính sách nhà cung cấp có thể thay đổi; đây là phương án đề xuất, không cam kết SLA/chi phí production.

## 2. Thứ tự và điều kiện hoàn thành

Chuỗi phụ thuộc chính: M0 → M1 → M2 → M3 → M4 → M6 → M7. M5 có thể thực hiện sau M2; không cần tách microservices. Ước lượng dưới đây là **ngày công kỹ thuật**, chưa gồm chờ nội dung/tài khoản/duyệt UI; hiệu chỉnh sau M1, không phải cam kết ngày giao.

### M0 — Chốt cơ sở thiết kế (2–3 ngày công)

- [x] Đọc yêu cầu, soạn AGENTS, PLAN, schema, API, CI và cấu hình local ban đầu.
- [x] Chia backlog trong `agent/TASKS.md`, quy định UI/UX dùng taste-skill trong `docs/design-style.md`, tiêu chuẩn kiểm thử trong `docs/testing-strategy.md` và nhật ký bàn giao trong `docs/PROGRESS.md` (22/11/2026; hoàn tất tài liệu, chưa triển khai ứng dụng).
- [ ] BA xác nhận cọc, tuổi, số chỗ em bé, ngày đến hạn, chính sách huỷ/hoàn và SLA gọi lại/báo giá.
- [ ] Chốt host demo, điều chỉnh tiêu chí deploy Cloudflare trong nguồn, xác nhận quota và điều kiện sử dụng.
- [ ] Nhận nhận diện, danh sách tour/giá/lịch khởi hành, nội dung pháp lý, media có quyền dùng; chốt page map và mẫu UI mobile.
- Đầu ra: quyết định nghiệp vụ được ghi, không còn giả định chặn bán tour thật. Có thể scaffold/demo bằng dữ liệu mẫu trước khi hoàn tất nội dung thật.

### M1 — Nền tảng ứng dụng và bảo mật (3–5 ngày công)

- [ ] Scaffold `frontend/`, `admin/`, `backend/` đúng architecture; pin công cụ/dependencies, lockfiles, npm scripts theo AGENTS.
- [ ] Config loader validate env; structured logs có request ID; `/healthz`, `/readyz`; kết nối DB có timeout/pool.
- [ ] Migration runner + seed idempotent; tài khoản admin bootstrap từ secret một lần, không mật khẩu hardcode. Dữ liệu test độc lập.
- [ ] Auth JWT access + refresh rotation/revocation, cookie/CSRF/CORS, RBAC 5 vai trò, audit logs, rate limit và Turnstile adapter.
- [ ] Tạo OpenAPI từ api-spec, test response schema; Dockerfile app/Nginx và mở rộng Compose khi file đã tồn tại.
- [ ] Kích hoạt job CI ứng dụng khi có manifest; migrations và test integration bắt buộc khi có persistence.
- Nghiệm thu: setup sạch theo hướng dẫn chạy được; login/logout/refresh, 401/403, secrets không vào bundle; CI lint/typecheck/build/test đạt.

### M2 — Catalog, nội dung và public website (6–9 ngày công)

- [ ] Migration/catalog API: destination/category/tour, itinerary, gallery, FAQ, route markers, related tours; lịch khởi hành/giá/phụ thu.
- [ ] Admin CRUD, clone tour không copy booking; tạo hàng loạt departure atomic, đóng/mở bán; archive thay xoá dữ liệu đã tham chiếu (FR-80–82).
- [ ] Home, list/filter/autocomplete/sort giữ query URL, tour detail, sticky booking, số chỗ thật, PDF itinerary placeholder chờ M6 (FR-01–08).
- [ ] Blog/categories/tags, destination landing, school/company, about/contact/policies/FAQ/gallery (FR-40–45, 60, 63–64, 66–73).
- [ ] Admin nội dung, banner, review/testimonial có kiểm duyệt và nguồn; SEO fields, sitemap, canonical, OG/schema; bản đồ route/văn phòng.
- [ ] Favourite/recent localStorage chỉ ID/slug, share link; UI disabled coupon/newsletter (FR-10–12, 26, 74).
- Nghiệm thu: responsive, keyboard, loading/empty/error; không lộ draft; tour không đủ policy/giá không publish; bộ lọc cho kết quả đúng.

### M3 — Booking, inventory, pricing (5–7 ngày công)

- [ ] Quote tính từ DB và snapshot immutable; validation tuổi, phụ thu, chính sách/cọc, người liên hệ và hành khách.
- [ ] Checkout 3 bước, create booking giữ chỗ 20 phút, countdown server time; cùng service cho admin tạo đơn (FR-20–25, 84).
- [ ] Expiry worker restart-safe + xử lý lazy expiry trong transaction; lookup guest che PII; admin sửa passenger/note/cancel (FR-31, 83).
- [ ] Test trên PostgreSQL: N yêu cầu tranh K chỗ không overbook; rollback không rò hold; request lặp cùng key chỉ một booking; khác payload 409; expire/cancel chạy hai lần không hoàn chỗ hai lần.
- Nghiệm thu: snapshot không đổi khi sửa giá tour, không tin tổng tiền client; lịch hết chỗ gợi ý lựa chọn khác; tiền không sai làm tròn.

### M4 — Mock payment và notification (3–5 ngày công)

- [ ] QR giả không thể dùng chuyển khoản thật, nhãn demo xuyên suốt; full/deposit/balance, intent và confirm idempotent.
- [ ] Transaction payment/booking/inventory/outbox chung; confirm sau expiry bị từ chối; hai intent cạnh tranh không thu dư.
- [ ] Email/SMS mock log trong DB, preview cho creator/admin; UI nói rõ giả lập. Retry outbox không nhân đôi notification (FR-27–29).
- [ ] Admin đối soát/refund giả lập, audit; refund không tự mở lại chỗ booking còn confirmed.
- [ ] Test payment cạnh tranh với expiry/cancel, retry/duplicate, cọc còn dư nợ, production từ chối mock.
- Nghiệm thu: duyệt tour → đặt → trả cọc/toàn bộ → xác nhận → tra cứu hoạt động với DB thật, không gửi tiền/tin thật.

### M5 — Sales, khách đoàn, khách hàng (3–5 ngày công)

- [ ] Form báo giá school/company, callback, xác nhận đã nhận + SLA cấu hình, Turnstile/rate limit (FR-41, 61, 86).
- [ ] Admin assignment/status/notes/follow-up due; đề xuất báo giá versioned và attachment riêng tư, danh sách khách/lịch sử đơn/notes (FR-85, 88).
- [ ] Không tự gộp khách chỉ vì trùng SĐT/email; chỉ nhân viên được quyền xác nhận gắn contact.
- Nghiệm thu: sales thao tác đúng quyền, không xem route admin bằng guest; không có public tracking/upload đoàn/ký hợp đồng ở MVP.

### M6 — Báo cáo, export, analytics (3–5 ngày công)

- [ ] Doanh thu thu thực = payments succeeded trừ refunds succeeded, phân biệt gross booking value, demo data; ngày theo Việt Nam, lọc tour/nguồn/ngày.
- [ ] Số đơn, tỷ lệ lấp đầy theo departure, tỷ lệ chốt báo giá; dashboard không nhân bản tiền do JOIN passenger (FR-95, 96, 98).
- [ ] Excel/CSV báo cáo/đơn/tour, PDF lịch trình/báo giá, chứng từ demo nếu thống nhất; font tiếng Việt, quyền tải file, chống spreadsheet formula injection (FR-08, 48, 99).
- [ ] GA4 + Search Console bằng env; events `view_item`, `begin_checkout`, `purchase`, `generate_lead`; purchase mock có nhãn demo và chỉ phát một lần. Không gửi PII hoặc access token.
- Nghiệm thu: mở file bằng phần mềm thực, tổng khớp DB; PDF tiếng Việt đúng; GA4 DebugView thấy event khi cấu hình tài khoản.

### M7 — Tích hợp, UAT và deploy demo (4–6 ngày công)

- [ ] E2E public/admin, cross-browser hai phiên bản mới, mobile 360/390/768 và desktop 1440; kiểm tra zoom, keyboard, contrast và reduced motion.
- [ ] Demo seed có tour trong/quốc tế, school/company, hết/sắp hết chỗ, giá trẻ em/em bé, cọc/full và đơn hết hạn; xoá seed không ảnh hưởng dữ liệu thật.
- [ ] Deploy Go listen `0.0.0.0:$PORT`, health check; Neon TLS + migrations qua direct URL; app dùng pooled URL; worker dọn khi awake và lazy expiry.
- [ ] Vercel roots public/admin; configure API proxy/same-origin theo API spec, R2 public/private tách quyền, Turnstile hostname thực; không wildcard CORS.
- [ ] Test cold start, restart, timeout, mất mạng sau create/confirm; không retry POST không có idempotency key.
- [ ] UAT đủ 11 tiêu chí mục 8 MVP, ghi evidence/URL, demo cho nhóm 30+, trường và doanh nghiệp; sửa lỗi và bàn giao runbook setup/seed/deploy.
- Nghiệm thu: đủ tiêu chí chức năng demo; không dùng kết quả demo free tier chứng minh NFR tải/SLA production.

## 3. Backlog sau MVP

| Giai đoạn | Hạng mục và cổng nghiệm thu |
| --- | --- |
| P1 — Production | Host ổn định/worker luôn chạy, staging riêng, domain/TLS, migration/deploy rollback, backup và restore drill; real payment callback ký số/idempotency/reconciliation, real email/SMS, customer auth + reset password/OTP, 2FA admin, Redis khi cần. Đối soát sandbox trước live. |
| P1 — NFR | Test 100 phiên hoạt động thực; p95 API thường ≤500ms, booking ≤2s, trang server ≤2s; dự phòng 200–300 phiên; LCP ≤2.5s, INP ≤200ms, CLS ≤0.1. Mục tiêu uptime 99.5%, RPO≤24h/RTO≤4h phải đo và chốt riêng cho payment. Pentest không còn high/critical, rà pháp lý/retention với công ty. |
| P2 — Chuyển đổi | Coupon có hạn/quota atomic, abandoned checkout có consent, chat/Zalo, chuỗi chăm sóc, review sau chuyến, wishlist tài khoản, yêu cầu huỷ/đổi theo policy, báo cáo funnel và newsletter thật. |
| P3 — Mở rộng | So sánh tour, upload đoàn, public quote tracking có auth, hợp đồng/ký, CRM/kế toán/e-invoice, đa ngôn ngữ. Đánh giá từng mục S/C theo nhu cầu; không mặc định làm hết trước go-live. |

## 4. Cấu hình, CI và vận hành local

- `docker-compose.yml` hiện chỉ chạy PostgreSQL; Redis optional, chưa cần MVP. Không khai báo app build khi chưa có Dockerfile. `.env.example` là hợp đồng cấu hình **dự kiến** cho M1, Compose chỉ đọc biến PostgreSQL/cổng Redis.
- Tạo `.env` từ mẫu, chạy `docker compose --env-file .env config --quiet`, `docker compose up -d postgres`. Không commit `.env`; `down` giữ volume, `down -v` xoá dữ liệu.
- Backend trên host dùng `127.0.0.1`; khi container hoá đổi host thành service `postgres`. Migration dùng DB owner riêng ở production; app dùng role tối thiểu. Không tự chạy migration đồng thời trong mỗi replica.
- CI chạy kiểm tra tài liệu/Compose ngay. Khi có manifest, không skip các script chất lượng; service PostgreSQL riêng cho integration. Có workflow hook deploy demo, chưa có deployment ứng dụng hoặc secret production.
- CD demo dùng `.github/workflows/deploy.yml`: chỉ chạy sau CI thành công của push lên `main`, với đúng SHA đã kiểm tra. Khi cả ba app và backend Dockerfile tồn tại, workflow đòi ba deploy hook trong GitHub environment `demo` (`RENDER_DEPLOY_HOOK_URL`, `VERCEL_PUBLIC_DEPLOY_HOOK_URL`, `VERCEL_ADMIN_DEPLOY_HOOK_URL`), gọi Render trước rồi Vercel. Hiện thiếu app nên ghi skip có chủ đích; chưa deploy. Hook chỉ xác nhận đã nhận yêu cầu, không chứng minh ứng dụng đã healthy. Cần tạo project Render/Vercel, Neon, cấu hình env và migration/health check trước khi bật deploy thật.
- Khi scaffold phải thêm kiểm tra migration up trên DB sạch, rollback có kiểm soát, OpenAPI contract, build Docker và E2E trước M7; CI ban đầu không tuyên bố đã có các check này.

## 5. Rủi ro và thông tin cần cung cấp

| Vấn đề | Hành động |
| --- | --- |
| Free tier ngủ/quota | Chỉ demo; lazy expiry, thông báo đang khởi động, không lưu file local lâu dài; nâng cấp trước bán thật. |
| Thay đổi luật giá/chỗ sau khi có đơn | Snapshot và price_version; tăng/giảm capacity có khoá và điều kiện; không viết đè booking cũ. |
| Chưa có nội dung thật | Seed gắn nhãn, không xuất bản số giấy phép/đánh giá/ảnh chưa được duyệt. |
| Privacy và bảo mật | Minimise dữ liệu, audit RBAC, signed URL, retention do công ty/pháp chế chốt; không coi tài liệu này là kết luận pháp lý. |
| Tài khoản tích hợp | Cần quyền repo/host/DB/R2/Turnstile/GA4/Search Console khi triển khai; chỉ lưu credential trong secret manager. |
| Ngân sách/nhân sự/deadline | Chưa được cung cấp; ước lượng M0–M7 khoảng 29–45 ngày công, có thể điều chỉnh sau scaffold và duyệt giao diện. |
