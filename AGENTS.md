# Hướng dẫn phát triển Blue Sky Group

Áp dụng cho toàn repository. Đặt file này tại root để Codex tự nạp hướng dẫn theo phạm vi thư mục. Các chỉ dẫn của người dùng trong phiên hiện tại được ưu tiên; file này không thay thế chỉ dẫn hệ thống. Giao tiếp, tài liệu và nội dung sản phẩm dùng tiếng Việt; tên mã nguồn/API/database dùng tiếng Anh.

## 1. Đọc trước khi làm

1. `docs/mvp-requirements.md`: phạm vi cần triển khai hiện tại.
2. `docs/final-requirements.md`: yêu cầu dài hạn và mã FR/UX/NFR/INT.
3. `docs/architecture.md`: stack, module và cấu trúc thư mục.
4. `agent/PLAN.md`: thứ tự công việc, quyết định đề xuất, các điểm chưa thống nhất.
5. `docs/database-schema.md`, `docs/api-spec.md`: hợp đồng dữ liệu và giao tiếp.

MVP quyết định tính năng nào làm ngay; final không tự động mở rộng MVP. Khi tài liệu mâu thuẫn, đọc bảng quyết định trong PLAN, ghi rõ giả định và hỏi khi ảnh hưởng nghiệp vụ chưa thể quyết định; không âm thầm đổi yêu cầu nguồn. Tài liệu thiết kế là đề xuất triển khai, không phải bằng chứng tính năng đã hoàn thành.

## 2. Kiến trúc và phạm vi

- Public: `frontend/`, Next.js App Router, TypeScript, SSR/SSG cho SEO. Admin: `admin/`, React SPA + TypeScript. Backend: `backend/`, Go + Gin modular monolith. PostgreSQL là nguồn dữ liệu giao dịch duy nhất.
- Giữ module theo architecture; handler xử lý HTTP, service giữ luật nghiệp vụ/transaction, repository truy cập DB. Dùng interface ở ranh giới tích hợp cần thay mock/real; tránh kiến trúc nhiều lớp không cần thiết.
- REST `/api/v1`, contract trong api-spec. Không cho frontend truy cập DB/R2 bằng secret. Backend quyết định giá, số chỗ, quyền và trạng thái đơn.
- MVP: thanh toán, email, SMS giả lập có nhãn rõ; tài khoản chỉ dành cho nhân viên. Redis, thanh toán thật, tài khoản khách, mã giảm giá, self-service huỷ/đổi thuộc giai đoạn sau.
- Không thay PostgreSQL bằng D1/SQLite. Phương án demo đã chọn trong hạn mức miễn phí: Render Free cho Go, Neon Free cho PostgreSQL; hạn mức và ngủ máy được ghi trong PLAN. Chưa chốt/triển khai dịch vụ chỉ bằng việc ghi vào tài liệu.

## 3. Quy tắc không được bỏ qua

- Chống overbooking bằng transaction PostgreSQL và khoá lịch khởi hành; cùng một service cho đơn web và đơn thủ công. Không dùng Redis hay bộ đếm frontend làm nguồn số chỗ.
- Giữ chỗ theo giờ DB; kiểm tra hạn ngay khi thao tác, không chỉ phụ thuộc worker. Payment confirm, expiry, cancel, refund phải idempotent và có kiểm thử cạnh tranh.
- Tiền VND dùng số nguyên, không float. Chốt bản sao giá/chính sách trong đơn. Đặt cọc không đồng nghĩa đã trả đủ; không sửa lịch sử giao dịch để làm khớp tổng.
- Mock chỉ bật trong local/demo. Production phải từ chối khởi động khi mock payment/notification/Turnstile được bật. Nút khách xác nhận chuyển khoản thật không được tự ghi nhận đã thu tiền.
- RBAC kiểm tra ở backend và từng object; ẩn nút không phải phân quyền. Hash mật khẩu, token ngẫu nhiên chỉ lưu hash, không log secret/token/PII đầy đủ. Dữ liệu khách và bản export không public.
- Validate input, parameterize SQL, sanitize rich text, giới hạn upload, chống CSRF và rate limit auth/lookup/form; Turnstile phải được xác minh server-side trên demo công khai.
- Không lưu số thẻ/CVV. Không commit `.env`, thông tin truy cập hay dữ liệu khách thật. Demo dùng dữ liệu giả có nhãn; ảnh/đánh giá thật chỉ khi có quyền sử dụng và nguồn xác thực.
- Không tự deploy, mua dịch vụ hoặc gửi thông báo thật ngoài phạm vi người dùng đã giao. Không xoá volume/database hoặc ghi đè thay đổi của người dùng.

## 4. Trải nghiệm và nội dung

- Mobile-first, chữ tối thiểu 16px, ưu tiên 17–18px mobile; vùng chạm tối thiểu 44×44px; tương phản AA, keyboard/focus và nhãn lỗi đầy đủ.
- Checkout tối đa 3–4 bước; tổng tiền, cọc, dư nợ, chính sách và hotline luôn rõ. Giữ dữ liệu khi validation lỗi. Không lưu PII checkout trong localStorage.
- Tôn trọng reduced motion, không tự phát âm thanh, lazy load media ngoài viewport; có loading/empty/error/expired/sold-out states.
- Số chỗ/đánh giá/ưu đãi phải có nguồn. Hiển thị rõ “Giả lập — không chuyển tiền thật”, “Email/SMS giả lập”; không thông báo đã gửi thật.
- Tính năng hoãn có UI phải disabled và giải thích, không giả vờ xử lý thành công.

## 5. Quy trình mỗi thay đổi

1. Kiểm tra trạng thái checkout, nhánh hiện tại và các hướng dẫn con, đọc code liên quan trước khi sửa. Giữ thay đổi đang có của người dùng; không tự reset/force-push.
2. Xác định FR/NFR và phạm vi; cập nhật PLAN khi bắt đầu/kết thúc công việc. Không đánh dấu hoàn thành chỉ vì đã có thiết kế.
3. Giữ diff tập trung. Không thêm dependency khi thư viện sẵn có đủ; dùng lockfile và phiên bản công cụ đã chốt khi scaffold.
4. Thay dữ liệu phải có migration có thứ tự, backfill/rollback hoặc kế hoạch phục hồi. Không sửa migration đã áp dụng. Cập nhật schema và API cùng thay đổi code.
5. Kiểm tra phù hợp: lint/typecheck/build frontend, `go vet`, `go test -race`, integration PostgreSQL cho logic booking/payment. Không thay integration bằng SQLite. Test phải kiểm tra hành vi, nhất là lỗi, quyền và concurrency.
6. Khi có ứng dụng, mỗi `frontend/` và `admin/` phải có npm scripts `lint`, `typecheck`, `test:ci`, `build` và `package-lock.json`; backend có `go.mod`, `go.sum`, test đọc `TEST_DATABASE_URL`. Mọi skip CI phải có lý do thấy rõ; không dùng `--if-present` để che thiếu check.
7. Báo lại file thay đổi, kết quả kiểm tra thật, phần chưa chạy và giới hạn. Ghi tiến độ có bằng chứng trong PLAN và `docs/PROGRESS.md`.

### Git và bàn giao task

- `main` là nhánh mặc định, chỉ nhận code đã chạy tốt sau khi qua CI, review và nghiệm thu liên quan. `develop` là nhánh tích hợp code chính; CI push/PR trên `develop` dùng đúng workflow và jobs như `main`. CD demo chỉ chạy sau CI của push vào `main`.
- Trước **mỗi module lớn**, tạo nhánh mới từ `develop` đã cập nhật: `feature/<module>` (ví dụ `feature/catalog`, `feature/booking`, `feature/payment`, `feature/admin`, `feature/public-ui`, `feature/reporting`). Module kéo dài nhiều task thì dùng chung nhánh module; module khác tạo nhánh khác. Sửa tài liệu/quy trình độc lập dùng `docs/<topic>` hoặc `chore/<topic>` nếu cần PR riêng.
- Mỗi task Txxx hoàn tất: chạy kiểm chứng trong TASKS/testing-strategy, cập nhật trạng thái và evidence trong PROGRESS, commit riêng cho task và **push ngay nhánh đó lên remote**. Message rõ, dạng `feat(booking): T303 prevent overselling on concurrent holds` hoặc `test(payment): T402 cover expiry race`; dùng `docs`, `fix`, `chore` khi phù hợp. Không gộp nhiều task chưa hoàn tất vào một commit; không ghi DONE trước test. Nếu task cần nhiều commit, mỗi commit nói rõ phần làm và ID, commit cuối đóng task; push sau khi task xong.
- Mở PR `feature/<module>` → `develop` sau khi module đủ task và CI xanh; xử lý review/CI, merge, cập nhật local `develop`. PR `develop` → `main` sau khi các task của đợt phát hành qua integration/UAT và CI xanh. Không push code module thẳng vào `main`; không dùng force-push trên nhánh chung.
- Task block/gián đoạn: ghi PROGRESS và giữ working tree, không commit nửa vời với nhãn DONE. Nếu cần chia sẻ WIP, commit có tiền tố `wip(<module>): Txxx ...`, push và ghi rõ chưa đạt gate; không merge WIP. Quy tắc quyền branch protection trên GitHub phải được cấu hình riêng; tài liệu không tự thực thi bảo vệ nhánh.

## 6. Lệnh hiện có

**Quyền chạy app:** người dùng tự chạy local. Agent không tự khởi động app, dev server, Docker Desktop hoặc các service/container local nếu chưa được người dùng yêu cầu rõ trong phiên đó. Nếu kiểm thử cần khởi động service, giải thích và xin phép trước; có thể tiếp tục kiểm tra tĩnh/build không khởi động app. Workflow CI vẫn chạy trong môi trường CI theo cấu hình đã có. Chỉ dẫn này ưu tiên hơn các lệnh chạy/test tự động khác trong tài liệu dự án.

Đã có frontend/admin/backend M1–M4, catalog/CMS/media, quote/booking/holds/lookup, sổ thanh toán/hoàn tiền và bản xem trước thông báo giả lập; nghiệm thu taste T703 còn mở (docs/ui-qa-m2-m3.md). Xem README để chạy ứng dụng.

```powershell
Copy-Item .env.example .env  # chỉ khi chưa có .env
docker compose --env-file .env config --quiet
docker compose up -d postgres
docker compose ps
docker compose down         # giữ dữ liệu volume; không thêm -v nếu chưa được phép xoá
```

Compose mặc định cấp PostgreSQL; profile `app` chạy migration/backend/frontend/admin/Nginx, Redis chỉ qua `production-tools`. Tạo `.env` an toàn bằng `python scripts/init_local.py`, chạy `docker compose --profile app up -d --build`, bootstrap riêng. TEST_DATABASE_URL chỉ dùng DB hậu tố `_test`; mỗi suite tạo schema cô lập.


## Quyết định người dùng 22/11/2026

Được giả định giá/đặt chỗ và thông tin công ty cho demo, bắt buộc sửa được trong Admin, có validation/audit/version/snapshot, không đổi ngược đơn đã đặt. Tour có poster do team thiết kế, kèm dữ liệu có cấu trúc. Đã chọn Render Free + Neon Free trong quota, không tự nâng gói mất phí. Không deadline bắt buộc. Ảnh poster mẫu đã nhận và được ghi trong `docs/poster-reference.md`; dùng đúng nguồn, không suy diễn giá/lịch/đối tượng chưa có trên ảnh. Chi tiết bổ sung trong PLAN/schema/API thay thế các giả định cũ chờ BA cho demo.

## Animation bằng GSAP

Người dùng yêu cầu dùng GSAP cho animation của website. `frontend/` cài `gsap` và `@gsap/react` trong package/lockfile; dùng cho hero entrance, timeline nhiều phần và gallery/moment có mục đích theo `docs/design-style.md`. Đọc skill `gsap-web` khi triển khai, cùng taste-skill cho giao diện. Trong React dùng `useGSAP()`/`gsap.context()` và cleanup khi unmount; đăng ký plugin trước khi dùng. Chỉ import trong client component, không truy cập `window` khi SSR. Tôn trọng `prefers-reduced-motion`, đảm bảo nội dung hiện sẵn khi JS không chạy, ưu tiên transform/opacity. Không thêm smooth-scroll/pin chỉ để trang trí; checkout/admin giữ thao tác rõ ràng.

## Tiết kiệm công sức bằng ponytail

Người dùng yêu cầu Agent áp dụng skill `ponytail` khi xây dựng. Trước mỗi task code/review, đọc skill hiện có trong catalog và dùng mức `full` mặc định: hiểu toàn bộ nơi thay đổi tác động, chọn cách nhỏ nhất giải quyết đủ yêu cầu, tái dùng code/thư viện đang có, tránh thêm abstraction hoặc dependency không cần thiết. Chỉ viết test có ý nghĩa cho logic mới; không bỏ validation, bảo mật, accessibility, concurrency hoặc kiểm tra bắt buộc để tiết kiệm token. Câu trả lời gọn, nêu kết quả kiểm chứng thật và rủi ro còn lại. User có thể đổi mức hoặc yêu cầu dừng skill ở phiên sau.
