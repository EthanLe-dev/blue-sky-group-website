# Phiếu nghiệm thu MVP

Dùng bản demo có nhãn và các địa chỉ `.example`. Kiểm tra trên điện thoại của người dùng và trên desktop. Mỗi ca ghi người thử, thiết bị, SHA, kết quả và bước tái hiện nếu lỗi. CI giúp đối chiếu hành vi chức năng. Kết quả với người dùng thật và URL cloud ghi riêng.

| # | Người thử | Cách thử / kết quả cần thấy | Bằng chứng có thể lấy trong repo | Việc còn chờ |
| --- | --- | --- | --- | --- |
| 1 | Khách | Tìm tour, mở menu trên màn nhỏ, thao tác bàn phím, poster mở/đóng | Kiểm tra responsive và screenshot Chromium/Firefox/WebKit | Đánh giá thẩm mỹ, thiết bị thật |
| 2 | Khách | Chọn lịch và nhóm tuổi, tạo đơn, xác nhận rồi tra cứu | Playwright luồng toàn bộ và cọc/phần còn lại | Người thử độc lập |
| 3 | Khách | QR ghi rõ giả lập, mất phản hồi thử lại mà chỉ có một khoản thu | Playwright và kiểm thử PostgreSQL | URL demo |
| 4 | Khách / Admin | Email/SMS là preview, thấy nhật ký đúng quyền | Outbox integration và Playwright | Kiểm lại trên URL demo |
| 5 | Kế toán | Tải Excel và CSV, đối chiếu tổng trên bảng và xuất file | Ledger PostgreSQL, Playwright tải file, Excel đọc fixture M6 | Kiểm với dữ liệu demo được duyệt |
| 6 | Khách / Sales | Tải PDF lịch trình và từng phiên bản báo giá, đọc dấu tiếng Việt | Playwright và PDF M6 đã render | Nội dung báo giá demo duyệt |
| 7 | Khách | Mở bản đồ ở trang tour và Liên hệ, điểm mẫu có nhãn | Playwright với tile giả, marker thật | Tọa độ thật và tile trên URL demo |
| 8 | Người quản trị GA4 | Cho phép analytics, xem tour, đặt demo và gửi yêu cầu | Sự kiện giả trong Playwright | DebugView/GA4 thật và Search Console cần tài khoản |
| 9 | Admin / Sales / Kế toán | Sửa tour, đặt tại quầy, hoàn tiền đúng quyền, export và truy cập bị chặn | Playwright và RBAC integration | Nhân viên dùng thử theo vai trò |
| 10 | Khách | Turnstile hợp lệ cho phép gửi, token sai bị từ chối | Adapter test/local và kiểm tra form | Widget thật trên hostname công khai |
| 11 | Chủ dự án | Mở URL public/Admin, khởi động lạnh, lưu file TMP qua R2 đúng quyền | Cấu hình và script smoke theo SHA | Tài khoản host, URL và lần deploy thật |

Khi có URL, mời ít nhất một người mua tour thuộc nhóm 30+, một đại diện trường/doanh nghiệp và một người có vai trò Admin dùng thử bằng dữ liệu mẫu. Ghi bước họ mất thời gian hoặc không tự hoàn thành, sửa lỗi cản trở, rồi xác nhận chỉ xuất bản các ảnh/nội dung được duyệt. Tải file thất bại, chuyển trạng thái tiền sai, lộ dữ liệu private hoặc form gửi được khi captcha sai là lỗi chặn phát hành.

T703 còn các điểm thẩm mỹ trong [báo cáo trước](ui-qa-m2-m3.md). Responsive/keyboard đạt trên CI không thay thế quyết định duyệt giao diện của người dùng.
