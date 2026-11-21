# Blue Sky Group

Nền tảng demo: Next.js public, React Admin, Go/Gin API, PostgreSQL17. Chức năng hiện có: nền tảng auth/RBAC/guest/settings, chưa có module bán tour hoàn chỉnh.

## Chạy local bằng Docker

```powershell
python scripts/init_local.py
# Không chạy lại để ghi đè .env; script sẽ từ chối nếu file đã có.
docker compose --profile app up -d --build
docker compose --profile app run --rm backend bootstrap
```

Public: http://localhost:3000 · Admin: http://localhost:5173 · API: http://localhost:8080 · OpenAPI: http://localhost:8080/openapi.json · gateway: http://localhost:8088.

Tài khoản bootstrap nằm trong `.env` đã gitignore (`BOOTSTRAP_ADMIN_EMAIL`, `BOOTSTRAP_ADMIN_PASSWORD`); không có mật khẩu cố định, không đưa secret vào Git/chat. Lệnh bootstrap không ghi đè account đã tồn tại. Đăng nhập lần đầu rồi dùng cài đặt đặt tour/công ty/SLA. Local Docker dùng captcha mock; trước demo công khai phải build với Turnstile real + key, HTTPS, secure cookies và origin/hostname đúng.

Migration là service riêng chạy trước API, runner có advisory lock và version table; không migration trong mỗi replica. Production dùng migration role/direct connection, API role hạn chế. Sau đổi `.env`, recreate service liên quan. Default chỉ tạo PostgreSQL; profile app mới chạy ứng dụng. `docker compose down` giữ volume; không dùng `-v` cho dữ liệu cần giữ.

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
