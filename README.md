# Blue Sky Group

Nền tảng demo: Next.js public, React Admin, Go/Gin API, PostgreSQL17. Chức năng hiện có: nền tảng auth/RBAC/guest/settings, chưa có module bán tour hoàn chỉnh.

## Chạy local bằng Docker

Người dùng tự chạy app; Agent chỉ khởi động local khi được yêu cầu rõ. Mở Docker Desktop, rồi mở PowerShell tại thư mục repository. Với máy hiện tại đã có `.env`, Admin và dữ liệu, chỉ cần:

```powershell
docker compose --profile app up -d --build
```

Xem trạng thái/log và dừng app:

```powershell
docker compose --profile app ps
docker compose --profile app logs -f backend
# Ctrl+C thoát xem log, không dừng app.
docker compose --profile app --profile production-tools down
```

Lệnh down giữ dữ liệu volume. Không thêm `-v`. Chỉ dùng các lệnh sau cho lần setup đầu tiên trên máy chưa có `.env`/Admin:

```powershell
python scripts/init_local.py
# Không chạy lại để ghi đè .env; script sẽ từ chối nếu file đã có.
docker compose --profile app up -d --build
docker compose --profile app run --rm backend bootstrap
```

Public: http://localhost:3000 · Admin: http://localhost:5173 · API: http://localhost:8080 · OpenAPI: http://localhost:8080/openapi.json · gateway: http://localhost:8088.

Tài khoản bootstrap nằm trong `.env` đã gitignore (`BOOTSTRAP_ADMIN_EMAIL`, `BOOTSTRAP_ADMIN_PASSWORD`); không có mật khẩu cố định, không đưa secret vào Git/chat. Lệnh bootstrap không ghi đè account đã tồn tại. Đăng nhập lần đầu rồi dùng cài đặt đặt tour/công ty/SLA. Local Docker dùng captcha mock; trước demo công khai phải build với Turnstile real + key, HTTPS, secure cookies và origin/hostname đúng.

Migration là service riêng chạy trước API, runner có advisory lock và version table; không migration trong mỗi replica. Production dùng migration role/direct connection, API role hạn chế. Sau đổi `.env`, recreate service liên quan. Default chỉ tạo PostgreSQL; profile app mới chạy ứng dụng. `docker compose down` giữ volume; không dùng `-v` cho dữ liệu cần giữ.

### Nếu migrate báo database authentication failed

`POSTGRES_PASSWORD` chỉ đặt mật khẩu lúc volume được khởi tạo lần đầu. Sửa biến này trong `.env` không đổi mật khẩu role đã lưu trong PostgreSQL. Xem `docker compose logs migrate postgres` để phân biệt lỗi đăng nhập với lỗi migration SQL. Không xoá volume để xử lý lỗi mật khẩu.

Đồng bộ `.env` với mật khẩu database cũ, hoặc khi PostgreSQL đã chạy, mở psql qua Unix socket và đặt mật khẩu role theo giá trị trong `.env`:

```powershell
docker compose exec postgres psql -U blue_sky -d postgres
```

Trong psql chạy `\password blue_sky`, nhập mật khẩu khớp `.env`, rồi `\q`. Thay `blue_sky` bằng POSTGRES_USER thực tế nếu đã đổi user. Không dán mật khẩu vào câu lệnh Git hoặc chat. Sau đó tự chạy lại `docker compose --profile app up -d --build`. Healthcheck PostgreSQL kiểm tra đăng nhập TCP thật để báo sai mật khẩu sớm.

## Kiểm tra

```powershell
npm --prefix frontend ci
npm --prefix frontend run lint
npm --prefix frontend run typecheck
npm --prefix frontend run test:ci
npm --prefix frontend run build
npm --prefix admin ci
npm --prefix admin run lint
npm --prefix admin run typecheck
npm --prefix admin run test:ci
npm --prefix admin run build
```

Backend: Go1.26, `go vet ./...`, `go test -race -count=1 ./...`, `go build ./...` trong backend. Integration yêu cầu TEST_DATABASE_URL trỏ DB tên kết thúc `_test`; tests tạo/dọn schema riêng, không xoá bảng của môi trường khác. Nếu không có URL, integration được ghi skip; CI luôn cấp PostgreSQL và URL, không chấp nhận skip để nghiệm thu.

Browser smoke local: từ admin chạy `npx playwright install chromium`, `npm run test:e2e` sau Docker/ bootstrap. Test đọc credential từ `.env`, không in secret, đổi hold20→21→20, kiểm reload và logout, public proxy cookie/CSRF. Không dùng test này trên demo có dữ liệu thật.

## Quy trình

`develop` là nhánh tích hợp, nhánh module `feature/<module>`; task hoàn tất được commit và push. PR module → develop, bản nghiệm thu → main. CI chạy chung cho main/develop/PR; CD main hiện là hook, chưa chứng minh provider deploy đúng revision/readiness, xem T705. Không deploy chỉ vì CI xanh.

Tài liệu chính: [TASKS](agent/TASKS.md), [PLAN](agent/PLAN.md), [PROGRESS](docs/PROGRESS.md), [API](docs/api-spec.md), [schema](docs/database-schema.md).
