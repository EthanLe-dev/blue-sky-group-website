# Kế hoạch triển khai Blue Sky Group

Ngày: 22/11/2026. Trạng thái: **M0 hoàn tất cho demo, M1 nền tảng đã triển khai**. Phạm vi hiện tại là MVP; frontend/admin/backend và Docker/migration nền tảng đã có. Git được khởi tạo để đẩy tài liệu và cấu hình. Checkbox chỉ đánh dấu khi có kết quả kiểm chứng.

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
- [x] Người dùng cho phép giả định cho demo; bắt buộc sửa các giá trị qua Admin (22/11/2026).
- [x] Chốt Render Free + Neon Free trong quota cho demo; cập nhật nguồn. Kiểm quota tài khoản và smoke test thực tế ở T704 trước deploy.
- [x] Inventory và page-map/wireframe demo tại docs/content-inventory.md và docs/page-map.md. Người dùng cho phép dữ liệu thiếu dùng mẫu sửa ở Admin; xác minh nội dung thật trước bán chính thức.
- Đầu ra: quyết định nghiệp vụ được ghi, không còn giả định chặn bán tour thật. Có thể scaffold/demo bằng dữ liệu mẫu trước khi hoàn tất nội dung thật.

### M1 — Nền tảng ứng dụng và bảo mật (3–5 ngày công)

- [x] Scaffold frontend Next.js/GSAP, React/Vite Admin, Go/Gin; có lockfile, scripts kiểm tra và Dockerfiles.
- [x] Config loader validate env; structured logs có request ID; `/healthz`, `/readyz`; kết nối DB có timeout/pool.
- [x] Migration runner + seed idempotent; tài khoản admin bootstrap từ secret một lần, không mật khẩu hardcode. Dữ liệu test độc lập.
- [x] Auth JWT access + refresh rotation/revocation, cookie/CSRF/CORS, RBAC 5 vai trò, audit logs, rate limit và Turnstile adapter.
- [x] Tạo OpenAPI từ api-spec, test response schema; Dockerfile app/Nginx và mở rộng Compose khi file đã tồn tại.
- [x] Kích hoạt job CI ứng dụng khi có manifest; migrations và test integration bắt buộc khi có persistence.
- Kiểm chứng local: setup Docker, login/logout/refresh/RBAC/settings, vet/race/integration/OpenAPI, frontend/admin lint/typecheck/test/build và hai browser E2E đạt. CI remote và merge gate theo PR #1; không coi local pass là remote pass. T108b lookup storage chuyển M3 sau bảng booking T301; OpenAPI hiện bao phủ endpoint M1, mở rộng cùng mỗi module. T113/T114 cấu hình Admin đã có, media document picker theo T215.

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

- [x] Form báo giá trường học/doanh nghiệp và gọi lại, biên nhận theo SLA cấu hình, Turnstile và giới hạn gửi (FR-41, 61, 86).
- [x] Admin phân công, chuyển trạng thái, ghi chú và theo dõi hạn. Báo giá có phiên bản, tệp đính kèm riêng tư và ghi nhận kết quả ngoại tuyến (FR-85).
- [x] CRM có hồ sơ, ghi chú và lịch sử đơn sau xác minh (FR-88). Nhân viên xác nhận gắn từng đơn, kể cả khi trùng thông tin liên hệ.
- CI trên `b76c2f0` đạt 6/6 job, gồm Go race/PostgreSQL và Playwright. Chi tiết T501–T504 nằm trong [tiến độ](../docs/PROGRESS.md). [GitHub Actions 45009896743](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45009896743)

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
- CD demo dùng `.github/workflows/deploy.yml` với công tắc `DEMO_DEPLOY_ENABLED` và environment demo. Chỉ chạy sau CI của push main. Render nhận đúng SHA, Vercel dựng nguồn từ SHA đã kiểm tra. Workflow chờ readiness và chạy smoke trên URL ổn định. Cấu hình và rollback xem docs/demo-release.md.
- Khi scaffold phải thêm kiểm tra migration up trên DB sạch, rollback có kiểm soát, OpenAPI contract, build Docker và E2E trước M7; CI ban đầu không tuyên bố đã có các check này.

## 5. Rủi ro và thông tin cần cung cấp

| Vấn đề | Hành động |
| --- | --- |
| Free tier ngủ/quota | Chỉ demo; lazy expiry, thông báo đang khởi động, không lưu file local lâu dài; nâng cấp trước bán thật. |
| Thay đổi luật giá/chỗ sau khi có đơn | Snapshot và price_version; tăng/giảm capacity có khoá và điều kiện; không viết đè booking cũ. |
| Chưa có nội dung thật | Seed gắn nhãn, không xuất bản số giấy phép/đánh giá/ảnh chưa được duyệt. |
| Privacy và bảo mật | Minimise dữ liệu, audit RBAC, signed URL, retention do công ty/pháp chế chốt; không coi tài liệu này là kết luận pháp lý. |
| Tài khoản tích hợp | Cần quyền repo/host/DB/R2/Turnstile/GA4/Search Console khi triển khai; chỉ lưu credential trong secret manager. |
| Ngân sách/nhân sự/deadline | Người dùng không đặt deadline bắt buộc; ước lượng M0–M7 khoảng 29–45 ngày công, có thể điều chỉnh sau scaffold và duyệt giao diện. |

## Quyết định bổ sung ngày 22/11/2026 — Admin và poster

Người dùng cho phép giả định giá/đặt chỗ và thông tin công ty để triển khai demo, với điều kiện **đều sửa được ở Admin**. Quyết định này thay thế các mục đang chờ BA về giá trị ban đầu trong tài liệu; chưa phải duyệt chính sách bán thật. Không có deadline bắt buộc, vẫn thực hiện theo dependency và quality gates.

- Mặc định demo: giữ chỗ20 phút, cọc30%, trả dư trước ngày đi7 ngày, adult≥12/child2–<12/infant<2; adult/child1 chỗ, infant0; tối đa20 khách/đơn. Policy huỷ demo: nhân viên xét duyệt theo chính sách hiển thị, không tự hoàn tiền. SLA demo: phản hồi trong1 ngày làm việc. Tất cả hiển thị nhãn demo khi chưa được thay bằng dữ liệu thật.
- Admin quản lý mặc định đặt chỗ, giá/tuổi/chỗ/cọc/hạn trả theo lịch khởi hành, policy huỷ theo tour, thông tin doanh nghiệp/liên hệ/giấy phép/tài liệu/SLA. Thông tin chưa có dùng placeholder rõ ràng, không bịa giấy phép hay số điện thoại thật; link gọi/Zalo chưa cấu hình phải disabled.
- Thay đổi cấu hình có validation, version, audit; chỉ tác động đơn mới. Quote đang hiệu lực dùng snapshot cấu hình đã chốt; booking cũ giữ nguyên giá/policy/hạn, không cập nhật ngược. Sửa giá/tuổi/chỗ/cọc của departure tăng price_version và yêu cầu báo giá lại.
- Tour do team thiết kế poster. Admin upload/thay poster và nhập dữ liệu tour có cấu trúc; poster không thay trường giá/lịch/itinerary dùng tìm kiếm, SEO và checkout. Chưa nhận được ảnh mẫu trong chat hoặc workspace (chỉ có logo), cần bổ sung để chốt bố cục/tỉ lệ poster.
- Chọn Render Free cho Go API và Neon Free cho PostgreSQL để demo nhỏ, trong quota miễn phí; không tự nâng gói trả tiền. Cần smoke test thực tế, nhất là cold start và PDF, trước khi xác nhận đủ năng lực. Cloudflare vẫn cho Turnstile/R2, không làm Go host/DB. Miễn phí backend/DB không đồng nghĩa đã xác nhận chi phí toàn hệ thống.


Nguồn kiểm tra ngày22/11/2026: [Render Free](https://render.com/docs/free) (750 giờ/workspace/tháng, ngủ sau15 phút) và [Neon Free](https://neon.com/pricing) (100 CU-giờ/project/tháng,1GB storage). Khi deploy kiểm tra lại gói/quota và giới hạn chi tiêu; không cam kết miễn phí vô hạn.

### Cập nhật poster mẫu — 22/11/2026

Đã nhận poster Ninh Chữ1051×1497px từ người dùng, lưu [tham chiếu](../docs/poster-reference.md). T003a hoàn tất thu nhận/phân tích ảnh; T003 vẫn cần nội dung tour đầy đủ và hồ sơ quyền ảnh. Ghi chú “chưa nhận ảnh” trước đó là trạng thái lịch sử, không còn là blocker. Giá/liên hệ chép từ ảnh chưa tự áp dụng làm dữ liệu hiện hành.

### GSAP và frontend scaffold — 22/11/2026

Người dùng yêu cầu cài GSAP cho animation. T101 scaffold Next.js App Router + TypeScript và cài `gsap`/`@gsap/react` trên nhánh `feature/public-ui` từ develop. Các trang marketing sẽ dùng GSAP có mục đích, reduced motion và cleanup; checkout/admin ưu tiên rõ ràng. Chưa gọi T208/T703 DONE chỉ vì có thư viện.

### Bàn giao M1

README và docs/platform-runbook.md là nguồn lệnh chạy hiện tại. Docker profile app thực thi migration trước API, bootstrap tách riêng; credentials trong .env ignored. PR #1 vào develop chứa M0/M1. Main giữ nguyên cho tới release nghiệm thu. CI remote là cổng merge, xem trạng thái trực tiếp trên PR.

Mốc xác nhận: PR #1 đã PASS CI run44953203198 và merge develop tại ce0bbb9 ngày22/11/2026. Main chưa nhận M1. M0/M1 nền tảng đã bàn giao, tiếp tục M2; source code chưa có catalog/booking/payment.

### Triển khai M2/M3 — đang QA ngày 22/11/2026
- Feature/catalog chứa schema/API/catalog/media; feature/booking tiếp nối dependency catalog để kết nối checkout hoàn chỉnh. Main giữ nguyên.
- Logo dùng file JPG công ty chỉ định. Ảnh điểm đến public domain ghi nguồn trong docs/media-credits.md; poster gốc được nhúng không cắt chữ. Dữ liệu bổ sung có nhãn demo, sửa qua Admin.
- Đã cho phép chạy local kiểm thử rồi dừng trong đợt này. Không bootstrap lại, không đổi mật khẩu hay xoá volume.
- Module UI/CMS/booking đã có code, đang kiểm browser/contract/taste gate và CI; chưa coi toàn bộ DONE. M4 thanh toán/thông báo và M5 form khách đoàn chưa bật.

### Kết thúc phần chức năng M2/M3 — 22/11/2026

- M2/T201–T218 và M3/T301–T308 đã triển khai và có evidence trong PROGRESS. Nguồn code kiểm chứng `0ee97b82`: [CI44998866529 PASS](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/44998866529), PR#3 tích hợp vào develop; commit bàn giao tài liệu cũng chờ CI trước merge. Không đưa code sang main hoặc deploy cloud trong đợt này.
- Local cuối: frontend/admin lint/typecheck/unit/build PASS (1 warning stylesheet vendor); Go vet/test-race/build PASS, platform80.297s; browser6/6 PASS13.6s; nginx-t và redaction signed query PASS. Docker Compose đã down, ps rỗng, không -v; dữ liệu/mật khẩu Admin giữ nguyên.
- Nghiệm thu UI/T703 còn mở: taste OPEN19 từ vòng3, có bổ sung chức năng sau đó nên cần review UI tiếp ở M7. Chi tiết/waiver/perf trong docs/ui-qa-m2-m3.md; không gọi là UAT hoặc hoàn tất thẩm mỹ.
- Tiếp theo là M4 payment/notification giả lập, sau khi người dùng giao tiếp. M3 hiện chỉ tạo pending hold; không có xác nhận thu tiền hoặc gửi email/SMS. M5 form báo giá, M6 export/PDF và M7 deploy/UAT vẫn chưa triển khai.

### M4 — Thanh toán và thông báo giả lập

T401–T406 dùng chung transaction và thứ tự khóa của booking. Checkout hỗ trợ toàn bộ, cọc và phần còn lại. Admin đối soát, ghi hoàn tiền và xem nhật ký theo quyền. Mọi giao dịch và thông báo M4 dùng dữ liệu giả lập. CI đầu tiên trên PR #4 đã đạt 6/6 job, gồm Go/PostgreSQL và Playwright. PR sẽ chỉ nhập `develop` sau CI của bản cập nhật.

M5 tiếp tục với form báo giá và gọi lại khi được giao. Nghiệm thu UI T703 vẫn mở.

### M6 — Báo cáo, xuất file và analytics

T601–T605 ở PR #6. Dashboard dùng ledger và cohort theo ngày Việt Nam. Excel/CSV theo quyền và giới hạn dòng. PDF nhúng phông tiếng Việt, báo giá theo phiên bản. Analytics chỉ gửi sau khi người dùng đồng ý. Bản mẫu Excel đã mở bằng Microsoft Excel và PDF đã render bằng Poppler. Kiểm tra tích hợp thực hiện trên CI. GA4 DebugView và Search Console chờ cấu hình thuộc tính thật.

### M7 — Đang triển khai

Nhánh feature/mvp-release, PR #7. Fixture demo có tour quốc tế, chỗ sắp hết và hết chỗ, tiền cọc/toàn bộ và đơn hết hạn. CI bổ sung hành trình mất phản hồi thanh toán, đơn tại quầy, đa trình duyệt, khôi phục DB riêng và kiểm tra SHA demo. Hạ tầng mẫu và bàn giao tại docs/demo-release.md, checklist tại docs/uat-mvp.md.

Vercel, Render và Neon dành cho demo. Deploy thật, khởi động lạnh, private file trên R2, Turnstile cloud, GA4/Search Console, rollback cloud và UAT độc lập chờ quyền tài khoản/URL/người thử. T703 còn review thẩm mỹ, không dùng screenshot CI để đánh dấu nghiệm thu thị giác.
