# Task phát triển Blue Sky Group

Ngày 22/11/2026. Nguồn: [PLAN](PLAN.md), [API](../docs/api-spec.md), [schema](../docs/database-schema.md), [thiết kế](../docs/design-style.md), [kiểm thử](../docs/testing-strategy.md). Đây là backlog thực thi; M0/M1 nền tảng đã có code và kiểm chứng local, merge theo CI PR #1.

## Cách sử dụng

- Trạng thái: TODO, DOING, BLOCKED, DONE. Task không có trạng thái ghi riêng là TODO; T001/T002 đã DONE theo quyết định demo. Ghi trạng thái cạnh ID khi nhận việc; chỉ một task DOING cho mỗi Agent. Không mặc định dùng nhiều Agent.
- Task độc lập để kiểm chứng sau khi các phụ thuộc đã DONE. Mỗi task nên hoàn thành trong một phiên (khoảng 1–4 giờ); vượt mức thì tách ID con và ghi lý do trước khi tiếp tục.
- Đọc đầu vào, chốt phạm vi, triển khai đầu ra, chạy kiểm chứng ở hàng tương ứng. DONE cần đường dẫn artifact, lệnh/kết quả hoặc bằng chứng browser trong PROGRESS; không dùng “đã viết code” làm bằng chứng.
- Nếu thiếu tài khoản/nội dung, ghi BLOCKED cùng điều kiện gỡ; chọn task không phụ thuộc. Dùng fixture giả có nhãn cho UI, không coi fixture là tích hợp hoàn tất.
- Test transaction dùng PostgreSQL. Khi sửa hợp đồng, cập nhật API/schema và fixture cùng task. Xem quality gate trong testing-strategy trước khi đóng task.
- Với module lớn, tạo `feature/<module>` từ `develop` trước task đầu; mỗi task DONE có commit riêng chứa ID Txxx và được push ngay lên nhánh module. Module hoàn tất/CI xanh thì PR vào `develop`; đợt phát hành nghiệm thu mới PR `develop` vào `main`. CI hai nhánh chung một workflow. Xem quy tắc chi tiết trong [AGENTS](../AGENTS.md).

## M0 — Cơ sở dự án

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T001 | DONE — Người dùng cho phép giả định và yêu cầu sửa qua Admin (22/11/2026). Ghi quyết định BA: tuổi, cọc, chỗ em bé, hạn trả dư, huỷ/hoàn, SLA vào PLAN | — | Có quyết định/người xác nhận; không đánh dấu đề xuất là đã duyệt |
| T002 | DONE — Chốt Render Free + Neon Free trong quota, cập nhật nguồn (22/11/2026). Chốt host demo và sửa sai lệch Cloudflare/D1 trong tài liệu nguồn | — | Architecture/MVP/PLAN thống nhất và ghi phương án miễn phí/hạn mức được chọn |
| T003a | DONE — Lưu và phân tích poster mẫu Ninh Chữ | — | File gốc1051×1497; docs/poster-reference.md, phân biệt dữ liệu ảnh với thông tin đã xác minh |
| T003 | DONE — inventory demo docs/content-inventory.md; Lập danh mục asset/nội dung thật: logo, ảnh, tour, giấy phép, liên hệ, policy | — | Mỗi asset có nguồn/quyền dùng/trạng thái; thiếu gì được liệt kê |
| T004 | DONE — page map/wireframe docs/page-map.md; Page map và wireframe home/list/detail/checkout/admin theo design-style | T003 hoặc fixture có nhãn | Review luồng mobile, CTA, nội dung cần thiết; chưa gọi là UI đã nghiệm thu |

## M1 — Nền tảng

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T101 | DONE 22/11/2026 — Scaffold Next.js + TypeScript, lockfile và 4 scripts CI; cài GSAP + @gsap/react | — | npm ci, lint, typecheck, test:ci, build chạy; smoke test GSAP timeline thật |
| T102 | DONE — Scaffold React/Vite admin và 4 scripts CI | — | Cài sạch/build; test render route login |
| T103 | DONE — Scaffold Go/Gin, config loader, log redaction, health/ready | — | go vet/test/build; env thiếu và DB lỗi trả đúng lỗi, secret không log |
| T104 | DONE — DB pool, migration runner, test harness PostgreSQL tách dữ liệu | T103 | Up DB sạch, chạy lại không lặp; rollback/restore thử trong DB test |
| T105 | DONE — Migration identity/RBAC/guest/audit, seed quyền, bootstrap admin | T104 | FK/unique đúng; seed hai lần không trùng; không có password mặc định |
| T106 | DONE — Login/logout/refresh rotation/reset password mock | T105 | Sai mật khẩu, expired, refresh reuse, revoke session đều có test |
| T107 | DONE — Middleware RBAC, CSRF, Origin/CORS, rate limit, Turnstile adapter | T106 | Ma trận 401/403, reset/restart limiter, captcha fail; production chặn mock |
| T108a | DONE — Guest session storage/cookie/CSRF | T105,T107 | Token hash, reuse/TTL/origin, session isolation qua API |
| T108b | DONE — Lookup token storage/FK booking (chuyển M3) | T301 | Token hash/TTL/read-only và không cấp quyền thanh toán |
| T109 | DONE — Chuyển api-spec thành OpenAPI và response validator test | T103 | Spec parse hợp lệ, envelope/error/money/date khớp contract |
| T110 | DONE — Public/admin API clients và same-origin proxy | T101,T102,T108a | Cookie/Set-Cookie/CSRF đi qua proxy; lỗi mạng/refresh không tạo vòng lặp |
| T111 | DONE — Dockerfile từng app + Compose app/Nginx | T101,T102,T103 | Build sạch, health check, routing; không đóng gói secret |
| T112 | DONE — CI run44953203198 PASS, PR #1 merged — Kích hoạt CI app, migration/contract tests và branch check runbook | T104,T109,T111 | PR kiểm tra lỗi thật làm CI fail; scripts không bị skip khi có manifest |

| T113 | DONE — API cấu hình booking_defaults/company_profile/SLA, version/audit/snapshot | T105,T107,T109 | Q19/Q20: validation, quyền, version/cache và không thay đơn cũ |
| T114 | DONE — Admin forms cài đặt đặt tour/thông tin công ty | T102,T113 | Save/reload/preview, nhãn demo, lỗi field và stale version |

## M2 — Nội dung và khám phá tour

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T201 | DONE — Tokens, font Việt, Button/Input/Dialog/Toast/empty-error states | T101,T004 | Keyboard, 320px/390px, zoom200%, dấu Việt, reduced motion |
| T202 | DONE — Layout public: nav/mobile menu/footer/hotline | T201 | DONE — Focus trap/restore, skip link, safe area, liên hệ thật hoặc demo rõ |
| T203 | DONE — Migration media/content/catalog/departures/prices/quotes | T104 | Up/FK/checks; capacity âm và ngày sai bị từ chối |
| T204 | DONE — Catalog read API, search/filter/sort/autocomplete | T203,T109 | Bộ lọc ngày/giá/chỗ cùng departure; draft không public, pagination đúng |
| T205 | DONE — Service CRUD/clone/publish/archive tour | T203,T107 | Clone không copy booking; publish thiếu dữ liệu fail; RBAC |
| T206 | DONE — Departure bulk/pricing/capacity/open-close API | T205 | DONE — Bulk rollback toàn bộ; price_version tăng đúng; capacity không dưới chỗ đã dùng |
| T207 | DONE — Admin tour editor và departure/pricing forms | T102,T205,T206 | Save/reload giữ dữ liệu, lỗi field/stale version đúng; không nhận số tiền float |
| T208 | DONE chức năng; taste OPEN19 ở T703 — Home với ảnh thật, tìm kiếm, tour nổi bật và CMS sections | T202,T204 | Browser desktop/mobile, hero không che search/CTA; taste review |
| T209 | DONE — Tour list/filter URL/sort/autocomplete UI | T202,T204 | Back giữ filter, empty/error, keyboard suggestions và retry |
| T210 | DONE — Tour detail/gallery/itinerary/price calendar/sticky CTA | T202,T204 | Gallery bằng phím/vuốt; lịch sold-out; CTA không che nội dung |
| T211 | DONE — Favourite/recent/share bằng ID/slug | T209,T210 | Storage hỏng/disabled vẫn dùng được; không lưu PII |
| T212 | DONE — CMS CRUD posts/pages/banner/FAQ/reviews + admin editor | T203,T107,T102 | Rich text sanitize, approved-only, draft/publish và quyền |
| T213 | DONE — Blog/detail/destination/FAQ/about/policies/gallery | T202,T212 | Slug404, ảnh alt, heading/link đúng, loading/empty/error |
| T214 | DONE — School/company/contact pages và bản đồ markers/office | T202,T212 | Map lazy load, attribution, fallback địa chỉ, không bịa toạ độ |
| T215 | DONE — Media upload local/R2 và private download authorization | T203,T107 | Magic bytes/size/path traversal/private object và signed URL expiry |
| T216 | DONE — SEO meta/canonical/OG/sitemap/robots/schema | T208,T210,T213 | HTML SSR có nội dung; draft không index; structured data không bịa rating |

| T217 | DONE — Schema/API poster media/revision/verification và invalidate khi sửa tour/giá/lịch | T205,T206,T215 | Q21: stale revision, public filtering, MIME; không lấy giá từ ảnh |
| T218 | DONE — Admin upload/preview/verify poster và public viewer responsive | T217,T210 | Ảnh không crop chữ, zoom/keyboard/mobile, structured content vẫn đầy đủ |

## M3 — Giá và đặt chỗ (FR-20–25,31,83–84)

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T301 | DONE — Migration booking/items/passengers/holds/events/idempotency | T203,T108a | Unique quote/code, FK, constraints; migrations trên DB sạch |
| T302 | DONE — Quote service: tuổi, tiền, phụ thu, cọc, snapshot/version | T301,T206,T113 | Boundary tuổi/ngày/cọc làm tròn; stale quote; dùng luật đề xuất chỉ cho demo |
| T303 | DONE — Booking create transaction + idempotency + counters | T302 | N request tranh K chỗ, rollback giữa transaction, duplicate key/hash conflict |
| T304 | DONE — Lazy expiry + restart-safe sweep | T303 | DONE — Expire chạy hai lần; clock sau lock; process ngủ/restart không giữ chỗ sai |
| T305 | DONE — Creator read và guest lookup read-only che PII | T303,T108b | IDOR, wrong contact cùng lỗi, lookup token không confirm payment/đọc passenger |
| T306 | DONE — Admin tạo/sửa/notes/cancel/complete dùng chung service | T303,T304,T107 | Manual+web tranh cùng kho; cancel hai lần; completed không nhả chỗ |
| T307 | DONE — Checkout 3 bước, quote/update/validation và hold countdown | T110,T210,T302,T303,T304 | Quay lại giữ input, server total, timeout/sold-out/price changed; không lưu PII localStorage |
| T308 | DONE — Lookup UI và admin booking list/detail | T305,T306,T102 | Filter/query, privacy, stale version; keyboard và error states |

## M4 — Payment/notification mock (FR-27–29)

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T401 · DONE | Payment/refund/outbox/log migrations và intent service | T301 | Một pending intent; purpose/full/deposit/balance và amount do server quyết định |
| T402 · DONE | Mock confirm transaction và production route guard | T401,T304 | Confirm cạnh tranh expire/cancel, duplicate; cọc confirmed+partially_paid |
| T403 · DONE | Refund mock/đối soát ledger/admin audit | T402,T306 | Refund cạnh tranh không quá paid; refund không tự trả chỗ |
| T404 · DONE | Outbox lease/retry/dedupe và email/SMS mock preview | T402 | Crash trước/sau log, retry 5 lần, không double notification, không gửi thật |
| T405 · DONE | Payment QR giả/confirmation/notification preview UI | T307,T402,T404 | QR không chuyển tiền thật; không thông báo gửi thật; poll bounded, double-click |
| T406 · DONE | Admin payment/refund/log UI | T308,T403,T404 | Sales không refund; accountant đúng quyền; API lỗi không báo thành công |

## M5 — Sales/khách đoàn (FR-40–45,61,85–88)

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T501 · DONE | Migration/API tiếp nhận yêu cầu, ghi chú, báo giá và tệp đính kèm | T104,T107,T215 | PostgreSQL CI kiểm tra ngày, đồng ý, chống gửi trùng, trạng thái và nhân viên đủ quyền |
| T502 · DONE | Form báo giá và gọi lại, biên nhận theo SLA | T214,T501,T110 | Playwright CI kiểm tra captcha lỗi, bấm gửi lặp, giữ nội dung và SLA từ cấu hình |
| T503 · DONE | Admin phân công, ghi chú, xử lý báo giá và tệp đính kèm | T102,T501 | CI kiểm tra tính bất biến của bản phát hành, phiên bản mới, xác nhận ngoại tuyến và tải tệp riêng tư |
| T504 · DONE | CRM hồ sơ khách, tìm kiếm, lịch sử và ghi chú | T105,T306 | CI kiểm tra quyền, email/SĐT trùng, liên kết cũ và lịch sử sau xác minh |

Bằng chứng M5 trên `b76c2f0` trong [GitHub Actions 45009896743](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45009896743).

## M6 — Reporting/export/analytics (FR-08,48,95–99,INT-04)

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T601 · DONE | Revenue/occupancy/quote metrics API | T403,T501 | Dataset known sums, midnight VN, refund khác ngày, không double JOIN |
| T602 · DONE | Dashboard admin + filter/date/empty/error | T601,T102 | Số liệu khớp API; executive chỉ read; mock label |
| T603 · DONE | Excel/CSV export report/booking/tour | T601,T306,T205 | Đọc file bằng parser + mở thực, formula injection, row cap, private access |
| T604 · DONE | PDF tour và group quote | T210,T503 | Font Việt, page break, totals/version, private quote, renderer chặn SSRF |
| T605 · DONE phần mã | GA4/Search Console integration | T405,T502,T216 | Không PII, purchase một lần và nhãn demo; DebugView khi có tài khoản |

Bằng chứng M6 trên `aeaea662` trong [GitHub Actions 45022045577](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45022045577). T605 còn xác minh thuộc tính Google thật khi có tài khoản.

## M7 — Nghiệm thu và vận hành

| ID | Task và đầu ra | Phụ thuộc | Kiểm chứng hoàn thành |
| --- | --- | --- | --- |
| T701 · DONE | Seed demo có tất cả trạng thái cần test | T404,T501 | Seed idempotent, production chặn seed, không có khách thật |
| T702 · DONE | E2E journey và CI browser tests | T405,T406,T503,T603,T604,T701 | Full/cọc/balance/lookup, manual booking, leads, export và access denied |
| T703 · CI đạt, chờ duyệt thị giác | Taste gate + responsive/accessibility/motion review | T208–T214,T307,T405 | Screenshot/keyboard/reduced-motion, các ngoại lệ ghi rõ; không chỉ screenshot desktop |
| T704 · Đã chuẩn bị, chờ cloud | Demo deploy host/env/proxy/DB/R2/Turnstile | T002,T111,T112,T701 | Migration/health/TLS/origin, cold start/restart, private files; không secret trong bundle |
| T705 · Đã chuẩn bị, chờ deploy và rollback thật | Hoàn thiện CD gắn revision và kiểm tra deployment hoàn tất | T704 · Đã chuẩn bị, chờ cloud | Hook phải deploy SHA đã CI kiểm tra hoặc có cơ chế từ chối revision lệch; chờ host healthy và smoke; rollback diễn tập |
| T706 · Đã chuẩn bị, chờ người thử | UAT với khách 30+, trường/doanh nghiệp và admin | T702,T703,T704,T605 | 11 tiêu chí mục8 MVP có evidence; ghi lỗi và sửa trước duyệt |
| T707 · Runbook và drill CI đạt, chờ xác minh trên cloud | Runbook setup/seed/deploy/backup demo và bàn giao | T706,T705 | Người khác làm theo trên môi trường sạch, ghi remaining production gaps |

Bằng chứng M7 trên `783e6f86` trong [CI 45025910519](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45025910519). Chi tiết phần cần hỗ trợ trong docs/demo-release.md và docs/uat-mvp.md.

T208–T214 nghĩa toàn bộ các ID trong khoảng. T001/T002 đã hoàn tất quyết định demo. T003 đã có poster tham chiếu qua T003a, còn thiếu nội dung đầy đủ/quyền ảnh; không chặn scaffold bằng fixture. Bán thật cần thay và duyệt các thông tin demo.

## Sau MVP

Chỉ tách thành task thực thi khi giai đoạn được giao: P1 real payment/reconciliation, customer auth, staging, 2FA, backup/restore, giám sát và load/security gates; P2 coupon/consent/newsletter/chăm sóc/review/huỷ đổi; P3 roster/contracts/CRM/invoice/đa ngôn ngữ. Theo PLAN, không tự đưa chúng vào MVP.

## Mẫu đóng task

`ID — DONE | ngày | nhánh | commit đã push | file thay đổi | lệnh và kết quả | evidence UI nếu có | hạn chế | task tiếp theo`.
Ghi bản đầy đủ trong PROGRESS; nếu test fail, task vẫn DOING/BLOCKED. Không sửa lịch sử cũ để che kết quả fail.


## Điều chỉnh theo yêu cầu ngày 23/11/2026

| Task | Thay đổi | Kiểm chứng |
| --- | --- | --- |
| T218b · DONE | Xóa Cẩm nang hoàn toàn: route/blog, sitemap, API, seed và migration 000008 xóa dữ liệu cũ | PostgreSQL nâng cấp có dữ liệu/FK và chạy lại, tour/media/pages được giữ; seed idempotent; public/admin API 404; Playwright URL blog 404, sitemap không blog; lint/types/unit/build và CI |


## Nâng cấp giao diện ưu tiên Zalo — 23/11/2026

| Task | Phạm vi | Gate độc lập |
| --- | --- | --- |
| T703l · DONE | Tokens public, Nunito Sans toàn luồng, nút và trạng thái chung | lint/types/unit, không đổi backend |
| T703m · DONE | Hero, Trang chủ, closing CTA, hỗ trợ responsive | build/types, browser hero/Zalo/mobile/video |
| T703n · DONE | Cards, danh sách, detail, contact/group, các trang public | tìm tour, lọc, hết chỗ, checkout/form không hồi quy |
| T703o · TODO | Admin sidebar, login, list/editor/report/status | lint/types/unit/build và browser permissions/menu/form/export |
| T703p · DOING | contact_click/select_item, consent và QA thị giác | consent/PII, 320–1440px, keyboard/reduced motion và CI xanh |
