# QA nâng cấp giao diện ưu tiên Zalo — 23/11/2026

Phạm vi đã duyệt: du lịch rực rỡ, Zalo chính, public và Admin. Không backend/database/dependency mới. Không chạy app/service local.

## Kiểm chứng local

- T703l–T703n: frontend lint/types/unit và build Webpack đạt; warning stylesheet kit đã có. T703p thêm 2 unit consent/contact, tổng 6 đạt. Typecheck cả hai app đạt.
- T703o: Admin lint/types/5 unit/build đạt; ca mới mở đúng fieldset và giữ dữ liệu khi save thất bại. Reports chỉ trình bày response hiện có, không số liệu mới.
- GATE taste: đã khởi tạo, chưa chạy browser checker/perf local vì không có quyền khởi động app. Không GATE PASS, không tuyên bố 60fps/Core Web Vitals hoặc tăng CTR. Waiver/ngoại lệ dự án theo design-style.md vẫn áp dụng.

## Browser và bằng chứng

- PR public #21: CI Linux build mặc định, PostgreSQL và Chromium/Firefox/WebKit. Kiểm hero không quảng cáo tour cụ thể, video lặp/fallback, font tiếng Việt, keyboard/menu, ba kênh liên hệ, 320–1440px, detail và form không bị thanh đáy che. [CI 45208665160](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45208665160) đạt 6/6 jobs: Chromium18, Firefox6, WebKit6; live video chỉ Chromium (skip có lý do trên engine phụ).
- Mở rộng analytics kiểm link hoạt động trước consent, contact_click/select_item sau consent, dừng sau từ chối; không điện thoại/href trong event.
- PR Admin #22: [CI tích hợp 45210903923](https://github.com/EthanLe-dev/blue-sky-group-website/actions/runs/45210903923) đạt 6/6, gồm menu dọc/Sheet/focus, lỗi/empty/loading, lưu/reload, form grouped, report/export và ảnh desktop/mobile. Bộ public18/Firefox6/WebKit6 đạt. Bản cuối nâng nhãn lên16px có build local và chờ CI trước merge.
- Ảnh CI dùng dữ liệu demo. Cần người dùng duyệt trực quan; screenshot CI không thay UAT độc lập và không có dữ liệu CTR thực tế.

- Đã xem ảnh Trang chủ 1440/390 và contact từ CI. Phát hiện CSS cũ flex-direction trên mọi link card và body padding của fixed header cũ; T703p sửa ở lượt tích hợp, thêm assertion nút nằm ngang/không phình và header bắt đầu y=0. Không chỉ dựa vào test PASS để nghiệm thu thị giác.


## Hợp đồng đo click và bàn giao

- `contact_click`: `contact_channel` chỉ zalo/facebook/phone, `cta_position` chỉ các vị trí đã định nghĩa; `item_id` tùy chọn phải UUID. `select_item`: list home/catalog, item UUID. Không đưa href, hotline, slug chứa nội dung tự do hoặc nội dung form vào payload. Listener được cleanup, không ngăn link mở trang đích; không có consent thì không đo.
- Metadata, nhãn và chú thích mới dùng từ 16px; chữ chính 17px, Nunito Sans Vietnamese local. Không thêm dependency hoặc tải font từ CDN.
- Chỉ tích hợp develop sau CI xanh. Không tự deploy main, không khởi động app local. Người dùng tự chạy `bluesky run` để duyệt. Hiệu quả CTR cần đo sau khi có traffic thật; bản này không cam kết tỷ lệ tăng.

- Đã xem ảnh Admin sidebar/settings và báo cáo mobile của CI tích hợp: màu/menu, metric và nhãn rõ; bổ sung nền workspace cho màn cấu hình không dùng class workspace, và toolbar xếp dọc mobile để nút export không bị ép thành nhiều dòng. Có assertion kích thước nút export.
