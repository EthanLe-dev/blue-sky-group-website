# TÀI LIỆU YÊU CẦU WEBSITE CÔNG TY DU LỊCH

| Thông tin              | Nội dung                                                   |
| ---------------------- | ---------------------------------------------------------- |
| **Dự án**              | Website bán tour và chăm sóc khách hàng của Blue Sky Group |
| **Gửi đến**            | Bộ phận IT                                                 |
| **Người lập**          | Lê Ngọc Quý - BA                                           |
| **Phiên bản**          | 0.2 (bổ sung tech stack & tham chiếu MVP)                  |
| **Ngày lập**           | 21/11/2026 (cập nhật: 22/11/2026)                          |
| **Trạng thái**         | Chờ IT phản hồi, đánh giá khả thi và ước lượng             |
| **Tài liệu liên quan** | `mvp-requirements.md` (Yêu cầu MVP), `architecture.md` (Kiến trúc hệ thống), `database-schema.md` (CSDL) |

**Quy ước mức ưu tiên:** **M** = Bắt buộc (Must), **S** = Nên có (Should), **C** = Có thì tốt (Could).

---

## 1. Tổng quan

### 1.1 Bối cảnh và mục tiêu

Công ty cần một website du lịch hiện đại, đóng vai trò **kênh bán tour chính** và **bộ mặt chuyên nghiệp** của thương hiệu. Website không chỉ để giới thiệu tour mà phải **thuyết phục khách đặt tour ngay trong lần truy cập**, đồng thời thể hiện được sự chu đáo, chuyên nghiệp và chất lượng dịch vụ của công ty.

### 1.2 Mục tiêu kinh doanh

| Mục tiêu                                      | Chỉ số đo lường (đề xuất, cần thống nhất lại)                       |
| --------------------------------------------- | ------------------------------------------------------------------- |
| Tăng tỉ lệ đặt tour sau khi vào web           | Tỉ lệ chuyển đổi (lượt truy cập → đặt tour hoặc gửi yêu cầu tư vấn) |
| Giảm bỏ dở giữa chừng khi đặt                 | Tỉ lệ hoàn tất quy trình đặt tour                                   |
| Thu hút khách đoàn (trường học, doanh nghiệp) | Số yêu cầu báo giá đoàn mỗi tháng                                   |
| Giảm tải cho nhân viên tư vấn                 | Tỉ lệ khách tự tra cứu và tự đặt được                               |
| Tăng độ tin cậy thương hiệu                   | Thời gian ở lại trang, tỉ lệ thoát, lượt đánh giá                   |

### 1.3 Phạm vi

**Trong phạm vi:**

- Website công khai (giao diện khách hàng) hiển thị tốt trên máy tính, máy tính bảng và điện thoại.
- Hệ thống quản trị nội bộ (back office) cho nhân viên.
- Đặt tour, thanh toán trực tuyến, quản lý đơn.
- Chức năng riêng cho khách đoàn: trường học, doanh nghiệp.
- Thông báo email/SMS, báo cáo cơ bản.

**Ngoài phạm vi (giai đoạn này):**

- Ứng dụng di động (native app).
- Bán vé máy bay, đặt khách sạn lẻ như một đại lý trực tuyến (OTA).
- Hệ thống kế toán và nhân sự đầy đủ (chỉ tích hợp hoặc xuất dữ liệu khi cần).

---

## 2. Đối tượng người dùng

| Nhóm                                                  | Mô tả                                                                                     | Nhu cầu chính                                                           |
| ----------------------------------------------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| **A. Khách lẻ / gia đình** (đa số từ 30 tuổi trở lên) | Quyết định cẩn trọng, quan tâm an toàn, lịch trình rõ ràng, giá minh bạch, uy tín công ty | Dễ đọc, dễ tìm, thông tin đầy đủ, có người tư vấn khi cần               |
| **B. Trường học** (giáo viên, ban phụ huynh)          | Tour ngoại khoá, cần đảm bảo an toàn, báo giá theo số lượng học sinh, hồ sơ pháp lý       | Yêu cầu báo giá nhanh, tour mẫu, hồ sơ năng lực, bảo hiểm, quy trình rõ |
| **C. Doanh nghiệp** (HR, hành chính)                  | Team building, du lịch nhân viên, quan tâm tuỳ biến chương trình và hợp đồng              | Gói tuỳ biến, báo giá, hợp đồng và hoá đơn, quản lý danh sách đoàn      |
| **D. Nhân viên nội bộ**                               | Tư vấn/sales, chăm sóc khách hàng, biên tập nội dung, kế toán, quản trị                   | Công cụ quản lý đơn, tour, giá, nội dung, báo cáo                       |

**Hệ quả thiết kế:** vì khách chủ yếu từ 30 tuổi trở lên, giao diện cần **chữ dễ đọc, nút bấm lớn, luồng đặt tour đơn giản, ít bước, thông tin giá và điều khoản rõ ràng**, đồng thời hỗ trợ khách tìm đến con người (hotline, Zalo) bất cứ lúc nào.

---

## 3. Nguyên tắc trải nghiệm người dùng (định hướng xuyên suốt)

Đây là **ưu tiên hàng đầu** của dự án, cần được ưu tiên ngang với chức năng.

1. **Sống động, cuốn hút:** hình ảnh và video chất lượng cao, bố cục giàu cảm xúc, kể được câu chuyện điểm đến.
2. **Mượt mà:** tải nhanh, chuyển trang và hiệu ứng mượt, không giật, không nhảy bố cục.
3. **Đơn giản để đặt:** từ lúc chọn tour đến lúc thanh toán càng ít bước càng tốt; luôn thấy rõ nút hành động tiếp theo.
4. **Minh bạch:** giá gồm/không gồm gì, chính sách huỷ/đổi, lịch trình từng ngày đều rõ ràng, không có chi phí ẩn.
5. **Chu đáo, chuyên nghiệp:** website phải toát lên sự quan tâm: tư vấn kịp thời, thông tin chuẩn bị chuyến đi, xác nhận và nhắc nhở đúng lúc, giao tiếp lịch sự.
6. **Đáng tin:** hiển thị giấy phép kinh doanh lữ hành, đánh giá thật, hình ảnh đoàn thật, thông tin liên hệ rõ ràng, thanh toán an toàn.

---

## 4. Yêu cầu chức năng

### 4.1 Khám phá và tìm tour (khách truy cập)

| ID    | Yêu cầu                                                                                                                                                                                                            | Ưu tiên |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- |
| FR-01 | **Trang chủ:** banner/hero (ảnh hoặc video), tour nổi bật, điểm đến phổ biến, ưu đãi, đánh giá khách hàng, cam kết dịch vụ, tin tức/cẩm nang, ô tìm kiếm nổi bật                                                   | M       |
| FR-02 | **Danh mục tour** theo: trong nước/quốc tế, điểm đến, loại hình (nghỉ dưỡng, khám phá, hành hương, team building, ngoại khoá, du lịch nhân viên...)                                                                | M       |
| FR-03 | **Tìm kiếm** theo điểm đến, ngày khởi hành, ngân sách, số ngày, số người; có gợi ý khi gõ (autocomplete)                                                                                                           | M       |
| FR-04 | **Bộ lọc và sắp xếp** theo giá, thời gian, ngày đi, phương tiện, mức độ vất vả, đánh giá; giữ nguyên bộ lọc khi quay lại                                                                                           | M       |
| FR-05 | **Trang chi tiết tour:** thư viện ảnh/video, tóm tắt (thời gian, nơi đi, phương tiện, khách sạn), lịch trình theo ngày, dịch vụ bao gồm/không bao gồm, chính sách huỷ/đổi, lưu ý, bản đồ hành trình, FAQ, đánh giá | M       |
| FR-06 | **Lịch khởi hành và giá** hiển thị trực quan (lịch/bảng): ngày đi, giá theo từng đối tượng, số chỗ còn, trạng thái (còn chỗ / sắp hết / hết chỗ)                                                                   | M       |
| FR-07 | **Thanh đặt tour cố định** (sticky) trên trang chi tiết, luôn thấy giá và nút "Đặt tour"/"Nhận tư vấn"                                                                                                             | M       |
| FR-08 | **Tải lịch trình PDF** để xem và chia sẻ cho gia đình/công ty                                                                                                                                                      | S       |
| FR-09 | **So sánh** 2–3 tour                                                                                                                                                                                               | C       |
| FR-10 | **Yêu thích / lưu tour** (không cần đăng nhập cũng lưu tạm được)                                                                                                                                                   | S       |
| FR-11 | **Tour đã xem gần đây** và gợi ý tour liên quan                                                                                                                                                                    | S       |
| FR-12 | **Chia sẻ tour** qua Zalo, Facebook, sao chép liên kết                                                                                                                                                             | S       |

### 4.2 Đặt tour và thanh toán

| ID    | Yêu cầu                                                                                                                                          | Ưu tiên |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------- |
| FR-20 | **Đặt tour không bắt buộc tạo tài khoản** (guest checkout); khuyến khích tạo tài khoản sau khi đặt                                               | M       |
| FR-21 | Luồng đặt tour gọn, **tối đa 3–4 bước** (chọn ngày/số người → thông tin khách → xác nhận & thanh toán), có thanh tiến trình, có thể quay lại sửa | M       |
| FR-22 | Nhập số lượng theo đối tượng (người lớn, trẻ em, em bé) với **giá tính tự động và hiển thị tức thì**                                             | M       |
| FR-23 | Nhập thông tin liên hệ và danh sách hành khách; có yêu cầu đặc biệt (ăn chay, người cao tuổi, trẻ nhỏ...)                                        | M       |
| FR-24 | **Giữ chỗ có thời hạn** (ví dụ 15–30 phút) trong lúc khách thanh toán; hết hạn tự nhả chỗ                                                        | M       |
| FR-25 | **Chống bán vượt số chỗ** khi nhiều khách đặt cùng lúc                                                                                           | M       |
| FR-26 | Mã giảm giá / khuyến mãi; hiển thị rõ tổng tiền và khoản giảm                                                                                    | S       |
| FR-27 | **Thanh toán trực tuyến** qua cổng thanh toán phổ biến tại Việt Nam (thẻ nội địa/quốc tế, ví điện tử, QR chuyển khoản)                           | M       |
| FR-28 | Tuỳ chọn **thanh toán toàn bộ hoặc đặt cọc**, kèm hạn thanh toán phần còn lại; chuyển khoản thủ công có nhập xác nhận                            | M       |
| FR-29 | Trang xác nhận đặt tour và **email/SMS xác nhận** kèm mã đơn, tóm tắt hành trình, hướng dẫn tiếp theo                                            | M       |
| FR-30 | **Lưu tiến trình** nếu khách thoát giữa chừng và **nhắc lại** qua email (nếu khách đồng ý)                                                       | S       |
| FR-31 | **Tra cứu đơn** bằng mã đơn và số điện thoại/email (cho khách không có tài khoản)                                                                | M       |
| FR-32 | Yêu cầu **huỷ/đổi** tour trực tuyến theo chính sách; hiển thị mức phí dự kiến                                                                    | S       |
| FR-33 | Xuất **hoá đơn / chứng từ thanh toán**; hỗ trợ thông tin xuất hoá đơn công ty                                                                    | S       |

### 4.3 Dịch vụ cho khách đoàn (trường học, doanh nghiệp)

| ID    | Yêu cầu                                                                                                                                                               | Ưu tiên |
| ----- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| FR-40 | **Khu vực riêng** "Dành cho Trường học" và "Dành cho Doanh nghiệp" với nội dung, tour mẫu và cam kết riêng                                                            | M       |
| FR-41 | **Biểu mẫu yêu cầu báo giá đoàn**: loại đoàn, số lượng người, ngày dự kiến, điểm đến, ngân sách, yêu cầu đặc biệt; gửi xong có xác nhận và thời gian phản hồi cam kết | M       |
| FR-42 | **Tour mẫu và gói tuỳ biến**: ngoại khoá học đường, team building, du lịch nhân viên, gala dinner, hội nghị kết hợp nghỉ dưỡng                                        | M       |
| FR-43 | **Hồ sơ năng lực** để tải về: giấy phép, bảo hiểm, đối tác, dự án tiêu biểu, khách hàng đã phục vụ                                                                    | M       |
| FR-44 | **Cam kết an toàn** cho học sinh (nhân sự đi cùng, bảo hiểm, quy trình xử lý sự cố) hiển thị nổi bật cho nhóm trường học                                              | M       |
| FR-45 | Hình ảnh, video và **đánh giá/giới thiệu từ khách đoàn** (trường, công ty) đã sử dụng dịch vụ                                                                         | S       |
| FR-46 | **Tải lên danh sách thành viên đoàn** (mẫu Excel) cho khách đoàn đã chốt                                                                                              | S       |
| FR-47 | Theo dõi **trạng thái yêu cầu** (đã nhận → đang soạn báo giá → đã gửi báo giá → chốt) cho người yêu cầu                                                               | C       |
| FR-48 | Tải báo giá/hợp đồng/lịch trình dạng PDF; ký hoặc xác nhận trực tuyến                                                                                                 | C       |

### 4.4 Tài khoản khách hàng

| ID    | Yêu cầu                                                                      | Ưu tiên              |
| ----- | ---------------------------------------------------------------------------- | -------------------- |
| FR-50 | Đăng ký/đăng nhập bằng email hoặc số điện thoại (OTP); đăng nhập Google/Zalo | S                    |
| FR-51 | Quên/đặt lại mật khẩu an toàn                                                | M (nếu có tài khoản) |
| FR-52 | Trang "Đơn của tôi": lịch sử, trạng thái, hoá đơn, tải lịch trình/voucher    | S                    |
| FR-53 | Hồ sơ lưu sẵn thông tin hành khách thường đi cùng để đặt lần sau nhanh hơn   | C                    |
| FR-54 | Viết đánh giá tour sau chuyến đi                                             | S                    |

### 4.5 Hỗ trợ và thể hiện sự chu đáo

| ID    | Yêu cầu                                                                                                                                                | Ưu tiên |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- |
| FR-60 | Hotline, Zalo, email **hiển thị nhất quán** ở mọi trang; nút liên hệ nhanh cố định trên điện thoại                                                     | M       |
| FR-61 | **Yêu cầu gọi lại** (nhập số điện thoại, chọn khung giờ), có cam kết thời gian phản hồi                                                                | M       |
| FR-62 | **Chat trực tuyến** (hoặc tích hợp Zalo OA/Messenger), có giờ làm việc rõ ràng và chế độ để lại lời nhắn ngoài giờ                                     | S       |
| FR-63 | **Câu hỏi thường gặp (FAQ)** theo chủ đề: đặt tour, thanh toán, huỷ/đổi, visa, trẻ em, người cao tuổi, bảo hiểm                                        | M       |
| FR-64 | **Cẩm nang chuẩn bị chuyến đi** (checklist hành lý, thời tiết, lưu ý văn hoá, giấy tờ cần thiết) hiển thị trên trang tour và gửi qua email sau khi đặt | S       |
| FR-65 | **Email/SMS chăm sóc theo hành trình**: xác nhận đặt → nhắc thanh toán → thông tin trước ngày đi → cảm ơn và xin đánh giá sau chuyến                   | S       |
| FR-66 | Trang **"Về chúng tôi / Vì sao chọn chúng tôi"**: giấy phép lữ hành, số năm hoạt động, số khách phục vụ, đội ngũ, giải thưởng, đối tác                 | M       |
| FR-67 | Trang chính sách rõ ràng: điều khoản đặt tour, thanh toán, huỷ/đổi, bảo mật dữ liệu                                                                    | M       |

### 4.6 Nội dung và SEO

| ID    | Yêu cầu                                                                                                                                                  | Ưu tiên |
| ----- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| FR-70 | **Blog / cẩm nang du lịch / tin tức** có danh mục, thẻ, bài liên quan                                                                                    | M       |
| FR-71 | Trang điểm đến (landing page) giới thiệu điểm đến và liệt kê tour liên quan                                                                              | S       |
| FR-72 | Thư viện ảnh/video có phân loại; trang "Khoảnh khắc khách hàng"                                                                                          | S       |
| FR-73 | **SEO kỹ thuật:** URL thân thiện, thẻ tiêu đề/mô tả tuỳ chỉnh từng trang, sitemap, dữ liệu có cấu trúc (schema), thẻ chia sẻ mạng xã hội, tốc độ tải tốt | M       |
| FR-74 | Đăng ký nhận bản tin/ưu đãi (có đồng ý rõ ràng)                                                                                                          | S       |

### 4.7 Hệ thống quản trị (back office)

| ID    | Yêu cầu                                                                                                          | Ưu tiên |
| ----- | ---------------------------------------------------------------------------------------------------------------- | ------- |
| FR-80 | Quản lý **tour**: nội dung, ảnh/video, lịch trình, dịch vụ gồm/không gồm, trạng thái hiển thị, nhân bản tour     | M       |
| FR-81 | Quản lý **lịch khởi hành và số chỗ**; tạo hàng loạt ngày khởi hành; đóng/mở bán                                  | M       |
| FR-82 | Quản lý **giá** theo đối tượng, theo mùa/ngày lễ, phụ thu; quản lý khuyến mãi và mã giảm giá có thời hạn         | M       |
| FR-83 | Quản lý **đơn đặt tour**: xem, lọc, cập nhật trạng thái, ghi chú nội bộ, thêm/sửa hành khách, huỷ, hoàn tiền     | M       |
| FR-84 | **Tạo đơn thủ công** cho khách đặt qua điện thoại/quầy (để mọi đơn nằm cùng một hệ thống và dùng chung số chỗ)   | M       |
| FR-85 | Quản lý **yêu cầu báo giá đoàn**: tiếp nhận, phân công nhân viên, theo dõi trạng thái, ghi chú, đính kèm báo giá | M       |
| FR-86 | Quản lý **yêu cầu tư vấn / gọi lại** và nhắc nhở nhân viên xử lý đúng hạn                                        | S       |
| FR-87 | Quản lý **nội dung**: bài viết, banner, trang tĩnh, FAQ, đánh giá (duyệt trước khi hiển thị)                     | M       |
| FR-88 | Quản lý **khách hàng**: thông tin, lịch sử đặt, ghi chú, danh sách nhận bản tin                                  | S       |
| FR-89 | **Phân quyền theo vai trò** và nhật ký thao tác (ai sửa gì, khi nào)                                             | M       |
| FR-90 | Mẫu email/SMS có thể chỉnh sửa nội dung mà không cần lập trình viên                                              | S       |

### 4.8 Báo cáo và phân tích

| ID    | Yêu cầu                                                                                      | Ưu tiên |
| ----- | -------------------------------------------------------------------------------------------- | ------- |
| FR-95 | Báo cáo doanh thu và số đơn theo ngày/tháng, theo tour, theo nguồn khách                     | M       |
| FR-96 | Báo cáo tỉ lệ lấp đầy chỗ theo lịch khởi hành                                                | S       |
| FR-97 | Phễu chuyển đổi: xem tour → bắt đầu đặt → thanh toán → hoàn tất; điểm khách bỏ dở nhiều nhất | S       |
| FR-98 | Báo cáo yêu cầu báo giá đoàn và tỉ lệ chốt                                                   | S       |
| FR-99 | Xuất báo cáo ra Excel/CSV                                                                    | M       |

---

## 5. Yêu cầu giao diện và trải nghiệm chi tiết

### 5.1 Phù hợp nhóm khách từ 30 tuổi trở lên

| ID    | Yêu cầu                                                                                                                                          | Ưu tiên |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------- |
| UX-01 | **Cỡ chữ nội dung tối thiểu 16px** (ưu tiên 17–18px trên mobile), độ giãn dòng thoáng, không dùng chữ quá mảnh                                   | M       |
| UX-02 | **Độ tương phản màu** đạt tối thiểu WCAG 2.1 mức AA; không chỉ dựa vào màu để truyền đạt thông tin                                               | M       |
| UX-03 | **Nút và vùng chạm lớn** (tối thiểu khoảng 44×44px), khoảng cách đủ rộng để không bấm nhầm                                                       | M       |
| UX-04 | Điều hướng **đơn giản, nhất quán**, nhãn menu bằng từ ngữ quen thuộc; luôn biết "mình đang ở đâu" (breadcrumb)                                   | M       |
| UX-05 | Biểu mẫu **ít trường nhất có thể**, nhãn rõ ràng, thông báo lỗi dễ hiểu và chỉ ra cách sửa, giữ lại dữ liệu đã nhập                              | M       |
| UX-06 | Luôn có **đường tắt tới con người** (hotline, Zalo, gọi lại) ở các điểm khách dễ phân vân: trang chi tiết tour, trang đặt tour, trang thanh toán | M       |
| UX-07 | Không dùng cửa sổ bật lên (popup) gây phiền; không tự phát âm thanh/video có tiếng                                                               | M       |
| UX-08 | Hỗ trợ **phóng to chữ** theo cài đặt trình duyệt/thiết bị mà bố cục không vỡ                                                                     | S       |

### 5.2 Sống động, cuốn hút và mượt mà

| ID    | Yêu cầu                                                                                                                                                           | Ưu tiên |
| ----- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| UX-10 | Phong cách thị giác chuyên nghiệp, ấm áp, giàu cảm xúc; **ảnh/video thực tế, chất lượng cao** (không dùng ảnh stock gượng ép); bộ nhận diện thương hiệu nhất quán | M       |
| UX-11 | **Hero** trang chủ bằng ảnh/video nền nhẹ; video phải tự tối ưu và **không làm chậm trang**                                                                       | M       |
| UX-12 | **Chuyển động tinh tế có mục đích** (hiệu ứng khi cuộn, hover, chuyển trang) giúp dẫn mắt, không phô trương; tôn trọng cài đặt "giảm chuyển động" của người dùng  | S       |
| UX-13 | **Thư viện ảnh/video** trên trang tour: xem toàn màn hình, vuốt trên điện thoại, tải ảnh dần (lazy load)                                                          | M       |
| UX-14 | **Bản đồ hành trình** trực quan và lịch trình dạng dòng thời gian dễ đọc                                                                                          | S       |
| UX-15 | **Trạng thái chờ mượt** (skeleton/loading) thay vì màn hình trắng; phản hồi tức thì khi bấm                                                                       | M       |
| UX-16 | Thiết kế **mobile-first**: phần lớn khách sẽ xem bằng điện thoại, thao tác một tay thuận tiện                                                                     | M       |

### 5.3 Tăng tỉ lệ đặt tour

| ID    | Yêu cầu                                                                                                                                | Ưu tiên |
| ----- | -------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| UX-20 | **Lời kêu gọi hành động (CTA) rõ ràng** và nhất quán: "Xem lịch khởi hành", "Đặt tour", "Nhận tư vấn miễn phí"                         | M       |
| UX-21 | Hiển thị **"giá từ ..."** ngay ở danh sách và giá chi tiết theo ngày ở trang tour; không để khách phải đoán                            | M       |
| UX-22 | **Dấu hiệu tin cậy** gần nút đặt và ở bước thanh toán: giấy phép, đánh giá, thanh toán an toàn, chính sách huỷ rõ ràng, cam kết hỗ trợ | M       |
| UX-23 | **Thông tin khan hiếm/khẩn cấp phải trung thực** (ví dụ "còn 3 chỗ" lấy từ số liệu thật); tuyệt đối không tạo số liệu giả              | M       |
| UX-24 | **Đánh giá và hình ảnh khách thật** hiển thị ngay trên trang tour                                                                      | M       |
| UX-25 | Khách không đặt ngay vẫn có đường lựa chọn nhẹ: lưu tour, nhận lịch trình PDF, yêu cầu gọi lại                                         | S       |
| UX-26 | Giảm ma sát thanh toán: đặt không cần tài khoản, tự điền thông tin quen thuộc, nhiều phương thức thanh toán phổ biến                   | M       |
| UX-27 | Hiển thị **tổng tiền cuối cùng** ở mọi bước, không xuất hiện phí lạ ở bước cuối                                                        | M       |

### 5.4 Thể hiện sự chu đáo và chuyên nghiệp

| ID    | Yêu cầu                                                                                                                    | Ưu tiên |
| ----- | -------------------------------------------------------------------------------------------------------------------------- | ------- |
| UX-30 | Giọng văn **lịch sự, ấm áp, rõ ràng**; thông tin đầy đủ nhưng không rối                                                    | M       |
| UX-31 | Mọi biểu mẫu/lời nhắn có **cam kết thời gian phản hồi** và xác nhận đã nhận                                                | M       |
| UX-32 | **Email/SMS giao dịch** được thiết kế đẹp, đúng thương hiệu, nội dung hữu ích (không chỉ thông báo khô khan)               | S       |
| UX-33 | Trang lỗi (404, lỗi thanh toán...) **thân thiện, có hướng dẫn tiếp theo** và thông tin hỗ trợ                              | M       |
| UX-34 | Khi có sự cố (hết chỗ, lỗi thanh toán), hệ thống **đề xuất phương án thay thế** (ngày khác, tour tương tự, liên hệ tư vấn) | S       |

---

## 6. Yêu cầu phi chức năng

### 6.1 Hiệu năng và tải

| ID     | Yêu cầu                                                                                                                                                                                   | Ưu tiên |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| NFR-01 | Hệ thống phục vụ **tối thiểu 100 người dùng đồng thời** (concurrent users) trong điều kiện sử dụng thực tế: duyệt tour, tìm kiếm, đặt tour, thanh toán                                    | M       |
| NFR-02 | Thiết kế có **dư địa tăng tải** (khuyến nghị chịu được 2–3 lần mức trên, tức khoảng 200–300 người dùng đồng thời) cho giai đoạn chạy chiến dịch/khuyến mãi                                | S       |
| NFR-03 | **Thời gian phản hồi** ở mức 100 người dùng đồng thời: trang công khai (tải từ máy chủ) ≤ 2 giây ở phân vị 95; API thông thường ≤ 500 ms ở phân vị 95; thao tác đặt tour/giữ chỗ ≤ 2 giây | M       |
| NFR-04 | **Chỉ số trải nghiệm web (Core Web Vitals)** trên điện thoại đạt mức "Tốt": LCP ≤ 2,5 giây, INP ≤ 200 ms, CLS ≤ 0,1                                                                       | M       |
| NFR-05 | Tối ưu ảnh/video: định dạng hiện đại, nhiều kích thước theo thiết bị, tải dần, dùng **CDN** cho tài nguyên tĩnh                                                                           | M       |
| NFR-06 | **Bộ nhớ đệm (cache)** cho các trang và dữ liệu đọc nhiều (danh sách tour, trang chi tiết)                                                                                                | M       |
| NFR-07 | **Kiểm thử tải** trước khi vận hành chính thức, có báo cáo kết quả và điểm nghẽn đã xử lý                                                                                                 | M       |

> **Lưu ý cho IT:** "100 người dùng đồng thời" cần được xác nhận cách hiểu (100 phiên đang hoạt động cùng lúc hay 100 yêu cầu mỗi giây) để thiết kế kịch bản kiểm thử. Đề xuất hiểu là **100 phiên người dùng hoạt động đồng thời** với hành vi duyệt/đặt tour thực tế. Xem thêm mục 12.

### 6.2 Tính sẵn sàng và vận hành

| ID     | Yêu cầu                                                                                                                                                          | Ưu tiên |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| NFR-10 | **Thời gian hoạt động mục tiêu ≥ 99,5%** mỗi tháng (đề xuất, IT đánh giá lại theo khả năng hạ tầng)                                                              | M       |
| NFR-11 | **Sao lưu tự động** hằng ngày; thử khôi phục định kỳ; lưu trữ bản sao lưu ở nơi tách biệt                                                                        | M       |
| NFR-12 | Mục tiêu khôi phục sau sự cố (đề xuất): mất dữ liệu tối đa (RPO) ≤ 24 giờ, thời gian khôi phục (RTO) ≤ 4 giờ; đơn đặt tour/thanh toán nên có mức bảo vệ chặt hơn | S       |
| NFR-13 | **Giám sát và cảnh báo** (tình trạng hệ thống, lỗi, tốc độ, lỗi thanh toán), nhật ký (log) tập trung                                                             | M       |
| NFR-14 | Có **môi trường thử nghiệm (staging)** tách biệt khỏi môi trường chạy thật; triển khai có quy trình, có thể quay lui (rollback)                                  | M       |
| NFR-15 | Có thể **nâng cấp mà không gián đoạn lâu** hoặc nằm ngoài giờ cao điểm                                                                                           | S       |

### 6.3 Bảo mật và bảo vệ dữ liệu

| ID     | Yêu cầu                                                                                                                                                                      | Ưu tiên |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| NFR-20 | Toàn bộ website chạy **HTTPS**; mã hoá dữ liệu nhạy cảm khi lưu trữ                                                                                                          | M       |
| NFR-21 | Phòng chống các lỗ hổng phổ biến theo **OWASP Top 10** (SQL injection, XSS, CSRF...); kiểm tra bảo mật trước khi vận hành                                                    | M       |
| NFR-22 | **Không lưu số thẻ thanh toán** trên hệ thống của công ty; thanh toán xử lý hoàn toàn qua cổng thanh toán đạt chuẩn                                                          | M       |
| NFR-23 | Mật khẩu được băm an toàn; hỗ trợ giới hạn số lần đăng nhập sai; **xác thực hai bước cho tài khoản quản trị**                                                                | M       |
| NFR-24 | **Phân quyền tối thiểu cần thiết**; nhật ký thao tác trên dữ liệu khách hàng và đơn hàng                                                                                     | M       |
| NFR-25 | Tuân thủ **quy định bảo vệ dữ liệu cá nhân hiện hành của Việt Nam** (thu thập có đồng ý, mục đích rõ ràng, quyền truy cập/xoá dữ liệu); pháp chế/IT xác nhận văn bản áp dụng | M       |
| NFR-26 | Chống spam và bot cho biểu mẫu (captcha thân thiện hoặc giải pháp tương đương)                                                                                               | S       |
| NFR-27 | Bản vá bảo mật và cập nhật thư viện định kỳ                                                                                                                                  | M       |

### 6.4 Khả năng tương thích, truy cập và bảo trì

| ID     | Yêu cầu                                                                                                                                        | Ưu tiên |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| NFR-30 | **Responsive** từ điện thoại đến màn hình lớn; hỗ trợ hai phiên bản mới nhất của Chrome, Safari, Edge, Firefox và trình duyệt trên iOS/Android | M       |
| NFR-31 | Đạt tiêu chuẩn truy cập **WCAG 2.1 mức AA** ở các luồng chính (đọc, tìm tour, đặt tour)                                                        | S       |
| NFR-32 | Giao diện **tiếng Việt** là chính; thiết kế sẵn khả năng thêm ngôn ngữ (ví dụ tiếng Anh) về sau                                                | S       |
| NFR-33 | Mã nguồn có **tài liệu kỹ thuật**, kiểm thử tự động cho các luồng quan trọng (đặt tour, giữ chỗ, tính giá, thanh toán)                         | M       |
| NFR-34 | Bàn giao **toàn bộ mã nguồn, tài liệu, tài khoản quản trị** cho công ty; không bị khoá phụ thuộc không cần thiết vào một nhà cung cấp          | M       |

---

## 7. Tích hợp với hệ thống bên ngoài

| ID     | Hệ thống                                                                         | Mục đích                                                 | Ưu tiên |
| ------ | -------------------------------------------------------------------------------- | -------------------------------------------------------- | ------- |
| INT-01 | Cổng thanh toán (ví dụ VNPay, MoMo, ZaloPay, VietQR/ngân hàng, cổng thẻ quốc tế) | Thanh toán trực tuyến, nhận kết quả giao dịch, hoàn tiền | M       |
| INT-02 | Dịch vụ email giao dịch                                                          | Gửi email xác nhận, nhắc nhở, bản tin                    | M       |
| INT-03 | Dịch vụ SMS / Zalo OA (ZNS)                                                      | Xác nhận, OTP, nhắc lịch                                 | S       |
| INT-04 | Google Analytics, Google Search Console, Tag Manager                             | Đo lường hành vi, SEO, đo chuyển đổi                     | M       |
| INT-05 | Google Maps (hoặc bản đồ tương đương)                                            | Bản đồ hành trình, văn phòng                             | S       |
| INT-06 | Công cụ chat (Zalo, Messenger, live chat)                                        | Tư vấn trực tuyến                                        | S       |
| INT-07 | Facebook Pixel / công cụ quảng cáo                                               | Đo hiệu quả chiến dịch, tiếp thị lại                     | S       |
| INT-08 | CRM / phần mềm kế toán hiện có của công ty (nếu có)                              | Đồng bộ khách hàng, đơn hàng, doanh thu                  | C       |
| INT-09 | Dịch vụ hoá đơn điện tử                                                          | Xuất hoá đơn tự động                                     | C       |

---

## 8. Vai trò và phân quyền (đề xuất)

| Vai trò                         | Quyền chính                                                                                                                   |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| **Quản trị hệ thống**           | Toàn quyền, quản lý tài khoản, quản lý quyền, cấu hình, nhật ký                                                               |
| **Quản lý sản phẩm (tour/giá)** | Tạo/sửa tour, lịch khởi hành, giá, khuyến mãi, quản lý bài viết, banner, FAQ, thư viện ảnh                                    |
| **Nhân viên tư vấn / sales**    | Xem và xử lý đơn, tạo đơn thủ công, xử lý yêu cầu tư vấn và báo giá đoàn, xem đơn, hỗ trợ huỷ/đổi, xử lý phản hồi và đánh giá |
| **Kế toán**                     | Xem và đối soát thanh toán, hoàn tiền, hoá đơn, báo cáo doanh thu                                                             |
| **Ban lãnh đạo**                | Xem báo cáo, không sửa dữ liệu                                                                                                |

---

## 9. Tham chiếu Kỹ thuật & Kiến trúc

Toàn bộ chi tiết về Tech Stack, Phong cách thiết kế kiến trúc (Modular Monolith), Luồng xử lý đồng thời, Sơ đồ triển khai và Cấu trúc dự án được quản lý tập trung tại tài liệu kiến trúc riêng:

👉 **Xem chi tiết:** [architecture.md](file:///C:/nquyleee/coding/my-projects/blue-sky-group-website/docs/architecture.md)

---

## 10. Chỉ số thành công (KPI) đề xuất

| Chỉ số                                         | Mục tiêu (cần thống nhất với ban lãnh đạo) |
| ---------------------------------------------- | ------------------------------------------ |
| Tỉ lệ chuyển đổi (truy cập → đặt tour/tư vấn)  | [x]% sau [n] tháng                         |
| Tỉ lệ hoàn tất đặt tour khi đã bắt đầu         | ≥ [x]%                                     |
| Thời gian hoàn tất đặt tour                    | ≤ [x] phút                                 |
| Tốc độ tải trang trên điện thoại (LCP)         | ≤ 2,5 giây                                 |
| Số yêu cầu báo giá đoàn mỗi tháng              | [x] yêu cầu                                |
| Thời gian phản hồi yêu cầu tư vấn/báo giá      | ≤ [x] giờ làm việc                         |
| Điểm hài lòng khách hàng (đánh giá trung bình) | ≥ [x]/5                                    |

---

## 11. Phân giai đoạn triển khai

### Giai đoạn 0 – MVP Demo

> Chi tiết tại `mvp-requirements.md`.

| Nội dung chính                                                                                                                                                                                                                                                                                                        |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Toàn bộ giao diện như bản final. Deploy đơn giản: Vercel (frontend) + Cloudflare free (backend + DB + R2). Thanh toán giả lập (QR mẫu). Email/SMS giả lập (hiển thị trên giao diện). Xuất Excel/PDF. Bản đồ. Analytics GA4. Chưa cần: tên miền, staging, SSL riêng, CDN riêng, thanh toán thật, SMS/Zalo thật, Redis. |

### Giai đoạn 1 – Lõi bán tour (Production)

| Nội dung chính                                                                                                                                                                                                   |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Nâng cấp từ MVP: tên miền riêng, SSL, hạ tầng production, thanh toán trực tuyến thật (VNPay/MoMo), email/SMS thật, staging, Redis cache, tài khoản khách hàng, bảo mật & kiểm thử tải, sao lưu tự động, giám sát |

### Giai đoạn 2 – Tăng chuyển đổi và chăm sóc

| Nội dung chính                                                                                                                             |
| ------------------------------------------------------------------------------------------------------------------------------------------ |
| Khuyến mãi/mã giảm giá, lưu tiến trình và nhắc đặt tour, chat/Zalo, chuỗi email/SMS chăm sóc, đánh giá khách hàng, báo cáo phễu chuyển đổi |

### Giai đoạn 3 – Mở rộng

| Nội dung chính                                                                                                                                               |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| So sánh tour, theo dõi trạng thái báo giá đoàn, ký/xác nhận hợp đồng trực tuyến, tích hợp CRM/kế toán/hoá đơn điện tử, đa ngôn ngữ, tối ưu tìm kiếm nâng cao |

---

## 12. Tiêu chí nghiệm thu

### 12.1 Nghiệm thu MVP Demo (Giai đoạn 0)

> Chi tiết tại `mvp-requirements.md` mục 10.

1. Toàn bộ trang giao diện hiển thị đúng, responsive mobile → desktop.
2. Luồng **đặt tour → thanh toán QR giả lập → xác nhận** chạy trọn vẹn.
3. Email/SMS giả lập hiển thị đúng nội dung trên giao diện.
4. Xuất **Excel** và **PDF** hoạt động đúng.
5. **Bản đồ** và **Google Analytics** hoạt động.
6. Trang quản trị admin CRUD được tour, đơn hàng, nội dung.
7. Deploy thành công trên Vercel + Cloudflare và truy cập được.

### 12.2 Nghiệm thu Production (Giai đoạn 1+)

1. Tất cả yêu cầu mức **M** của giai đoạn tương ứng hoạt động đúng trên môi trường thử nghiệm và môi trường chạy thật.
2. **Kiểm thử tải** chứng minh hệ thống phục vụ ổn định 100 người dùng đồng thời với thời gian phản hồi theo NFR-03, có báo cáo kèm theo.
3. **Core Web Vitals** đạt ngưỡng "Tốt" trên các trang chính khi đo trên điện thoại.
4. Luồng **đặt tour → giữ chỗ → thanh toán → xác nhận** chạy thành công trọn vẹn với các phương thức thanh toán đã chọn, bao gồm các tình huống lỗi (thanh toán thất bại, hết hạn giữ chỗ, đặt trùng chỗ).
5. **Kiểm thử bảo mật** không còn lỗ hổng mức nghiêm trọng/cao chưa xử lý.
6. **Kiểm thử trải nghiệm người dùng** với nhóm khách đại diện (độ tuổi từ 30 trở lên, đại diện trường học/doanh nghiệp) hoàn thành được các tác vụ chính mà không cần hướng dẫn.
7. Bàn giao đầy đủ **mã nguồn, tài liệu kỹ thuật, tài liệu hướng dẫn sử dụng**, tập huấn nhân viên quản trị.

---

## 13. Giả định, ràng buộc và rủi ro

**Giả định**

- Công ty cung cấp nội dung tour, hình ảnh, video và nhận diện thương hiệu (hoặc có kế hoạch sản xuất nội dung chất lượng cao).
- Công ty đã có giấy phép kinh doanh lữ hành và các thoả thuận cần thiết với cổng thanh toán.
- Quy mô tải ở mức vừa; chưa cần kiến trúc phân tán phức tạp.

**Ràng buộc**

- Ngân sách và thời gian: [bổ sung].
- Tuân thủ pháp luật Việt Nam về thương mại điện tử, bảo vệ dữ liệu cá nhân và kinh doanh lữ hành.

**Rủi ro chính**

| Rủi ro                                                            | Hướng xử lý đề xuất                                                  |
| ----------------------------------------------------------------- | -------------------------------------------------------------------- |
| Đặt trùng chỗ / sai số chỗ khi nhiều người đặt cùng lúc           | Thiết kế xử lý đồng thời ngay từ đầu, kiểm thử riêng                 |
| Lỗi hoặc chậm callback từ cổng thanh toán gây lệch trạng thái đơn | Cơ chế đối soát, thử lại, cảnh báo cho nhân viên                     |
| Hình ảnh/video nặng làm chậm website                              | Chuẩn tối ưu media, CDN, lazy load, kiểm tra Web Vitals thường xuyên |
| Mở rộng phạm vi liên tục (cố làm đủ như website lớn ngay từ đầu)  | Chốt MVP, phân giai đoạn rõ ràng                                     |
| Thiếu nội dung chất lượng khi vận hành                            | Lập kế hoạch nội dung và người phụ trách trước khi ra mắt            |

---

## Phụ lục A – Thuật ngữ

| Thuật ngữ                                   | Giải thích                                                                           |
| ------------------------------------------- | ------------------------------------------------------------------------------------ |
| **Người dùng đồng thời (concurrent users)** | Số người đang sử dụng website cùng một thời điểm                                     |
| **Giữ chỗ (seat hold)**                     | Tạm giữ số chỗ trong thời gian ngắn trong lúc khách thanh toán                       |
| **Guest checkout**                          | Đặt tour không cần tạo tài khoản                                                     |
| **CDN**                                     | Mạng phân phối nội dung, giúp tải ảnh/video nhanh hơn                                |
| **Core Web Vitals**                         | Bộ chỉ số của Google đo tốc độ và độ mượt trải nghiệm web (LCP, INP, CLS)            |
| **SSR / SSG**                               | Kết xuất trang phía máy chủ / tạo trang tĩnh trước, tốt cho SEO và tốc độ            |
| **WCAG**                                    | Chuẩn quốc tế về khả năng truy cập nội dung web                                      |
| **RPO / RTO**                               | Lượng dữ liệu tối đa có thể mất / thời gian tối đa để khôi phục sau sự cố            |
| **OWASP Top 10**                            | Danh sách 10 rủi ro bảo mật ứng dụng web phổ biến nhất                               |
| **Modular Monolith**                        | Ứng dụng triển khai một khối nhưng chia thành các module nghiệp vụ tách biệt rõ ràng |
