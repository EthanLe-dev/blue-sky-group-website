# TIẾN ĐỘ DỰ ÁN BLUE SKY GROUP WEBSITE

## Bàn giao nhanh — 22/11/2026

- Đã cài `bluesky run` mở/build app và `bluesky off` dừng/giữ volume trên máy người dùng. Mở PowerShell mới, dùng từ bất kỳ thư mục; không khởi động app khi cài.

- Cập nhật mới nhất: người dùng yêu cầu tự chạy app. Đã dừng toàn bộ Compose của dự án bằng down (không -v); dữ liệu volume giữ nguyên. Agent không tự khởi động lại app/Docker/service local nếu chưa được yêu cầu rõ. Hướng dẫn chạy trong README.
- Giai đoạn: M0 inventory/wireframe đã đủ bằng dữ liệu demo; M1 có ứng dụng nền tảng chạy local Docker. Chưa có catalog/booking/payment hoàn chỉnh (M2–M4).
- Nhánh hiện tại `develop`. PR #1 đã qua CI và merge tại ce0bbb9; chứa cả tài liệu quyết định, frontend/GSAP/ponytail và nền tảng M0/M1. Main giữ nguyên f80a064.
- Chạy theo README: .env đã được tạo local bằng secret ngẫu nhiên, không commit; Docker/PostgreSQL/API/Next/Admin đang chạy, admin http://localhost:5173, API http://localhost:8080, public http://localhost:3000 (scaffold). Credential bootstrap nằm trong .env.
- Đã kiểm: backend vet/race/integration/OpenAPI, Admin4 unit tests+lint/types/build, frontend smoke/lint/types/build, Docker build/health, hai browser E2E, local backup/restore DB test. CI remote của PR #1 đã PASS tất cả jobs (run44953203198).
- T108b lookup token cần FK booking nên chuyển sang M3 sau T301; T108a guest hoàn tất. T113 default snapshot vào đơn sẽ kiểm ở T302. Công ty upload giấy phép chọn media chờ T215. Turnstile real adapter có, local dùng mock; chưa nghiệm thu siteverify với tài khoản thật.
- Tiếp theo sau M1: T201/T202 và T203 catalog; dùng poster Ninh Chữ + inventory demo, mọi dữ liệu thiếu chỉnh ở Admin theo kế hoạch. Đọc TASKS/PLAN/API/schema/design/testing trước sửa.

## Lịch sử thực hiện

### 22/11/2026 — Chuẩn bị thiết kế kỹ thuật

Đã đọc final/MVP/architecture, chuyển hướng dẫn Agent thành AGENTS.md ở root; soạn PLAN, database-schema, api-spec, Compose, env example, gitignore và CI. Thiết kế transaction PostgreSQL chống overbooking; tách confirmed/partially_paid; lazy expiry khi host ngủ; mock chỉ local/demo. Kiểm tra Compose/YAML/env và link tài liệu đạt trong phiên thực hiện. Đây là thiết kế, chưa có migration/code.

### 22/11/2026 — Khởi tạo Git và CI/CD

Đã khởi tạo Git local, push commit `f80a064`, tracking origin/main; xác minh GitHub default_branch=main và remote HEAD trỏ main. CI: https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/44945208713 — success. Deploy workflow: https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/44945247863 — success nhưng các bước deploy skipped vì chưa có ứng dụng.

CD chỉ gọi hook Render/Vercel khi có đủ app, cần secrets trong environment demo. Chưa tạo/cấu hình project host hay chạy deploy thật. Lưu ý cần xử lý T705: hook có thể deploy HEAD hiện tại của provider; việc Actions checkout đúng SHA chưa đảm bảo revision host. Cần pin/verify revision và readiness trước nghiệm thu CD đầy đủ.

### 22/11/2026 — Chia task, thiết kế UI và chiến lược kiểm thử

- Yêu cầu: điền TASKS.md, design-style.md, testing-strategy.md, cập nhật PROGRESS để tiếp tục ở phiên mới.
- Đã đọc PLAN và taste-skill 2.5.1; các file TASKS/design/testing ban đầu trống, PROGRESS có ghi chú chuẩn bị cũ được giữ bên dưới.
- TASKS: task ID ổn định theo M0–M7, dependency, đầu ra và cách kiểm chứng riêng. Tất cả task triển khai ở TODO, không giả định có ứng dụng.
- Design: typography tiếng Việt, tokens đề xuất, layout public/admin/checkout, motion timing/reduced motion, quy trình taste và ngoại lệ theo yêu cầu dự án.
- Testing: unit/integration/contract/E2E, case tiền/chỗ/race/RBAC/PII, quality gates demo/production, accessibility/browser/performance/restore và evidence.
- Kiểm chứng của phiên: rà liên kết nội bộ, dependency task, Markdown fences và git diff whitespace; kết quả được xác nhận khi hoàn tất lượt sửa này. Không chạy build/test app hoặc taste gate vì chưa có trang ứng dụng.
- Bước tiếp: đọc backlog, chọn T101 hoặc T103; tạo nhánh module từ `develop`, ghi task DOING trước thay đổi, cập nhật PLAN/PROGRESS khi đóng task. Chưa yêu cầu thực thi scaffold trong phiên tài liệu này.

### 22/11/2026 — Thiết lập nhánh develop và quy trình Git

- Người dùng yêu cầu `develop` là nhánh code chính; `main` chỉ nhận code đã chạy tốt. CI trên develop và main phải giống nhau. Mỗi module lớn có nhánh riêng; xong task nào commit/push task đó để rollback rõ.
- Từ `main` đã tạo và push `develop`. Năm file tài liệu hoàn tất từ lượt trước được commit `65d0708 docs: define implementation tasks and quality standards` và push trên `develop`.
- Đã bổ sung quy tắc nhánh/PR/commit/push vào AGENTS, TASKS, PLAN và PROGRESS. CI hiện có trigger push `main, develop` và `pull_request` chung jobs; deploy workflow chỉ theo push `main` sau CI.
- Cần xác nhận sau khi push commit quy trình: remote develop HEAD, CI run develop. Branch protection/required checks trên GitHub chưa được cấu hình, nên điều kiện review/merge hiện là quy trình ghi trong tài liệu, chưa phải ràng buộc server.

## Quy tắc cập nhật cho phiên sau

Giữ lịch sử theo thời gian, thêm entry mới, không sửa kết quả cũ thành pass. Cập nhật “Bàn giao nhanh” mỗi phiên. Chỉ ghi sự kiện có bằng chứng; ghi rõ proposed/implemented/verified/deployed. Nếu bị gián đoạn, lưu task đang làm, file đã đổi, lệnh cuối, lỗi còn tồn tại, dữ liệu/process cần giữ và bước chạy tiếp an toàn. Không ghi password/token/PII vào nhật ký.

Mẫu entry:

```text
Ngày/giờ Việt Nam — task ID — trạng thái
Mục tiêu và phạm vi:
Đầu vào / quyết định mới:
File đã thay đổi / nhánh / commit đã push hoặc working tree:
Lệnh kiểm tra + kết quả thật + môi trường:
Evidence (CI/artifact/screenshot):
Phần chưa kiểm chứng / lỗi còn mở:
Blocker và điều kiện gỡ:
Bước tiếp theo cụ thể:
```

## Ghi chú gốc — trước khi khởi tạo Git

Cập nhật lần cuối: 22/11/2026

## 1. Trạng thái Tài liệu
- [x] **`final-requirements.md`**: Tài liệu đặc tả yêu cầu kinh doanh và hệ thống bản hoàn chỉnh. (Đã tách kiến trúc sang file riêng).
- [x] **`mvp-requirements.md`**: Đặc tả phạm vi bản MVP demo (Giao diện đầy đủ, thanh toán QR giả lập, email/SMS giả lập, xuất Excel/PDF, bản đồ, analytics).
- [x] **`architecture.md`**: Thiết kế kiến trúc hệ thống chuẩn (Tech stack, Modular Monolith, sơ đồ triển khai Vercel + Cloudflare, cơ chế lock chỗ, cấu trúc thư mục).
- [x] **`database-schema.md`**: Thiết kế cơ sở dữ liệu quan hệ PostgreSQL chi tiết cho toàn bộ các module.

## 2. Kế hoạch Thực hiện Tiếp theo
- [ ] Khởi tạo khung dự án theo cấu trúc đề xuất (`frontend`, `admin`, `backend`, `docker-compose.yml`, `nginx`).
- [ ] Xây dựng Database Migration scripts cho PostgreSQL.
- [ ] Xây dựng khung Backend Go + Gin với kiến trúc Modular Monolith.
- [ ] Dựng giao diện Frontend Next.js (Khách hàng) và React (Admin).

### 22/11/2026 — T001/T002: Chốt demo, Admin settings và poster

Người dùng cho phép giả định giá/đặt chỗ/thông tin công ty, bắt buộc sửa ở Admin; tour có poster từ team thiết kế; chọn Render+Neon nếu miễn phí đủ demo; không deadline. Đã cập nhật PLAN, requirements, architecture, schema/API, design/testing, AGENTS; thêm T113/T114 cho cấu hình và T217/T218 cho poster (tổng66 task). Kiểm tra nguồn chính thức gói miễn phí; năng lực demo vẫn cần T704 smoke test.

Ảnh poster chưa có trong chat/tool/workspace; workspace chỉ có logo. Không khẳng định đã phân tích ảnh mẫu. Cần người dùng đính kèm lại, chưa chặn scaffold. Nhánh `docs/admin-settings-and-posters` từ develop; task tài liệu T001/T002 có commit riêng cho quyết định Admin và hạ tầng, chưa có code/Admin thật. Kiểm links/fences/IDs và git diff --check trước push; chưa build/test app. Bước tiếp: T101/T103; settings và poster triển khai theo dependency, không hardcode chính sách demo.

### 22/11/2026 — T003a: Nhận poster mẫu

Đã xem ảnh người dùng gửi, lưu nguyên bản vào docs/assets/ninh-chu-poster-reference.png (1051×1497px). Ghi thông tin đọc từ poster và giới hạn xác minh trong docs/poster-reference.md; cập nhật design-style/PLAN/AGENTS/TASKS. T003a DONE, T003 chưa DONE vì ảnh không có giá theo đối tượng/ngày khởi hành/chỗ/chính sách đầy đủ. Thông tin liên hệ trên ảnh chưa xác nhận hiện hành, không tự seed thành dữ liệu thật.

Kiểm chứng: dimensions qua Pillow, bản sao SHA-256 bằng file đính kèm, Markdown links và git diff --check. Chưa có UI để chạy browser/gate. Commit/push trên nhánh docs/admin-settings-and-posters; chưa merge develop. Bước tiếp: scaffold T101/T103, dùng poster làm tham chiếu và fixture có nhãn theo nhu cầu.

### 22/11/2026 — T101: Next.js scaffold và GSAP

Người dùng yêu cầu cài GSAP và bắt buộc Agent dùng cho animation. Trên nhánh `feature/public-ui` từ develop, đã scaffold `frontend/` Next.js App Router + TypeScript, cài `gsap` và `@gsap/react`, thêm scripts CI và smoke test timeline. Cập nhật AGENTS/design/TASKS/PLAN. Chưa dựng website hoặc xác nhận UI/taste gate. Kết quả lệnh lint/typecheck/test/build/CI sẽ ghi sau khi hoàn tất. Lưu ý nhánh docs/admin-settings-and-posters chứa quyết định poster/Admin riêng, chưa merge vào develop khi bắt đầu nhánh này.

T101 verification: Node24.16/npm11.13 local, `npm ci`, `npm run lint`, `npm run typecheck`, `npm run test:ci` (1 pass), `npm run build` đều exit0 sau sửa test sang ESM. Next16.4.6, GSAP3.15.1. UI mới là scaffold “Hello world”, chưa triển khai animation thật/taste gate; T208/T703 mở. Commit/push feature/public-ui.

### 22/11/2026 — Quy tắc dùng ponytail

Người dùng yêu cầu Agent dùng skill `ponytail` khi xây dựng để tiết kiệm token. Đã thêm quy tắc vào AGENTS.md: mức full mặc định, chọn thay đổi nhỏ nhất đủ yêu cầu, tái dùng code và không bỏ các kiểm tra quan trọng. Đây là cập nhật hướng dẫn; chưa có code ứng dụng mới. Kiểm tra `git diff --check`, commit/push trên nhánh feature/public-ui.

### 22/11/2026 — M0 T003/T004

Người dùng cho phép dữ liệu thiếu dùng giả định sửa ở Admin. Inventory demo và page-map wireframe đã lập; poster giữ nguyên, thông tin thiếu có nguồn/trạng thái rõ. Đã hợp nhất nhánh tài liệu và frontend vào develop local, resolve bằng cách giữ cả quyết định poster/Admin lẫn GSAP/ponytail. Nhánh thực thi hiện tại feature/platform. M0 hoàn tất ở mức đầu vào/wireframe; UI/CRUD thực thi vẫn ở task sau. Docker Desktop khởi động được bằng CLI; Go sẽ chạy trong container.

### 22/11/2026 — T102 Admin scaffold

React/Vite/TypeScript Admin có route login có nhãn, giữ dữ liệu khi sai mật khẩu; 4 scripts CI và lockfile. npm install, lint/typecheck/test:ci (2 pass)/build đạt local. Chưa có widget Turnstile công khai, đang local mock; tích hợp UI auth ở T110.

### 22/11/2026 — Nền tảng Go M1 T103–T109, T113

Go1.26/Gin/pgx, config validation và log chỉ request ID/method/status (không headers/body), health/ready; runner migration transaction/advisory lock, identity/RBAC/guest/settings seed, bootstrap idempotent không default password; access JWT/session rotation/reuse/logout/reset, CSRF/Origin/rate-limit DB/Turnstile real adapter, settings version/audit. Mock reset preview AES-GCM, quyền admin và audit; không gửi thật. OpenAPI nền tảng và response validation dùng kin-openapi. Các module chưa có giữ Markdown contract; chưa tạo bảng booking trước M3.

`go vet`, `go test -race -count=1`, `go build` đạt trong container Go với PostgreSQL17; tests tạo schema riêng trong blue_sky_test, từ chối DB không hậu tố_test. Case: giá trị cấu hình ngoài giới hạn, stale409/audit, role403, unknown fields, origin/csrf sai, token reuse/revoke/expiry, reset một lần, limiter qua restart, migration chạy lại/FK và rollback DDL, response OpenAPI. Không có test ledger/booking vì chưa triển khai M3/M4. Các phần core liên kết chặt được kiểm chứng cùng một snapshot nền tảng; task UI/client/container có commit riêng tiếp theo.

T108 được tách T108a (guest) đã làm và T108b (lookup token) sang M3 vì phụ thuộc FK booking T301; không đánh dấu lookup hoàn tất giả. T113 xử lý default settings; áp dụng snapshot lên quote/booking sẽ kiểm lại ở T302. T114 đang form theo trường; file giấy phép đính kèm chờ media T215.

T104 restore drill: `python scripts/test_database_restore.py` PASS. Tạo hai DB test tên ngẫu nhiên, áp dụng migrations, sửa setting, pg_dump/pg_restore, đối chiếu5 roles và version2 rồi dọn đúng hai DB test. Không đụng volume/database ứng dụng. Đây là diễn tập local, không chứng minh RPO/RTO production.

### 22/11/2026 — T110 client/proxy và T114 Admin settings

Client giữ access token trong memory, refresh dedupe/retry401 một lần, restore/logout có CSRF; Next server proxy và Vite/Nginx same-origin. Form Admin sửa hold/cọc/tuổi/chỗ/hạn trả, policy, công ty/liên hệ/SLA với version. Hai E2E trên Docker/PostgreSQL đã PASS: login→save→reload→logout và public proxy cookie/CSRF. Lỗi đã sửa: GET cùng origin thiếu Origin được xác minh qua Referer+Sec-Fetch-Site; Nginx DNS cache cũ gây502 sau recreate được đổi sang resolver Docker. Thư viện media/tài liệu giấy phép đợi T215, không giả upload hoàn tất.

T114: form settings có test save optimistic version và giữ input khi conflict; admin lint/typecheck và4 unit tests PASS. E2E thao tác hold20→21→20/save/reload/logout PASS. Các nhóm cấu hình sửa bằng field label, không bắt người dùng nhập JSON.

### 22/11/2026 — T111 container và local runbook

Dockerfile backend nonroot/distroless, Next standalone nonroot, Admin Nginx unprivileged; profile app có migration riêng trước backend health, public/admin/gateway. `docker compose build`, `up -d --wait`, bootstrap và HTTP readiness/openapi/public/admin đều PASS. init_local tạo secret ngẫu nhiên trong .env ignored, không ghi đè file cũ. Test browser real đã kiểm token/cookie qua proxy. Nginx dùng Docker resolver tránh IP cũ sau recreate. README/runbook ghi lệnh chạy, migration/rollback và giới hạn local/demo. Chưa deploy Render/Neon.

T107 hồi quy browser: GET session-csrf cùng nguồn thiếu Origin giờ yêu cầu Referer allowlist + Sec-Fetch-Site=same-origin; request không xác minh bị403. T103 ready kiểm migration settings mới nhất. Bộ go vet/test-race/build chạy lại PASS; test schema/SQL/restore không dùng DB ứng dụng.

T112 CI lượt đầu44952940828 phát hiện frontend clean-checkout typecheck thiếu global LayoutProps (local đã có .next nên chưa lộ). Sửa layout dùng ReactNode tường minh; thêm typecheck trước Docker build để kiểm môi trường sạch. Admin và foundation đã pass; chưa ghi CI tổng PASS trước khi rerun đạt.

### 22/11/2026 — T112 bàn giao và kiểm tra remote

PR: https://github.com/EthanLe-dev/blue-sky-group-website/pull/1 (feature/platform → develop). Pipeline có foundation, frontend, admin, Go/PostgreSQL, container/browser và aggregate CI result. Đã thấy CI bắt lỗi LayoutProps ở checkout sạch, sửa bằng ReactNode và thêm Docker typecheck trước build. Các lần fail/cancel giữ trong GitHub history, không coi là pass. Commit tài liệu này ghi trạng thái triển khai; merge chỉ sau run của HEAD được GitHub xác nhận success. Xem PR cho kết quả cập nhật, không suy ra từ kết quả local.

Nền tảng local hiện đã chạy được: Admin/login/settings, refresh/revoke/CSRF/RBAC, DB migrations/bootstrap, API proxy, health, Docker. Frontend public còn scaffold, không có tour search/booking; bắt đầu M2 ở T201/T203 tiếp theo. T108b được chuyển M3 do FK booking, company media picker chờ T215; không có deploy thật, không gửi email/SMS/thanh toán thật. Toàn bộ credentials local nằm trong .env ignored. Test Go dùng DB blue_sky_test và schema riêng; restore drill chỉ tạo/dọn hai DB test có tên ngẫu nhiên.

### 22/11/2026 — Đã nghiệm thu CI và merge M0/M1

HEAD13cfcbf của PR #1 đã được GitHub xác nhận success: https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/44953203198. Tất cả jobs foundation, public, admin, Go/PostgreSQL, container/browser và aggregate đều PASS. Đã merge PR #1 vào develop tại ce0bbb9672fa493da4a3a368c73fe70411cbdf96; local develop đồng bộ. Main vẫn f80a064. Bản cập nhật nhật ký sau merge chỉ đổi tài liệu, không đổi source đã kiểm.

Có thể tiếp tục M2: T201/T202 UI tokens/layout, T203 catalog. T108b chỉ thực hiện sau T301. Local Docker vẫn chạy; không cần tạo lại .env hoặc xoá volume. Mở Admin5173 và lấy credential từ .env nếu cần kiểm tra; public3000 hiện là scaffold. Không có deploy demo thật trong phiên này.

### Người dùng yêu cầu dừng app và tự chạy local

Đã chạy `docker compose --profile app --profile production-tools down` không kèm `-v`; `compose ps` không còn service chạy. Không dừng ứng dụng/container khác ngoài dự án. Cập nhật AGENTS để không tự khởi động app/Docker/service local; kiểm thử cần service phải xin phép trước. README có hướng dẫn chạy lại bằng up, xem logs và down; .env/Admin/dữ liệu đã có không cần bootstrap lại. Các ghi chú “Docker vẫn chạy” ở mục lịch sử phía trên là trạng thái cũ.

### Lệnh tắt bluesky run/off

Người dùng yêu cầu lệnh mở/tắt app. Đã thêm scripts/bluesky.cmd và install_shortcuts.ps1, cài launcher vào LOCALAPPDATA/BlueSky/bin, user PATH và refresh PATH trong profile PowerShell7/Windows PowerShell, giữ nội dung/encoding profile cũ. Run kiểm Docker có sẵn rồi up -d --build; off down cả app/Redis không -v. `python scripts/test_shortcuts.py` PASS bằng Docker giả (dispatch run/off, invalid action, daemon lỗi); `bluesky help` và command discovery PASS. Không chạy app hoặc service mới. Nhánh chore/bluesky-shortcuts từ develop, độc lập PR #2.

### Đặt lại mật khẩu Admin local theo yêu cầu người dùng

Đã đổi hash mật khẩu tài khoản bootstrap hiện có theo giá trị tạm người dùng yêu cầu; thu hồi các phiên/token reset cũ và ghi audit. Giá trị chỉ lưu tham chiếu trong .env ignored, không ghi vào Git. Kiểm login qua Admin5173 và API8080 đều200, logout các phiên kiểm tra. Không khởi động/restart app; thao tác trên PostgreSQL và API đã chạy sẵn. Chính sách mật khẩu tối thiểu12bytes của bootstrap/reset API giữ nguyên; tài khoản local hiện tại đã đổi trực tiếp, không chạy lại bootstrap với mật khẩu tạm ngắn.

### Bắt đầu M2/M3 — dùng logo công ty

Người dùng chỉ định blue-sky-group-logo.jpg và cho phép chạy local để kiểm thử rồi dừng khi xong đợt này. Nhánh feature/catalog từ develop có giữ lệnh bluesky. T203 DOING trước, tiếp theo API/catalog/Admin/public và booking trên nhánh riêng. Không đổi mật khẩu Admin hoặc reset volume.

T203: migration000003_catalog tạo media/tours/destinations/categories/departures/prices/quotes/CMS với FK/unique/check tiền/chỗ/ngày. PostgreSQL integration/race/vet PASS, gồm thử capacity âm, ngày kết thúc sai, giữ quá capacity, giá âm đều bị từ chối. Không migration DB ứng dụng trong bước này; test dùng blue_sky_test/schema riêng.

### M2 — API catalog/media, 22/11/2026
- Người dùng cho phép chạy local để kiểm thử rồi dừng; PostgreSQL dùng database test và schema cô lập.
- CRUD/clone/publish tour, bulk departure rollback, price version, bộ lọc cùng departure, CMS sanitize, media magic bytes/private signed URL, poster revision và seed idempotent đã có integration tests.
- `go vet ./...` và `go test -race -count=1 ./...`: PASS (platform 36.348s, catalog PASS).
- M2 đang thực hiện; UI, OpenAPI, kiểm thử trình duyệt còn tiếp tục. Chưa đánh dấu toàn module DONE.

### M3 — Nền tảng transaction, 22/11/2026
- Nhánh feature/booking tạo từ develop và merge dependency feature/catalog trước khi code. M2 UI vẫn đang QA, chưa merge vào develop.
- T301 migration booking/items/passengers/holds/events/notes/idempotency/lookup tokens áp dụng được trên schema PostgreSQL sạch trong integration suite. Không sửa migration đã có.
- T302–T306 kiểm thử đầu tiên PASS (`go test -race -count=1 ./internal/platform -run "TestBooking|TestGuestLookup|TestAdminCancel"`, 11.240s): web/manual tranh 3 chỗ từ 12 request, replay, conflict payload, snapshot defaults, stale price, tuổi sai, sweep hai lần, IDOR/lookup masking/TTL/CSRF, rollback trigger lỗi và cancel lặp. Còn bổ sung biên tuổi/cọc/complete và HTTP contract.

### M3 — Kiểm chứng backend mở rộng
- `go vet` và toàn bộ `go test -race -count=1 ./...` PASS; platform 58.738s. Thêm biên ngày sinh đúng sinh nhật, cọc làm tròn lên, phụ thu bắt buộc/tuỳ chọn, báo giá hết hạn và chỉ full khi quá hạn cọc.
- Complete chỉ nhận chuyến đã kết thúc, giữ confirmed_seats và không tự ghi nhận tiền; cancel/retry không nhả chỗ lần hai.
- T302 đạt gate backend; task khác tiếp tục có bằng chứng riêng.

### M3 — Kiểm thử giao diện
- Playwright 5/5 PASS (11.3s): guest proxy/CSRF, login/settings save-reload/logout, catalog filter/poster Escape/favourites, checkout 3 bước → lookup che PII → admin cancel; tour editor save/reload/verify poster; responsive320/390/768/820/1024/1440 và trang CMS/slug404. Dữ liệu E2E giả; đơn vừa tạo được huỷ qua Admin, nội dung và cấu hình được phục hồi.
- T307 hoàn thành luồng quote/validation/hold/countdown/retry idempotency; PII chỉ trong React memory. Payment CTA disabled có giải thích M4, không ghi thu tiền.
- T308 có lookup chỉ đọc và Admin list/detail/edit/notes/manual/cancel/complete.

### M2 — UI và hợp đồng
- Tour editor/departure bulk/price-version, CMS theo resource, thư viện upload/metadata/private link, company document IDs và đối chiếu poster đã nối API thật. Các trường nâng cao itinerary/stops/prices dùng JSON có hướng dẫn; chưa có visual rich-text/drag-drop editor.
- Public dùng logo thật, Be Vietnam Pro tự host, GSAP scoped entrance/reduced motion, native dialog gallery/menu, filter URL/suggestions, saved/recent/share, CMS/blog/destinations/policies, lazy OpenStreetMap, metadata/sitemap/robots/JSON-LD. Không review giả.
- Docker frontend bổ sung public assets; media_data giữ tệp. .env.example có SITE_URL, Compose truyền env Turnstile runtime.
- OpenAPI M1–M3 validate PASS; bổ sung response validator cho catalog/home/gallery/404. Các test full race gần nhất platform57.587s PASS; kiểm browser5/5 PASS. Taste gate vẫn đang kiểm, T703 chưa đóng.

### Đóng chức năng M2/M3, tách nghiệm thu UI
- Go vet/race toàn suite PASS gần nhất57.587s; OpenAPI validate và M2 response cases PASS; Playwright5/5 PASS. Đã sửa 404 SSR, xác minh poster confirmed:true, metadata/media proxy headers và scope limiter checkout theo guest+IP.
- T201–T218 và T301–T308 đã có chức năng/kiểm chứng; T703 taste/UAT còn mở: `GATE: OPEN 19 (rounds used up)`, chi tiết docs/ui-qa-m2-m3.md. Không gắn nhãn hoàn tất nghiệm thu giao diện.
- PR #3 feature/booking→develop gồm dependency feature/catalog; chờ CI trước merge. Main không thay đổi; không deploy cloud.

### T306 — Ma trận vai trò
- TestBookingRoleMatrix PASS race3.771s: admin đọc/tạo quote/catalog; sales đọc/tạo quote nhưng không sửa catalog; product_manager có catalog nhưng bị chặn booking; accountant chỉ đọc booking; executive không có booking PII. Backend kiểm quyền trực tiếp từng route.

### T304 — Lazy expiry khi Admin sửa lịch
- Dùng cùng ExpireLocked trong transaction đã khoá departure; Admin giảm capacity không bị giữ bởi hold đã hết hạn khi worker chưa sweep. TestCapacityEditExpiresHoldsWhileLocked PASS, cùng bộ booking/race26.835s.
- Go vet và full race M2/M3 gần nhất68.957s PASS, gồm related posts không lộ draft và tag filter.

### T305 — Lookup và giới hạn theo mã
- Lookup token vẫn chỉ lưu hash/TTL15 phút, che contact và không cấp creator quyền. Bucket lookup dùng IP+mã đơn đã chuẩn hoá; không khoá nhầm mã đơn khác qua cùng proxy. TestLookupLimiterIsBoundToBookingCode và GuestLookup privacy PASS race7.368s. Không lưu contact/token vào URL hoặc localStorage.

T304: sau khi tách test inventory sang file riêng, bước compile báo thiếu import booking; đã bổ sung import. Không thay logic transaction; kiểm tra lại sau chỉnh file trước merge.

T304 compile sau tách file đã kiểm lại PASS; capacity/lookup/related-posts/public-image-cover tests PASS race12.330s.

M2 bổ sung theo MVP: Leaflet tải sau click với markers/attribution/cleanup; home search; blog category/tag/related approved-only; gallery kind/collection + viewer; tour advanced facets; phụ thu optional web/manual. Browser6/6 PASS15.0s, gồm tile mock để CI không phụ thuộc OSM network.

### Bản tích hợp cuối M2/M3
- Commit theo task đã push trên feature/booking, chứa dependency feature/catalog. Những sửa liên kết/kiểm thử của T304 có commit riêng bổ sung; không force-push hoặc sửa lịch sử để giấu lỗi compile đã ghi.
- Full Go vet/race M2/M3 PASS68.957s; targeted sau đó lookup/privacy7.368s, capacity/related-post/image-cover12.330s PASS. Browser6/6 PASS15.0s. Lint/typecheck/unit/build đều đã chạy, kiểm lại CI trên SHA cuối tại PR#3 trước merge.
- Perf0 FAIL4 WARN sau tối ưu ảnh, chi tiết docs/ui-qa-m2-m3.md. Nghiệm thu taste vẫn OPEN19 từ vòng3, T703 chưa DONE và chưa UAT thiết bị thật.
- Chưa triển khai M4/M5/M6, chưa deploy cloud. Seed/local giữ mật khẩu hiện tại, volume và sửa đổi người dùng. Khi bàn giao dừng toàn bộ service kiểm thử, không dùng down -v. Trạng thái CI/merge thực tế xem PR#3; main giữ nguyên.

T215 log privacy: Nginx access log chỉ ghi method/uri không query, media location tắt error log có thể chứa signed URL khi upstream lỗi. Backend không log body/query/token. Kiểm nginx -t và query redaction trong vòng kiểm thử cuối.

### Bàn giao kiểm chứng cuối

- Code0ee97b82 CI44998866529 **PASS toàn bộ6 jobs**: docs/config, public quality, admin quality, Go+PostgreSQL, container/browser, aggregate CI result.
- Local trên bản tích hợp cuối: `go vet ./... && go test -race -count=1 ./... && go build ./...` PASS (platform80.297s); lint/typecheck frontend/admin, frontend1unit/admin4unit và Docker builds PASS; Playwright6/6 PASS13.6s. Một warning stylesheet vendor đã ghi rõ, không tắt check để báo xanh.
- `nginx -t` cả hai proxy PASS; request signed query giả qua8088/5173 không xuất hiện trong log. Không in/log token thật.
- Đã dừng bằng `docker compose --profile app --profile production-tools down` (không -v), `compose ps` không còn service. Không khởi động lại sau bàn giao; người dùng tự `bluesky run` / `bluesky off`.
- PR#3 chứa cả dependency catalog và booking, dùng develop làm nhánh tích hợp. Kết quả merge/SHA cuối xem PR trực tiếp; main không nhận đợt này. Commit bàn giao docs không đổi code đã kiểm chứng, vẫn chờ CI theo quy trình.

### M4 — Thanh toán và thông báo giả lập

- Nhánh `feature/payment`, PR #4 vào `develop`. Thêm migration ledger, intent thanh toán, xác nhận toàn bộ/cọc/phần còn lại, hoàn tiền, đối soát từng đơn, worker và nhật ký xem trước email/SMS.
- Dùng QR chứa dòng chữ demo. Bản xem trước giữ mã đơn và nội dung mẫu, không gọi nhà cung cấp. Người tạo đơn có thể mở lại bằng liên kết trên cùng phiên trình duyệt.
- CI bản đầu `821b345` đạt 6/6 job tại run 45005020537. PostgreSQL kiểm tra cọc, phần còn lại, xác nhận lặp, không thu khi hết hạn, quyền sales/kế toán, hoàn cạnh tranh và lease/retry. Playwright kiểm tra QR, trạng thái đã thu, preview, tra cứu, hủy và hoàn tiền.
- Bản cập nhật thêm kiểm tra confirm cạnh tranh hủy, mở lại đơn cùng phiên, chặn bấm xác nhận hai lần và nhật ký toàn hệ thống. CI PR là bằng chứng kiểm tra của từng bản. Không khởi động app trên máy người dùng trong lần làm M4.

### T501 — API tiếp nhận và xử lý yêu cầu

- `feature/sales` xuất phát từ `develop` tại bản hợp nhất PR #4 (`ff59d0f`). Migration `000006_sales.sql` thêm yêu cầu đoàn, gọi lại, ghi chú, phiên bản báo giá và tệp đính kèm riêng tư.
- API dùng phiên khách, CSRF, giới hạn gửi và Turnstile. Gửi lại cùng khóa trả biên nhận đã lưu. Admin chỉ giao việc cho nhân viên đủ quyền đang hoạt động. Báo giá đã phát hành giữ nguyên nội dung và kết quả trao đổi được ghi nhận trong Admin.
- [CI 45008730188](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45008730188) đạt 6/6 job trên `bac6332`, gồm Go race/PostgreSQL và container/browser. Bộ test Go toàn bộ gặp giới hạn socket loopback của sandbox ở test R2 khi chạy local.

### T502 — Biểu mẫu tư vấn công khai

- Form báo giá trên trang trường học và doanh nghiệp, form gọi lại trên trang liên hệ, có lối vào từ trang tour. Biên nhận hiển thị mã và thời gian phản hồi lấy từ Cài đặt Admin. Biểu mẫu giữ nội dung khi gửi lỗi và giữ cùng khóa khi thử lại dữ liệu cũ.
- Lint frontend qua với một cảnh báo stylesheet có sẵn, TypeScript và unit test đạt. Turbopack local dừng ở tiến trình xử lý stylesheet Leaflet. CI kiểm tra build và luồng Playwright gồm captcha lỗi, giữ nội dung, bấm gửi hai lần và biên nhận.

### T503 — Hàng chờ Sales và báo giá đoàn

- Admin có hàng chờ báo giá và gọi lại, bộ lọc trạng thái/người phụ trách, hạn theo dõi theo giờ Việt Nam, ghi chú nội bộ và thông báo quá hạn trong danh sách.
- Nhân viên có thể tạo và sửa bản nháp, phát hành, tạo phiên bản từ báo giá cũ và ghi nhận kết quả trao đổi ngoại tuyến. Tệp đính kèm riêng tư dùng lưu trữ media hiện có và liên kết có hạn.
- Admin lint, TypeScript, 4 unit test và build đạt. Playwright đã nhận diện kịch bản phân công, phát hành, chốt báo giá và kiểm tra tải tệp; chạy kịch bản trên CI.

### T504 — CRM và lịch sử đơn đã xác minh

- Admin cho phép tạo, tìm và sửa hồ sơ khách cùng ghi chú. Hồ sơ có thể dùng cùng email hoặc số điện thoại. Lịch sử yêu cầu quyền đọc đơn và chỉ hiện liên kết đã được nhân viên xác minh.
- Nhân viên đối chiếu đúng mã và thông tin liên hệ của đơn trước khi xác nhận gắn. Liên kết giữ dấu thời gian và người xác minh, đồng thời ghi sự kiện trong cùng giao dịch. Liên kết cũ chờ xác minh trước khi hiện trong CRM.
- Go kiểm tra trùng thông tin liên hệ, quyền, liên kết trước M5, phiên bản cũ và gắn chéo hồ sơ. Kịch bản browser tạo hồ sơ trùng email với đơn và kiểm tra đơn chỉ hiện sau bước xác nhận. Kết quả tích hợp lấy từ CI.

### Kiểm chứng M5

- Nhánh `feature/sales` đã push commit theo task. T501 `bac6332`, T502 `50115ba`, T503 `716180d`, T504 `9170c07`. Commit `b76c2f0` bổ sung chốt chuyển trạng thái theo hợp đồng API và kiểm tra từ chối đổi sang “Không chốt” trước khi phát hành báo giá.
- CI bản đầy đủ `9170c07` và bản chốt `b76c2f0` đều đạt 6/6 job. Bản chốt kiểm tra build public/Admin, Go vet/race với PostgreSQL, container và 8 kịch bản Chromium. [CI bản đầy đủ](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45009350708) · [CI bản chốt](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45009896743)
- Kiểm thử tích hợp chạy trên CI. Tệp báo giá có thể đính kèm trong Admin. Xuất PDF tự động thuộc T604, nghiệm thu giao diện toàn site thuộc T703.
- Theo dõi lần kiểm chứng và hợp nhất vào `develop` trên [PR #5](https://github.com/EthanLe-dev/blue-sky-group-website/pull/5).

### M6 — Báo cáo và tài liệu xuất

- Nhánh `feature/reporting`, PR #6 vào `develop`. T601–T605 có báo cáo theo ngày Việt Nam, phân quyền dashboard, xuất Excel/CSV, PDF công khai theo lịch trình và private theo phiên bản báo giá, cấu hình GA4/Search Console.
- Kiểm thử PostgreSQL dùng dữ liệu hai hành khách, một khoản thu đúng ranh giới nửa đêm và một khoản hoàn vào ngày kế tiếp. Playwright tải XLSX/CSV/PDF và kiểm tra sự kiện bằng thẻ GA stub có xác nhận cho phép.
- Kiểm tra file mẫu local: Microsoft Excel mở ở chế độ chỉ đọc, đủ 4 hàng/2 cột, ô dữ liệu nguy hiểm không phải công thức, tiền âm đọc là số và 16 chữ số được giữ dưới dạng text. Poppler render PDF A4 ba trang, kiểm tra trực quan trang 1 và 2 có dấu tiếng Việt và không cắt nội dung ở chân trang. App local không được khởi động.
- CI PR #6 là bằng chứng tích hợp. DebugView thật và xác minh Search Console cần tài khoản và địa chỉ site, chưa thực hiện.
- CI bản `3ddd0f3` có tám ca trình duyệt đạt, gồm PDF và analytics. Ca tải CSV dừng tại nhãn select và fixture kiểm quyền PDF thiếu hai trường bắt buộc. Đã bổ sung nhãn truy cập rõ cho lựa chọn định dạng và điền đủ fixture. Build Windows chạy được với Webpack sau lỗi tiến trình Turbopack. Build Linux CI đạt.
- Bản `aeaea662` đạt 6/6 job CI, gồm PostgreSQL và 9 ca Playwright. Xem [run 45022045577](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45022045577).

### M7 — Chuẩn bị phát hành MVP

- Nhánh feature/mvp-release, PR #7. T701 thêm fixture không trùng, không sửa đơn và lịch mẫu đã chỉnh, demo có vị trí minh họa. T702 thêm cọc rồi trả phần còn lại sau mất phản hồi, đặt tại quầy và chặn truy cập private. T703 có script viewport, bàn phím, ảnh public và reduced motion trên Chromium/Firefox/WebKit.
- T704 có Render Blueprint, migration Neon direct, R2 và Turnstile thật bắt buộc cho demo. T705 CD có công tắc mặc định tắt, deploy đúng SHA và smoke sau khi promote. Unit Python kiểm revision lệch. T706 chuẩn bị phiếu 11 tiêu chí, T707 runbook và drill backup/restore trong database CI riêng.
- Hoãn phần cần người dùng: tài khoản/credential, chạy trên URL cloud, diễn tập rollback trên host, xác minh GA4/Search Console và nghiệm thu với người thử độc lập. App trên máy người dùng không khởi động.
- CI 45024473766 đạt Go/PostgreSQL và 11 ca Chromium. Luồng tại quầy cần tên truy cập rõ cho bộ chọn tour, còn đăng nhập lượt sáu chạm rate limit của app. Đã gộp ca tại quầy với báo cáo dùng cùng phiên Admin và giữ giới hạn đăng nhập.
- CI 45025034406 đạt 12 ca Chromium và smoke Firefox. WebKit báo lỗi các request tải trước trang tour/Liên hệ. Đã tắt prefetch cho các liên kết liên quan và kiểm lại qua thao tác bấm link thật.
- Run 45025522153 xác nhận thao tác bấm tour và Liên hệ trong WebKit, còn lỗi tải trước nền danh sách tour. Tắt prefetch trên các liên kết ở những trang kiểm tra thay vì bỏ assertion lỗi trình duyệt. Liên kết vẫn điều hướng khi bấm.
- CI `783e6f86` đạt 6/6 job. Chromium 12 ca và smoke trên Firefox/WebKit đạt, có ảnh responsive. Diễn tập dump/restore PostgreSQL đạt cùng số tour, đơn và yêu cầu đoàn. [CI 45025910519](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45025910519).

### Menu dọc Admin

Nhánh feature/admin-sidebar. Menu chuyển sang cột trái cố định khi cuộn trên desktop. Trên mobile, nút Menu mở danh sách dọc, Escape trả focus về nút mở và chọn mục đưa focus tới nội dung. Thứ tự mục theo quyền hiện có. Ca Playwright Admin kiểm vị trí, mục đang chọn, điều hướng bàn phím và không tràn ngang ở 390px. App local không khởi động.

### Làm mới giao diện Admin

- T703b trên `feature/admin-refresh`: Nunito Sans, sidebar desktop có biểu tượng, menu mobile dạng Sheet, cập nhật màn đăng nhập, tiêu đề và form. Danh sách đơn tách nhãn trạng thái, tên và tổng tiền, hiển thị riêng đang tải, lỗi có nút thử lại và kết quả trống.
- Ca Playwright hiện có kiểm menu, focus, font tiếng Việt và trạng thái danh sách đơn. Ảnh desktop/mobile được lưu qua CI. App local không được khởi động.
