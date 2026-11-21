# Nội dung demo — T003

Người dùng cho phép giả định mọi dữ liệu thiếu và chỉnh sửa ở Admin. Đây là inventory đầu vào, chưa phải nội dung bán thật.

| Nội dung | Nguồn / trạng thái | Nơi sửa ở Admin |
| --- | --- | --- |
| Logo | blue-sky-group-logo.jpg, người dùng cung cấp | Media/thông tin công ty |
| Poster Ninh Chữ | docs/assets/ninh-chu-poster-reference.png, nguyên bản người dùng | Tour/poster, revision/kiểm duyệt |
| Ninh Chữ 2 ngày 1 đêm | Giá quảng cáo1.650.000đ từ poster; itinerary xem poster-reference.md | Tour/lịch trình/giá departure |
| Lịch/giá trẻ em/chỗ | Giả định departure tương lai,30 chỗ, adult1650000/child1155000/infant0; chưa xác minh | Lịch khởi hành/giá/tuổi/chỗ |
| Quy định | Hold20 phút, cọc30%, trả dư7 ngày trước đi; huỷ do nhân viên xét | Cài đặt đặt tour/tour |
| Công ty | Tên Blue Sky Group; liên hệ trong poster là tham chiếu; email chưa có | Thông tin công ty; không bịa giấy phép |
| SLA | Giả định1 ngày làm việc | Cấu hình SLA |
| Blog/FAQ/đánh giá | Nội dung mẫu có nhãn; không tuyên bố review thật | CMS có duyệt trước public |

Chưa có quyền ảnh thành phần của poster được xác minh độc lập; dùng tham chiếu/demo, kiểm lại trước public chính thức. Mọi giá trị giả định có is_demo, Admin chỉnh được; thông tin trống không render thành dữ liệu thật. Dữ liệu nghiệp vụ phải là trường có cấu trúc, không OCR poster làm nguồn tiền.
