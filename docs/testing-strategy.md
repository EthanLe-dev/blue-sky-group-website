# Chiến lược kiểm thử và tiêu chuẩn chất lượng

Phiên bản 0.1 · 22/11/2026. Áp dụng public Next.js, React admin, Go/Gin, PostgreSQL và tích hợp mock/real. Tham chiếu [TASKS](../agent/TASKS.md), [API](api-spec.md), [schema](database-schema.md), [design](design-style.md). Các công cụ/gates dưới đây là yêu cầu triển khai, không tuyên bố đã có test suite.

## 1. Môi trường và dữ liệu

- Local/CI dùng PostgreSQL cùng major17, database riêng do test runner quản lý. `TEST_DATABASE_URL` chỉ trỏ DB test; guard tên DB/môi trường trước reset. Không dùng SQLite thay integration PostgreSQL, không đụng DB demo/production.
- Fixture deterministic có seed ID riêng: adult/child/infant, birthday boundary, leap day, phụ thu, cọc/full, hết/sắp hết chỗ, confirmed/expired/cancelled/completed, refund, nhiều passenger và nhiều vai trò.
- Dữ liệu giả, email `.example`/example.com, không dữ liệu khách thật. Mỗi test/suite cô lập schema/transaction; concurrency tests dùng kết nối DB độc lập và barrier để cạnh tranh thực, không bao hết trong cùng transaction test.
- Mock chỉ ranh giới dịch vụ email/SMS/payment/object storage ngoài; kiểm DB/HTTP/service thật khi integration. Dùng fake clock cho unit, nhưng expiry transaction cần test giờ DB và chờ lock thực. Tránh sleep cố định gây flaky.
- Demo công khai kiểm Turnstile thật/hostname, cookie/proxy, R2 thực. Không gọi thu tiền hoặc gửi tin thật trong CI. Sandbox production dùng account test riêng.

## 2. Các lớp kiểm thử

| Lớp | Công cụ đề xuất khi scaffold | Nội dung / thời điểm |
| --- | --- | --- |
| Static | ESLint/TypeScript, gofmt/go vet | Mỗi PR: style, types, unused code, lỗi Go |
| Unit | Go testing; Vitest + Testing Library | Giá, tuổi, permission, transitions, validation; frontend theo hành vi người dùng |
| Integration | Go test + PostgreSQL service | Transactions, migrations, constraints, session, ledger/counters, outbox; mỗi thay đổi backend |
| Contract | OpenAPI validator + HTTP tests | Request/response/envelope/status, pagination, money string, dates/enums, field privacy |
| E2E | Playwright | Browser → API → DB thật, journey quan trọng; không mock endpoint đang nghiệm thu |
| UI/accessibility | axe + keyboard/manual + taste scripts | Responsive, font Việt, focus, contrast, reduced motion, loading/error/empty |
| Performance | Lighthouse/Web Vitals, browser trace; k6 trước production | LCP/CLS lab, field INP khi có dữ liệu, API p95 và concurrency |
| Security | Dependency/secret scan, negative API tests; review/pentest trước live | Auth/IDOR/CSRF/XSS/upload/injection, không lộ secret/PII |

Không dùng snapshot lớn thay assertion hành vi. Không yêu cầu coverage100% để khuyến khích test vô ích; mọi nhánh luật tiền/chỗ/quyền/trạng thái có case normal/boundary/error. Báo coverage để tìm vùng thiếu, không coi tỷ lệ cao là chứng minh an toàn.

## 3. Ma trận case bắt buộc

| ID / yêu cầu | Kịch bản | Kết quả phải chứng minh |
| --- | --- | --- |
| Q01 FR-03–06 | Lọc ngày+budget+travellers, draft, giá ngày khác rẻ hơn | Tất cả điều kiện cùng departure; draft không lộ; sort/pagination ổn định |
| Q02 FR-22/28 | Tuổi tại ngày đi, phụ thu mandatory/optional, cọc làm tròn, tour cận ngày | Money nguyên đồng đúng; chỉ full khi qua due date; từ chối giá client |
| Q03 FR-24/25 | 20 request độc lập tranh5 chỗ, mỗi request1 chỗ | Đúng5 thành công nếu hợp lệ,15 sold-out; counters không âm/vượt; ledger/hold khớp |
| Q04 FR-24/25 | Request web và admin cùng đặt, inject lỗi sau tăng held trước commit | Cùng kho, rollback không rò hold/booking; không chỉ test tổng HTTP200 |
| Q05 FR-24 | Hold ngay trước/đúng/sau expiry; worker chạy2 lần; restart/ngủ | Chỉ release1 lần; confirm đọc giờ sau lock; request mới giải phóng hold quá hạn |
| Q06 FR-20/27 | Cùng key/cùng payload, cùng key/khác payload, mất response sau commit | Một tài nguyên; replay cùng kết quả/409 đúng; không nhân đơn/tiền |
| Q07 FR-27/28 | Full, deposit, balance; confirm cạnh tranh expiry/cancel | Một chuyển trạng thái hợp lệ; cọc partially_paid; không thu dư/chỗ âm |
| Q08 FR-83 | Cancel/refund lặp và concurrent, completed, giảm capacity | Không hoàn chỗ/tiền hai lần; refund không tự cancel; capacity giữ invariant |
| Q09 FR-29 | Worker crash trước/sau notification log, lease timeout, retry hết | Dedupe log, retry bounded/failed có thể xem, không gửi thật |
| Q10 FR-31/NFR-24 | Guest A đọc booking B; lookup token gọi payment/preview/passenger | 404/403; chỉ summary che PII; contact sai không tiết lộ field đúng |
| Q11 FR-89 | 5 vai trò gọi từng endpoint với/không quyền; JWT expired/revoked | 401/403 chuẩn, kiểm object; frontend ẩn nút không thay test backend |
| Q12 NFR-21/23/26 | Refresh reuse, CSRF/Origin sai, captcha fail, brute force/restart | Revoke family, từ chối mutation, rate-limit còn tác dụng sau restart |
| Q13 FR-80–88 | Publish thiếu policy/price, bulk departure lỗi1 dòng, quote issued sửa | Fail rõ; rollback cả bulk; quote immutable/version mới |
| Q14 FR-95–99 | Nhiều passenger/payments/refunds, qua nửa đêm VN, cohort quote | Tổng known fixture đúng; không double count JOIN; refund theo ngày riêng |
| Q15 FR-08/48/99 | Export có dấu, rất dài, hàng bắt đầu =/+/-/@, user không quyền | File mở được, font/page break đúng, không formula execution; không private leak |
| Q16 NFR-21 | SQL/XSS payload, file sai magic/size/path, SSRF qua PDF | Không execute/inject; reject đúng; renderer không đọc mạng nội bộ/file máy |
| Q17 INT-04 | Reload confirmation, retry purchase, PII trong form | Event không duplicate theo transaction, không contact/passenger/token trong analytics |
| Q18 UX-01–08/12–16 | 320–1440px, zoom200%, font chậm/lỗi, reduced-motion, keyboard | Không cắt dấu/overflow/che CTA; focus và nội dung dùng được khi motion tắt |

Sau các case tranh chỗ, query DB kiểm lại held bằng tổng active holds sau cleanup, confirmed bằng tổng seat_count confirmed/completed, paid/refunded bằng ledger succeeded. Không chỉ kiểm response. Thử race nhiều lượt với barrier có timeout; deadlock/timeout phải có xử lý rõ, không bỏ lỗi test.

## 4. E2E và nghiệm thu MVP

Journey bắt buộc: duyệt/filter → detail → quote → booking → full mock → confirmation → lookup; journey deposit → balance; expired hold/price changed/sold-out; admin manual booking/cancel/refund; group request → assignment → quote/PDF; callback; CMS publish; report Excel/CSV; role forbidden. Dùng UI và DB thật; preview email/SMS đúng nhãn giả lập.

Kiểm desktop Chromium, Firefox, WebKit tự động trước demo; WebKit không thay Safari/iOS thật. Nghiệm thu thủ công hai phiên bản gần nhất Chrome/Safari/Edge/Firefox, iOS Safari/Android Chrome theo thiết bị có sẵn; ghi phiên bản đã chạy và gap, không ghi “cross-browser pass” khi chỉ chạy Chromium. Viewport1440×900,1024×768,820×1180,768×1024,390×844 và320px. Kiểm touch, back/forward, mất mạng, retry, font swap, sticky bar/safe area.

Checklist11 tiêu chí mục8 MVP cần evidence riêng: trang responsive; checkout; mock QR; mock email/SMS; Excel; PDF; map; GA4; admin/RBAC; Turnstile; URL demo. Thiếu tài khoản analytics/host không được đánh dấu pass bằng mock. UAT có đại diện khách30+, trường/doanh nghiệp, sales/kế toán; ghi tác vụ, kết quả tự hoàn thành, lỗi quan sát và quyết định duyệt.

## 5. Gate theo giai đoạn

| Gate | Điều kiện |
| --- | --- |
| Hiện tại — chỉ tài liệu | Markdown/link/fence, git diff --check, Compose/YAML; chưa test app. CI GitHub f80a064 đã pass là bằng chứng cấu hình, không bằng chứng nghiệp vụ. |
| Mỗi PR khi có app | npm ci/lint/typecheck/test:ci/build; gofmt/vet/test -race/build, migrations sạch/contract/integration liên quan; không secret hoặc test failing bị che bằng skip |
| Trước merge thay đổi core | Q02–Q12 liên quan + regression tiền/chỗ; review schema/migration/rollback; task có evidence |
| Trước demo | Journey E2E +11 tiêu chí, responsive/accessibility/taste, Turnstile/proxy thực, PDF/Excel, không lỗi P0/P1; mọi skip/blocker ghi rõ |
| Trước production | Tất cả Must giai đoạn, sandbox payment callbacks/reconcile/late callback, pentest không còn critical/high, restore drill, load test/monitoring/SLA và rollback |

P0: mất/lộ dữ liệu, thu sai tiền, overbooking; P1: luồng chính bị chặn hoặc quyền sai; P2: chức năng phụ/lỗi UX có workaround; P3: thẩm mỹ nhỏ. P0/P1 chặn release, P2 cần chủ sản phẩm chấp nhận có hạn sửa, P3 vào backlog. Flaky test ghi issue/nguyên nhân, không xoá assertion hoặc tự retry đến xanh để coi hoàn thành. Một retry CI phục vụ chẩn đoán phải giữ trace lần fail.

## 6. Hiệu năng và vận hành

MVP kiểm budget media/trace, không bắt load test production trên host free. Trước production: tối thiểu100 phiên hoạt động với hành vi duyệt/tìm/đặt, không diễn giải thành100 request/s. Kịch bản đề xuất ramp5phút, steady15phút, cooldown5phút; phân bố70% đọc,20% search/detail,10% booking/payment sandbox; dataset đủ tour/chỗ để phân biệt lỗi hệ thống với sold-out dự kiến. Đo riêng request hợp lệ, tỷ lệ5xx/timeouts <1%, correctness không lỗi; báo vùng, cấu hình, phiên bản và cold/warm runs. Thử200–300 phiên đánh giá dư địa, không che kết quả suy giảm.

Ngưỡng NFR: API thường p95≤500ms, booking/hold p95≤2s, server page p95≤2s; mobile LCP≤2.5s, CLS≤0.1, INP≤200ms (field p75 khi đủ mẫu). Lighthouse là tín hiệu lab, không thay INP field. Animation mục tiêu gần60fps, kiểm dropped frames/long tasks trên thiết bị đã ghi; không chặn tương tác vì transition.

Production kiểm backup hằng ngày, restore sang DB riêng, RPO≤24h/RTO≤4h theo mục tiêu PLAN; payment cần mức bảo vệ riêng được chốt. Diễn tập deploy fail/rollback và migration tương thích. Uptime99.5% là mục tiêu đo qua giám sát, không thể nghiệm thu chỉ bằng một lần ping.

## 7. CI/CD và lưu bằng chứng

CI hiện có app jobs tự bật theo manifest. T112/T702 cần bổ sung E2E/contract/migration checks khi triển khai; không coi các bước chưa tồn tại là đã chạy. Deploy workflow hiện chỉ gửi hook, chưa bảo đảm host deploy đúng SHA hoặc hoàn tất. T705 phải pin revision/kiểm revision, chờ provider readiness và smoke trước đánh dấu deployment thành công; checkout SHA trong Actions tự nó không pin được deploy hook.

Khi có test artifact, lưu report JUnit/coverage/Playwright trace/screenshot/log redacted vào CI artifacts, đề xuất14ngày (30ngày cho release evidence); không commit PII/video quá lớn. PROGRESS ghi ID task, commit, môi trường/tool versions, lệnh, exit/result, URL/path evidence, failures/waivers và phần chưa chạy. Kết quả tài liệu và kết quả ứng dụng luôn phân biệt rõ.
