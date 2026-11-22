# Runbook nền tảng M1

## Migration và quay lui

Không sửa migration đã áp dụng. Runner áp dụng toàn bộ migration mới trong transaction PostgreSQL và advisory lock; lỗi rollback mọi DDL của lượt đó. Dữ liệu quan trọng phải backup trước nâng phiên bản. Schema nền tảng hiện chỉ additive; rollback ứng dụng về image trước tương thích schema, ưu tiên forward fix. Không cung cấp lệnh down xoá dữ liệu để rollback production. Khi cần khôi phục DB: dừng ghi, restore bản sao sang DB mới, kiểm role/session/settings/counters, rồi đổi connection và thu hồi sessions. Test restore là cổng production riêng, không khẳng định đã diễn tập production trong M1.

## Quyền và dữ liệu

5 role được seed; bootstrap từ secret, không default password. Access JWT15 phút trong memory, refresh cookie7 ngày; rotation/reuse revoke family, logout/reset revoke session. Cookies HttpOnly/Lax, demo HTTPS Secure; guest cookie24 giờ và CSRF. Lookup token cần FK booking, tách T108b ở M3 khi có bảng booking. Company document upload chờ T215, form hiện không giả vờ nhận file. Password reset mock preview là ciphertext, chỉ admin quyền staff.write đọc được, đọc có audit; reset token15 phút, dùng một lần.

Public settings DTO là company_profile/response_sla. Default booking chỉ áp dụng departure/quote/booking mới khi module đó được xây; lịch sử bất biến theo schema. Không dùng settings để đổi ledger.

## CI và branch checks

Required status đề xuất: `CI result` (foundation, frontend, admin, backend và container smoke nếu áp dụng). Kiểm tra branch protection ở GitHub Settings trước release; hiện chưa tự cấu hình required checks. PR failing không merge. Pipeline container dùng `.env` random riêng CI, migration/bootstrap, local-only captcha; không dùng secret production. Failure artifacts không chứa .env/cookie/password. Render/Neon chưa được tạo hoặc deploy trong phiên nền tảng.
