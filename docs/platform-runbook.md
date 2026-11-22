# Runbook nền tảng M1

## Migration và quay lui

Không sửa migration đã áp dụng. Runner áp dụng toàn bộ migration mới trong transaction PostgreSQL và advisory lock; lỗi rollback mọi DDL của lượt đó. Dữ liệu quan trọng phải backup trước nâng phiên bản. Schema nền tảng hiện chỉ additive; rollback ứng dụng về image trước tương thích schema, ưu tiên forward fix. Không cung cấp lệnh down xoá dữ liệu để rollback production. Khi cần khôi phục DB: dừng ghi, restore bản sao sang DB mới, kiểm role/session/settings/counters, rồi đổi connection và thu hồi sessions. Test restore là cổng production riêng, không khẳng định đã diễn tập production trong M1.

## Quyền và dữ liệu

5 role được seed; bootstrap từ secret, không default password. Access JWT15 phút trong memory, refresh cookie7 ngày; rotation/reuse revoke family, logout/reset revoke session. Cookies HttpOnly/Lax, demo HTTPS Secure; guest cookie24 giờ và CSRF. Lookup token cần FK booking, tách T108b ở M3 khi có bảng booking. Company document upload chờ T215, form hiện không giả vờ nhận file. Password reset mock preview là ciphertext, chỉ admin quyền staff.write đọc được, đọc có audit; reset token15 phút, dùng một lần.

Public settings DTO là company_profile/response_sla. Default booking chỉ áp dụng departure/quote/booking mới khi module đó được xây; lịch sử bất biến theo schema. Không dùng settings để đổi ledger.

## CI và branch checks

Required status đề xuất: `CI result` (foundation, frontend, admin, backend và container smoke nếu áp dụng). Kiểm tra branch protection ở GitHub Settings trước release; hiện chưa tự cấu hình required checks. PR failing không merge. Pipeline container dùng `.env` random riêng CI, migration/bootstrap, local-only captcha; không dùng secret production. Failure artifacts không chứa .env/cookie/password. Render/Neon chưa được tạo hoặc deploy trong phiên nền tảng.

## M2/M3 — catalog, media và đặt chỗ

Người dùng tự chạy bằng `bluesky run`, tự dừng bằng `bluesky off`. Agent chỉ được chạy kiểm thử local khi được cho phép trong phiên đó.

- Lần đầu cần dữ liệu demo sau migration và tài khoản đã tồn tại: `docker compose --profile app run --rm backend seed-demo`. Seed chỉ local/demo/test, không reset password và không ghi đè tour/trang đã có. Không chạy lại bootstrap bằng mật khẩu tạm ngắn của local.
- Public: http://localhost:3000 — danh mục/detail, checkout, lookup; Admin: http://localhost:5173 — Tour & Nội dung, Media, Đơn tour, Cài đặt. Giá theo từng lịch đi; defaults chỉ áp dụng dữ liệu mới.
- Company logo: `frontend/public/brand/logo.jpg` và `admin/public/brand/logo.jpg`, từ file người dùng chỉ định. Poster mẫu trong backend/demo-assets, Media cho thay poster tour; lưu rồi đối chiếu và bấm xác minh để public hiện poster. Sửa nội dung/giá/lịch cần xác minh lại.
- Media local dùng named volume `media_data`; không `down -v`. Backup phải có cả PostgreSQL và media. R2 dùng biến R2_* trong .env.example, chưa kiểm tra trên tài khoản cloud thật; test adapter dùng S3 endpoint giả lập có xác minh chữ ký AWS4.
- Các trường itinerary/prices/stops nâng cao trong Admin dùng JSON có nhãn cấu trúc; backend validate, tiền phải chuỗi số nguyên (`"1650000"`), không float. Lỗi giữ form đang sửa. Không nhập thông tin khách thật vào demo.
- M3 tạo đơn pending và giữ chỗ theo thời hạn chốt trong quote; worker30s và kiểm tra trực tiếp xử lý hết hạn. Checkout không tự gia hạn khi refresh. Checkout hỗ trợ thanh toán toàn bộ, đặt cọc và trả phần còn lại qua QR giả lập. Liên kết mở lại đơn cần phiên đã tạo đơn còn hiệu lực. Admin có thể huỷ đơn để trả chỗ, còn hoàn tiền là thao tác riêng của người có quyền payment.write. Đánh dấu hoàn thành chỉ dành cho đơn xác nhận của chuyến đã kết thúc. Nhật ký thông báo hiển thị bản xem trước và lượt tạo thất bại cho người có quyền notification.read.
- HTTP404 slug SSR dùng metadata blocking và không streaming loading toàn trang; nội dung pending có state trên form và API. Tham khảo Next [not-found](https://nextjs.org/docs/app/api-reference/file-conventions/not-found) và [metadata](https://nextjs.org/docs/app/getting-started/metadata-and-og-images).
- `SITE_URL`, APP_ENV và TURNSTILE_* cần đúng host trước public demo. Demo/local không index; deploy thật ở M7 kiểm tra lại canonical, TLS, Turnstile real và R2. Render+Neon chưa được tự deploy trong đợt code này.

Bổ sung M2: bản đồ dùng Leaflet1.9.4 lazy import, chỉ gọi tile OpenStreetMap sau khi bấm mở, giữ attribution và có liên kết địa chỉ fallback. Không geocode địa chỉ giả; vị trí tour mẫu cần xác minh. Tham chiếu [Leaflet stable](https://leafletjs.com/reference-1.9.4.html). Home có ô tìm nhanh, CMS banners/điểm đến/cẩm nang/đánh giá đã duyệt; Blog có category/tag/related; gallery có kind/collection và viewer. Phụ thu optional chọn ở checkout và form đơn thủ công, backend tự thêm phụ thu bắt buộc.
