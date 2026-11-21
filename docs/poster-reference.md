# Poster tham chiếu — Ninh Chữ

Nhận từ người dùng ngày22/11/2026, dùng để thiết kế website và quy trình quản lý poster. File gốc được giữ nguyên tại [poster Ninh Chữ](assets/ninh-chu-poster-reference.png), kích thước1051×1497px, tỉ lệ khoảng0.702:1 (dọc, gần A-series). Không coi đây là tỉ lệ bắt buộc cho mọi poster.

## Thông tin quan sát từ ảnh

| Trường | Nội dung trên poster | Cách sử dụng |
| --- | --- | --- |
| Thương hiệu | Blue Sky Group — Kiến tạo hành trình | Tham chiếu nhận diện/copy |
| Tour | Ninh Chữ | Tên tour mẫu |
| Thời lượng | 2 ngày 1 đêm | duration_days=2, duration_nights=1 |
| Giá quảng cáo | 1.650.000đ | Chỉ giá hiển thị trên ảnh; chưa biết đối tượng áp dụng/ngày đi/dịch vụ gồm. Không tự coi là giá adult cho mọi departure |
| Điện thoại | 0989 876 847 | Dữ liệu do poster cung cấp, cần xác nhận hiện hành trước xuất bản thật; chưa suy ra tài khoản Zalo |
| Địa chỉ | 183/61 Nguyễn Văn Khối, phường Thông Tây Hội, TP. Hồ Chí Minh | Chép theo ảnh, chưa xác minh địa chỉ hiện hành/toạ độ |
| Ngày 1 | Ăn sáng; Miếu Quan Âm sông Mao; Trung tâm trưng bày Văn hoá Chăm; ăn trưa; làng gốm Bàu Trúc; Trường Sơn Cổ Tự; nhận phòng; ăn tối | Dàn ý itinerary mẫu, chưa có giờ/chi tiết dịch vụ |
| Ngày 2 | Biển vô cực Mỹ Hòa; VQG Núi Chúa – Hang Rái, khu cứu hộ rùa biển; vườn nho; đồng muối Phương Cựu; Sắc Tứ Kim Sơn Tự; Linh đài Đức Mẹ Fatima Bình Chính; ăn trưa, khởi hành về | Dàn ý chép từ ảnh, cần biên tập/xác minh tên địa điểm trước xuất bản |

Chưa có trong ảnh: nơi đón, phương tiện, khách sạn cụ thể, ngày khởi hành, chỗ còn, giá trẻ em/em bé, phụ thu, bao gồm/không gồm, điều kiện huỷ/hoàn. Cho phép tạo fixture demo có nhãn theo quyết định người dùng; không suy diễn những trường này thành sự thật.

## Nhận diện và cách đưa lên web

Poster dùng xanh navy cho chữ/khối chính, xanh biển cho nền, vàng sáng cho nhấn giá, trắng/kem cho phần lịch trình. Đây là màu quan sát; cần lấy mã màu từ logo/source design để chốt brand tokens. Chữ script lớn dùng cho tên điểm đến trên ảnh; UI web dùng Be Vietnam Pro cho khả năng đọc, không mô phỏng font script ở body/form.

Cấu trúc poster: logo → tên tour → thời lượng → giá → ảnh trải nghiệm → lịch trình2 ngày → liên hệ. Trang web giữ phân cấp tên/thời lượng/giá nhưng biểu diễn bằng HTML, có CTA và ngày khởi hành lấy từ DB. Poster là asset quảng bá có viewer xem toàn bộ/zoom; card dùng ảnh cover riêng, không crop nửa poster làm khách mất thông tin.

Không cần dựng lại poster bằng CSS hay bóc hình từ collage khi chưa có ảnh nguồn. Các poster mới do team thiết kế cung cấp có thể khác kích thước; lưu dimensions và kiểm preview responsive. Ảnh chứa chữ nên giữ bản gốc để zoom, tạo bản nhẹ cho preview mà chữ vẫn đọc được. Không đưa ảnh1051px lên hero full-width1440px rồi khẳng định chất lượng gốc đáp ứng.

Nguồn asset: người dùng cung cấp để tham chiếu dự án; chưa có hồ sơ quyền từng ảnh thành phần. Trước công khai như nội dung chính thức, xác nhận quyền dùng với team. Không tự tìm/đổi địa danh, giá hoặc số điện thoại trong ảnh.
