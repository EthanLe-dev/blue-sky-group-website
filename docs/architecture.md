# KIẾN TRÚC HỆ THỐNG – BLUE SKY GROUP WEBSITE

| Thông tin      | Nội dung                                                                  |
| -------------- | ------------------------------------------------------------------------- |
| **Dự án**      | Website bán tour và chăm sóc khách hàng của Blue Sky Group                |
| **Tài liệu**   | Tài liệu Thiết kế Kiến trúc Hệ thống (System Architecture Document - SAD) |
| **Tham chiếu** | `final-requirements.md`, `mvp-requirements.md`, `database-schema.md`      |
| **Phiên bản**  | 1.0                                                                       |
| **Ngày lập**   | 22/11/2026                                                                |

---

## 1. Tổng quan Công nghệ (Tech Stack)

Hệ thống được thiết kế theo hướng hiện đại, phân tách rõ ràng giữa Public Client, Internal Admin Client và Backend Services.

| Thành phần            | Công nghệ                    | Mục đích & Vai trò                                                                                |
| --------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------- |
| **Frontend (Public)** | Next.js (App Router, SSR/SSG)| Giao diện khách hàng: tối ưu SEO, chỉ số Core Web Vitals, kết xuất nhanh, tương thích mobile tốt. |
| **Frontend (Admin)**  | React (SPA)                  | Giao diện quản trị nội bộ: tải trang mượt mà, quản lý trạng thái phức tạp, dashboard trực quan.   |
| **Backend API**       | Go + Gin                     | Hiệu năng cao, độ trễ thấp, xử lý đồng thời (concurrency) vượt trội, tiết kiệm tài nguyên.       |
| **Database**          | PostgreSQL                   | Cơ sở dữ liệu quan hệ giao dịch (ACID), đảm bảo toàn vẹn dữ liệu đặt chỗ và giao dịch.           |
| **Cache & Message**   | Redis                        | Lưu cache truy vấn đọc nhiều (tour, danh mục), rate limit, session (bổ sung giai đoạn Production).|
| **Lưu trữ đối tượng** | Cloudflare R2                | Lưu trữ ảnh chất lượng cao, video, tài liệu PDF, file Excel xuất ra (tương thích S3 API).        |
| **Reverse Proxy**     | Nginx                        | Cân bằng tải, reverse proxy, SSL termination, định tuyến API và static assets.                   |
| **Đóng gói**          | Docker & Docker Compose      | Chuẩn hóa môi trường phát triển (local), đóng gói container cho CI/CD và deployment.             |
| **Bảo mật & CDN**     | Cloudflare (Turnstile & WAF) | CDN phân phối tĩnh, chống DDoS, bảo vệ WAF, chặn spam bot biểu mẫu qua Cloudflare Turnstile.      |

---

## 2. Phong cách Kiến trúc Backend (Modular Monolith)

Hệ thống backend được xây dựng theo mô hình **Modular Monolith** kết hợp nguyên lý **Clean / Hexagonal Architecture** cho các module nghiệp vụ phức tạp.

### 2.1 Tại sao chọn Modular Monolith?
- **Đơn giản & Tối ưu vận hành:** Phù hợp hoàn hảo cho quy mô tải mục tiêu (100–300 người dùng đồng thời) mà không phát sinh chi phí vận hành mạng và overhead phức tạp của microservices.
- **Ranh giới module rõ ràng:** Các module độc lập về mặt logic, giao tiếp thông qua contract/interface nội bộ. Khi cần thiết trong tương lai, một module có thể dễ dàng tách thành service riêng mà không phá vỡ cấu trúc tổng thể.
- **Dễ dàng kiểm thử & Deploy:** Triển khai một khối ứng dụng duy nhất, đơn giản hóa quy trình CI/CD.

### 2.2 Các Module nghiệp vụ chính
1. **Catalog Module:** Quản lý điểm đến, danh mục tour, thông tin chi tiết tour, media, FAQ, đánh giá.
2. **Departure & Inventory Module:** Quản lý các đợt khởi hành, phân bổ số lượng chỗ, trạng thái mở/đóng bán.
3. **Pricing Module:** Bảng giá theo đối tượng (người lớn, trẻ em, em bé), phụ thu, chính sách giá theo mùa, mã giảm giá.
4. **Booking Module:** Luồng đặt tour (guest checkout & member), giữ chỗ tạm thời (seat lock/hold), hủy tour.
5. **Payment Module:** Giao diện tích hợp cổng thanh toán (VNPay, MoMo, VietQR) và cơ chế mock payment phục vụ thử nghiệm.
6. **Customer Module:** Thông tin liên hệ khách hàng, hồ sơ thành viên, lịch sử đặt tour.
7. **Group Quote Module:** Tiếp nhận và xử lý yêu cầu báo giá tour đoàn (trường học, doanh nghiệp).
8. **Content Module:** Quản lý bài viết blog, cẩm nang du lịch, banner, trang tĩnh.
9. **Notification Module:** Điều phối gửi email, SMS, thông báo giao dịch (hỗ trợ hàng đợi bất đồng bộ).
10. **Reporting Module:** Tổng hợp số liệu doanh thu, tỷ lệ lấp đầy, xuất báo cáo Excel (`excelize`) và tài liệu PDF.
11. **Auth & Identity Module:** Xác thực, phân quyền theo vai trò (RBAC) và nhật ký thao tác kiểm toán (audit log).

### 2.3 Cơ chế xử lý đồng thời & Giữ chỗ (Seat Locking)
- Tránh tình trạng **overbooking** (bán vượt số chỗ) khi nhiều khách hàng cùng bấm đặt tour một lúc.
- Áp dụng cơ chế **Pessimistic Locking (`SELECT ... FOR UPDATE`)** hoặc cơ chế **Atomic Update kèm ràng buộc số dư (`available_seats >= requested_seats`)** ở mức Database Transaction.
- Cột mốc giữ chỗ có thời hạn (`expires_at` khoảng 15–30 phút). Tác vụ nền (background worker/cron) định kỳ kiểm tra và hoàn trả chỗ chưa thanh toán vào kho trống.

---

## 3. Kiến trúc Triển khai (Deployment Architecture)

### 3.1 Môi trường MVP Demo (Tối ưu chi phí & Triển khai nhanh)
Phục vụ mục tiêu nghiệm thu giao diện, trải nghiệm người dùng thực tế và luồng nghiệp vụ không phát sinh chi phí hạ tầng.

```text
Browser → Vercel public/admin → Render Free (Go/Gin) → Neon Free (PostgreSQL)
                                        ├→ Cloudflare R2 (media)
                                        └→ Cloudflare Turnstile (form verification)
```

- **Frontend:** Next.js (Public) và React (Admin) deploy trực tiếp trên **Vercel** thông qua Git integration, tự động build và có sẵn CDN/SSL.
- **Backend API:** Go-Gin deploy trên Render Free; ngủ khi không có truy cập, không lưu file bền trên filesystem local.
- **Database:** PostgreSQL trên Neon Free; không dùng D1/SQLite. TLS cho kết nối, pooled URL cho app và direct URL cho migration.
- **Static Assets & Media:** Lưu trữ tại **Cloudflare R2** bucket kết hợp đường dẫn CDN.
- **Bảo mật Form:** Tích hợp widget Cloudflare Turnstile miễn phí chống bot submit form.

### 3.2 Môi trường Production (Sẵn sàng tải cao & Độ ổn định)
Sau khi nghiệm thu MVP thành công, hệ thống chuyển dịch lên hạ tầng Production:

```
                                  [ DNS & Cloudflare Pro ]
                                             │ (WAF, DDoS, Edge Cache, Turnstile)
                                             ▼
                                  [ Nginx Reverse Proxy ]
                                             │
                   ┌─────────────────────────┴─────────────────────────┐
                   ▼                                                   ▼
       [ Next.js Frontend Cluster ]                           [ Go Backend Cluster ]
       (Server-Side Rendering / SSG)                          (Gin Engine - Modular Monolith)
                                                                       │
                         ┌─────────────────────────────────────────────┼─────────────────────────────────────────────┐
                         ▼                                             ▼                                             ▼
                 [ PostgreSQL ]                                     [ Redis ]                                 [ Cloudflare R2 ]
             (Primary - Read Replica)                         (Cache, Session, Queue)                         (Images, PDFs, Media)
```

- Domain tùy chỉnh và chứng chỉ SSL bảo mật nâng cao.
- **Nginx:** Đảm nhận vai trò gateway, nén dữ liệu (gzip/brotli), giới hạn tần suất request (rate limiting) và chuyển tiếp lưu lượng.
- **Redis Cache:** Cache các API đọc danh sách tour, chi tiết tour, giảm tải 80% truy vấn trực tiếp vào PostgreSQL.
- **Worker & Async Queue:** Xử lý gửi email giao dịch, SMS OTP, tạo hóa đơn và xuất báo cáo nặng trong background.

---

## 4. Cấu trúc Thư mục Dự án

```
blue-sky-group-website/
├── frontend/                 # Next.js App Router (Public website)
│   ├── src/
│   │   ├── app/              # Routes: trang chủ, tour, booking, blog, contact...
│   │   ├── components/       # UI Primitives, Layouts, Widgets
│   │   ├── lib/              # API Fetch client, hooks, utils
│   │   └── styles/           # CSS modules / Tailwind styles
│   ├── public/               # Static assets cục bộ
│   ├── Dockerfile
│   └── package.json
│
├── admin/                    # React SPA (Hệ thống quản trị nội bộ)
│   ├── src/
│   │   ├── pages/            # Dashboard, quản lý tour, đơn hàng, báo cáo...
│   │   ├── components/       # Table, modal, form controls
│   │   └── lib/              # Admin state & API integration
│   ├── Dockerfile
│   └── package.json
│
├── backend/                  # Go + Gin REST API Server
│   ├── cmd/server/           # Main entry point (main.go)
│   ├── internal/             # Toàn bộ business logic theo Modular Monolith
│   │   ├── catalog/          # Tour, địa điểm, danh mục
│   │   ├── departure/        # Quản lý lịch khởi hành, số chỗ
│   │   ├── pricing/          # Tính giá, phụ thu, khuyến mãi
│   │   ├── booking/          # Quản lý giỏ hàng, đặt tour, giữ chỗ
│   │   ├── payment/          # Cổng thanh toán & mock QR
│   │   ├── customer/         # Khách hàng, hồ sơ
│   │   ├── groupquote/       # Tiếp nhận yêu cầu tour đoàn
│   │   ├── content/          # Bài viết, tin tức, FAQ, media
│   │   ├── notification/     # Điều phối email, SMS (mock & real)
│   │   ├── reporting/        # Xuất dữ liệu Excel/PDF, thống kê
│   │   ├── auth/             # JWT, vai trò người dùng, RBAC
│   │   └── common/           # Database pool, logging, middleware
│   ├── migrations/           # File kịch bản cấu trúc DB (.sql)
│   ├── Dockerfile
│   └── go.mod
│
├── docker-compose.yml        # Docker Compose phối hợp các dịch vụ local (Nginx, Go, PG)
├── nginx/
│   └── nginx.conf            # Cấu hình proxy & routing
│
├── docs/                     # Toàn bộ tài liệu đặc tả dự án
│   ├── final-requirements.md # Yêu cầu nghiệp vụ bản hoàn chỉnh
│   ├── mvp-requirements.md   # Yêu cầu và phạm vi bản MVP
│   ├── architecture.md       # Thiết kế kiến trúc hệ thống (file này)
│   ├── database-schema.md    # Thiết kế cơ sở dữ liệu chi tiết
│   └── PROGRESS.md           # Nhật ký tiến độ dự án
│
└── README.md
```

---

## 5. Tiêu chuẩn Kỹ thuật & Giao tiếp (Interfaces)

1. **Giao thức Giao tiếp:**
   - Client và Server giao tiếp hoàn toàn thông qua **RESTful API**, dữ liệu mã hóa định dạng `application/json`.
   - Chuẩn hóa mã phản hồi HTTP (`200 OK`, `201 Created`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `409 Conflict`, `500 Internal Error`).
   - Cấu trúc response thống nhất:
     ```json
     {
       "success": true,
       "data": {},
       "error": null,
       "message": "Thông điệp (nếu có)"
     }
     ```

2. **Bảo mật Dữ liệu & Phiên:**
   - Xác thực Admin bằng **JWT (JSON Web Token)** kèm cơ chế refresh token hoặc HttpOnly cookie.
   - Toàn bộ dữ liệu khách hàng nhạy cảm tuân thủ quy định bảo vệ dữ liệu cá nhân (mã hóa mật khẩu bằng bcrypt/argon2, không lưu thông tin thẻ ngân hàng trực tiếp).

3. **Xử lý Xuất file:**
   - **Excel:** Sử dụng thư viện `excelize` trong Go để ghi dữ liệu trực tiếp dưới dạng stream hoặc lưu vào bucket Cloudflare R2 và sinh pre-signed download URL.
   - **PDF:** Sinh file PDF lịch trình hoặc báo giá thông qua engine template HTML to PDF (`chromedp` / `wkhtmltopdf`).
