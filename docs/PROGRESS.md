# TIẾN ĐỘ DỰ ÁN BLUE SKY GROUP WEBSITE

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
