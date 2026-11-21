# TIẾN ĐỘ DỰ ÁN BLUE SKY GROUP WEBSITE

## Bàn giao nhanh — 22/11/2026

- Giai đoạn hiện tại: hoàn thiện tài liệu trước triển khai; chưa có frontend/admin/backend, migration hoặc Dockerfile ứng dụng.
- Repository: https://github.com/EthanLe-dev/blue-sky-group-website, nhánh mặc định `main` tại `f80a064`; nhánh code chính `develop` đã tạo và push. Tài liệu TASKS/design/testing được commit `65d0708` trên `develop`.
- Đọc theo thứ tự: AGENTS → MVP/final/architecture → PLAN → TASKS → schema/API → design/testing → phần nhật ký mới nhất trong file này. Không bắt đầu từ mục “tiếp theo” cũ mà bỏ qua bàn giao hiện tại.
- Task có thể bắt đầu ngay: T101/T102/T103 (scaffold); trước UI cần T003/T004, T201 và đọc taste-skill. T001/T002 cần chốt nghiệp vụ/host, có thể scaffold với fixture demo có nhãn trong lúc chờ.
- Cần xác nhận: cọc/tuổi/chỗ em bé/hạn dư nợ/huỷ hoàn/SLA; nội dung và quyền ảnh; host demo và tiêu chí Cloudflare trong tài liệu nguồn. Render+Neon vẫn là đề xuất.
- Chưa có URL website hoặc kết quả test ứng dụng. Không có task triển khai Txxx nào DONE. Không có secret ứng dụng được cấp trong các file tài liệu.

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

### 22/11/2026 — T101: Next.js scaffold và GSAP

Người dùng yêu cầu cài GSAP và bắt buộc Agent dùng cho animation. Trên nhánh `feature/public-ui` từ develop, đã scaffold `frontend/` Next.js App Router + TypeScript, cài `gsap` và `@gsap/react`, thêm scripts CI và smoke test timeline. Cập nhật AGENTS/design/TASKS/PLAN. Chưa dựng website hoặc xác nhận UI/taste gate. Kết quả lệnh lint/typecheck/test/build/CI sẽ ghi sau khi hoàn tất. Lưu ý nhánh docs/admin-settings-and-posters chứa quyết định poster/Admin riêng, chưa merge vào develop khi bắt đầu nhánh này.

T101 verification: Node24.16/npm11.13 local, `npm ci`, `npm run lint`, `npm run typecheck`, `npm run test:ci` (1 pass), `npm run build` đều exit0 sau sửa test sang ESM. Next16.4.6, GSAP3.15.1. UI mới là scaffold “Hello world”, chưa triển khai animation thật/taste gate; T208/T703 mở. Commit/push feature/public-ui.

### 22/11/2026 — Quy tắc dùng ponytail

Người dùng yêu cầu Agent dùng skill `ponytail` khi xây dựng để tiết kiệm token. Đã thêm quy tắc vào AGENTS.md: mức full mặc định, chọn thay đổi nhỏ nhất đủ yêu cầu, tái dùng code và không bỏ các kiểm tra quan trọng. Đây là cập nhật hướng dẫn; chưa có code ứng dụng mới. Kiểm tra `git diff --check`, commit/push trên nhánh feature/public-ui.
