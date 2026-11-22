# Bàn giao demo MVP

Bản demo dùng Vercel cho website và Admin, Render cho Go API và Neon cho PostgreSQL. Cloudflare Turnstile kiểm tra form công khai, R2 giữ file sau khi Render ngủ hoặc khởi động lại. Chỉ dùng dữ liệu mẫu. Thanh toán, hoàn tiền, email và SMS đều giả lập.

## Tạo môi trường

1. Tạo PostgreSQL riêng cho demo trong Neon. Lấy hai connection string cùng database và bắt buộc TLS. Dùng URL pooled cho `DATABASE_URL`, URL direct cho `DATABASE_MIGRATION_URL`. Giữ quyền tạo bảng trên kết nối migration.
2. Tạo hai bucket R2 riêng cho public/private. Để cả hai bucket tắt truy cập ẩn danh. Backend đọc qua SDK rồi quyết định quyền dựa trên metadata, tên file và chữ ký có hạn. Cấp token R2 chỉ đọc/ghi hai bucket. Tạo widget Turnstile cho hostname ổn định của Vercel public và Admin.
3. Tạo Go Web Service Free ở Render. Khi blueprint đã có trên nhánh phát hành, dùng [render.yaml](../render.yaml), hoặc tạo service Docker với Dockerfile `backend/Dockerfile` và context `backend`. Tắt auto-deploy, dùng health check `/readyz`, để Render cấp `PORT`. Điền các mục `sync: false` từ blueprint: URL Neon, hai secret khác nhau dài ít nhất 32 ký tự, danh sách origin HTTPS chính xác, secret/hostname Turnstile và thông tin R2. Render tiến hành migration qua kết nối direct trước khi khởi động HTTP. Dữ liệu và file giữ ngoài máy Render.
4. Tạo hai project Vercel có Root Directory trống khi deploy từ gói từng app. Dùng preset Next.js và Vite. Giữ hai URL HTTPS ổn định cho Turnstile, CORS và smoke test. Public cần `APP_ENV=demo`, `SITE_URL`, `API_UPSTREAM_URL` trỏ API Render, `PAYMENT_PROVIDER=mock`, `TURNSTILE_MODE=real`, `TURNSTILE_SITE_KEY`. Admin cần `VITE_TURNSTILE_MODE=real`, `VITE_TURNSTILE_SITE_KEY`. Không cấp secret backend cho Vercel. Cấu hình rewrite của Admin được tạo khi staging, giữ cookie theo origin Admin. Tắt deploy tự động bên nhà cung cấp và dùng CLI/CD. Tài khoản/gói host phải phù hợp mục đích demo.
5. Khi test tài khoản thật, chọn sao cho ổn định: Render Free có thể ngủ khi nhàn rỗi. Đo request đầu sau khi ngủ, tiếp tục thử lại từ UI khi API sẵn sàng. Kiểm tra giới hạn tải file thực tế qua Vercel bằng file nhỏ và file gần giới hạn nếu cần demo media lớn.

## Bootstrap và dữ liệu mẫu

Chuẩn bị file env ngoài Git, chỉ chứa cấu hình của database và bucket demo. Thêm `BOOTSTRAP_ADMIN_EMAIL` và mật khẩu riêng từ 12 ký tự trở lên trong lần tạo tài khoản. Dùng đúng SHA đã qua CI.

```powershell
docker build --tag blue-sky-demo-api ./backend
docker run --rm --env-file <duong-dan-env-demo> blue-sky-demo-api bootstrap
docker run --rm --env-file <duong-dan-env-demo> blue-sky-demo-api seed-demo
```

Bootstrap là thao tác một lần. Dọn biến mật khẩu sau khi tạo. Seed chạy lại được, không sửa tour, đơn, số chỗ đã thay đổi hoặc tài khoản Admin hiện có. Tuỳ thời điểm chạy, các lịch mẫu đã tồn tại có thể qua ngày. Tạo lịch tương lai từ Admin để kiểm tra tiếp. Có thể lọc `DEMO-DOAN` trong hàng đợi và tìm khách minh họa trong danh sách đơn. Vị trí liên hệ mẫu cần thay bằng tọa độ do công ty xác nhận.

## Phát hành và quay lại bản trước

- Trước nghiệm thu, có thể dựng bản demo từ SHA `develop` **đã qua CI**: deploy đúng commit đó trên Render, dùng `git archive <SHA> frontend admin` để lấy hai thư mục, rồi chạy `python scripts/demo_release.py prepare --root <thu-muc-giai-nen>` với biến `RELEASE_SHA` và `RENDER_URL`. Từ từng thư mục app chạy Vercel CLI liên kết đúng project. Sau khi hai bản dựng và kiểm tra URL preview thành công, đưa chúng lên các domain ổn định và chạy `python scripts/demo_release.py smoke`.
- Khi đã nghiệm thu và merge PR phát hành vào `main`, CI trên push `main` mới có thể kích hoạt workflow Deploy demo. Công tắc repo `DEMO_DEPLOY_ENABLED` chỉ đặt `true` sau khi cấp quyền và kiểm tra cấu hình. Environment `demo` lưu secret `RENDER_API_KEY` và `VERCEL_TOKEN`. Cấu hình các vars `RENDER_URL`, `RENDER_SERVICE_ID`, `PUBLIC_DEMO_URL`, `ADMIN_DEMO_URL`, `VERCEL_ORG_ID`, `VERCEL_PUBLIC_PROJECT_ID`, `VERCEL_ADMIN_PROJECT_ID`. Các URL là origin HTTPS, không có dấu slash cuối. File `revision.txt` được tạo từ SHA CI, Admin có rewrite cố định về API Render.
- Pipeline so SHA với đầu nhánh main, yêu cầu Render deploy SHA cụ thể và chờ `live` cùng readiness đúng revision. Vercel nhận nguồn xuất từ SHA đó, xây hai bản trước khi chuyển domain. Smoke đọc tour, thử tạo guest session an toàn và xác nhận API Admin từ chối khi chưa đăng nhập. Smoke có thể tạo guest session thử nhưng không tạo đơn hoặc gửi thông báo.
- Trước khi chuyển domain, lưu SHA, ID từ Render và hai ID deployment đang chạy trên Vercel. Nếu sau chuyển domain gặp lỗi, tắt công tắc deploy và đưa hai project Vercel về deployment trước. Chọn deployment Render trước từ Dashboard nếu schema tương thích. Chạy smoke với `RELEASE_SHA` của bản quay lại. Bản không có cơ chế báo revision cần kiểm tra thủ công theo provider.
- Rollback code giữ nguyên database. Nếu migration đã làm dữ liệu không tương thích, dừng ghi rồi phục hồi sang DB trống từ bản sao trước khi đổi kết nối. Không chạy SQL giảm phiên bản vào DB đang dùng. Kiểm tra bản sao trước khi chuyển kết nối.

## Sao lưu và kiểm tra khôi phục

- Trước release dùng `pg_dump --format=custom --file=demo.dump` với kết nối direct của database demo. Đặt credential qua biến môi trường của công cụ PostgreSQL hoặc tệp cấu hình nằm ngoài repo. Giữ dump ở kho riêng.
- Sao chép dữ liệu từ hai bucket R2 vào kho sao lưu riêng cùng thời điểm. Lưu bucket và timestamp đi kèm dump. Tránh công cụ đồng bộ xóa object trong nguồn. Database lưu khóa object, còn bản sao R2 giữ nội dung file.
- Dùng database trống riêng để chạy `pg_restore --exit-on-error --no-owner --no-privileges --dbname=<fresh-demo-db> demo.dump`. Đối chiếu số tour, đơn, group request và migration. Thử đọc một media public và private từ bản sao trước khi cân nhắc chuyển kết nối.
- Quy trình local không cần credential cloud. Cấu hình theo [runbook local](platform-runbook.md), người vận hành tự mở và tắt bằng `bluesky run` và `bluesky off`. CI dùng schema/dữ liệu độc lập và giữ log có thời hạn.

Secret không được gửi trong chat. Nếu chưa có tài khoản/URL, giữ CD tắt, dùng CI để kiểm thử và theo dõi các mục chờ trong [phiếu nghiệm thu](uat-mvp.md). Bản demo chưa bao gồm thanh toán/gửi thông báo thật, tải lớn hay cam kết uptime production.
