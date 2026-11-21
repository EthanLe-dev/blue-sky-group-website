# YÊU CẦU BẢN MVP – WEBSITE BLUE SKY GROUP

| Thông tin      | Nội dung                                                             |
| -------------- | -------------------------------------------------------------------- |
| **Dự án**      | Website bán tour và chăm sóc khách hàng của Blue Sky Group – Bản MVP |
| **Tham chiếu** | `final-requirements.md` (Yêu cầu bản final), `architecture.md` (Kiến trúc hệ thống) |
| **Người lập**  | Lê Ngọc Quý – BA                                                    |
| **Phiên bản**  | 0.1                                                                  |
| **Ngày lập**   | 22/11/2026                                                           |
| **Mục đích**   | Deploy nhanh để khách hàng trải nghiệm giao diện & luồng nghiệp vụ  |

---

## 1. Mục tiêu bản MVP

- **Show được toàn bộ giao diện** như bản final (trang chủ, danh mục, chi tiết tour, đặt tour, thanh toán, quản trị, v.v.).
- Cho khách hàng **trải nghiệm thực tế** các luồng chính: duyệt tour → đặt tour → thanh toán (giả lập) → xác nhận.
- **Deploy đơn giản** trên Vercel + Render Free + Neon Free để demo, chưa cần hạ tầng production.
- Các chức năng nghiệp vụ nặng (thanh toán thật, SMS, tên miền riêng...) sẽ bổ sung ở bản final.

---

## 2. Tham chiếu Kỹ thuật & Kiến trúc

Toàn bộ thông tin về Tech Stack, Mô hình kiến trúc, Cấu trúc module và Sơ đồ triển khai hạ tầng được quy định tập trung tại:

👉 **Xem chi tiết:** [architecture.md](file:///C:/nquyleee/coding/my-projects/blue-sky-group-website/docs/architecture.md)

---

## 3. Phạm vi MVP vs. Final

### 3.1 Những gì MVP **bao gồm** (giao diện đầy đủ + chức năng hoạt động)

| Nhóm                     | Chi tiết                                                                                                             | Tham chiếu bản final               |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| **Trang chủ**            | Hero banner/video, tour nổi bật, điểm đến, ưu đãi, đánh giá, cam kết, cẩm nang, ô tìm kiếm                        | FR-01                               |
| **Danh mục & tìm tour**  | Danh mục theo loại/điểm đến, tìm kiếm, bộ lọc, sắp xếp                                                             | FR-02, FR-03, FR-04                 |
| **Chi tiết tour**        | Thư viện ảnh/video, lịch trình, dịch vụ, chính sách, bản đồ hành trình, FAQ, đánh giá, lịch khởi hành & giá, sticky bar | FR-05, FR-06, FR-07                 |
| **Bản đồ**               | Google Maps / Leaflet cho bản đồ hành trình tour và bản đồ văn phòng                                                | FR-05 (bản đồ), INT-05             |
| **Đặt tour**             | Luồng 3-4 bước, chọn ngày/số người, nhập thông tin, xác nhận; giữ chỗ có thời hạn, chống bán vượt chỗ              | FR-20 → FR-25                       |
| **Thanh toán giả lập**   | Hiển thị mã QR giả lập (VietQR mẫu), nút "Xác nhận đã thanh toán" để demo; **không** kết nối cổng thanh toán thật  | FR-27, FR-28 (giả lập)             |
| **Xác nhận đặt tour**    | Trang xác nhận + email giả lập (ghi log / hiển thị nội dung email trên màn hình thay vì gửi thật)                   | FR-29 (giả lập)                     |
| **Tra cứu đơn**          | Tra cứu bằng mã đơn + SĐT/email                                                                                    | FR-31                               |
| **Khách đoàn**           | Trang "Dành cho Trường học" & "Dành cho Doanh nghiệp", form báo giá, tour mẫu, hồ sơ năng lực, cam kết an toàn      | FR-40 → FR-45                       |
| **Trang tĩnh**           | Giới thiệu, chính sách, FAQ, liên hệ                                                                                | FR-63, FR-66, FR-67                 |
| **Blog / Cẩm nang**      | Bài viết có danh mục, thẻ, bài liên quan                                                                            | FR-70                               |
| **Trang điểm đến**       | Landing page điểm đến + tour liên quan                                                                               | FR-71                               |
| **Thư viện ảnh**         | Ảnh/video phân loại, xem toàn màn hình                                                                              | FR-72, UX-13                        |
| **SEO nền tảng**         | URL thân thiện, meta tags, sitemap, schema markup, OG tags                                                           | FR-73                               |
| **Analytics**            | Google Analytics 4 + Google Search Console tích hợp sẵn                                                              | INT-04                               |
| **Quản trị (Admin)**     | CRUD tour, lịch khởi hành, giá, quản lý đơn, tạo đơn thủ công, quản lý báo giá đoàn, quản lý nội dung, phân quyền  | FR-80 → FR-89                       |
| **Xuất file**            | Xuất báo cáo ra **Excel** (`.xlsx`), xuất lịch trình/báo giá ra **PDF**                                             | FR-08, FR-48, FR-99                 |
| **Báo cáo cơ bản**       | Dashboard doanh thu, số đơn, báo cáo theo tour/thời gian                                                            | FR-95, FR-99                        |
| **Lưu / yêu thích tour** | Lưu tour vào localStorage (không cần đăng nhập)                                                                      | FR-10                               |
| **Tour đã xem**          | Gợi ý tour liên quan, tour đã xem gần đây                                                                           | FR-11                               |
| **Chia sẻ tour**         | Facebook, copy link                                                                                                  | FR-12 (không Zalo share API)        |
| **Responsive**           | Mobile-first, hỗ trợ điện thoại → desktop                                                                           | UX-16, NFR-30                       |
| **UX cho 30+**           | Chữ lớn, nút lớn, tương phản tốt, điều hướng đơn giản                                                               | UX-01 → UX-08                       |
| **Hiệu ứng**            | Skeleton loading, chuyển trang mượt, lazy load ảnh                                                                   | UX-12, UX-15                        |

### 3.2 Những gì MVP **chưa làm** (hoãn sang bản final / production)

| Hạng mục                          | Lý do hoãn                                                     | Tham chiếu bản final       |
| --------------------------------- | -------------------------------------------------------------- | --------------------------- |
| **Tên miền riêng**                | Chưa cần, dùng domain Vercel/Cloudflare mặc định               | —                           |
| **Máy chủ production**            | Dùng free tier Vercel + Render Free + Neon Free để demo                      | NFR-10 → NFR-15             |
| **Môi trường staging**            | Chỉ có 1 môi trường demo                                       | NFR-14                      |
| **SSL riêng**                     | Vercel & Cloudflare tự cấp SSL mặc định                        | NFR-20                      |
| **CDN riêng**                     | Vercel Edge + Cloudflare đã có CDN tích hợp                    | NFR-05                      |
| **DNS riêng**                     | Chưa cần, dùng subdomain mặc định                              | —                           |
| **Thanh toán online thật**        | Thay bằng QR giả lập + nút xác nhận; chưa ký cổng thanh toán  | INT-01                      |
| **SMS thật**                      | Giả lập: hiển thị nội dung SMS trên màn hình / ghi log         | INT-03                      |
| **Zalo OA / ZNS**                 | Chưa tích hợp; chỉ hiển thị link Zalo liên hệ                 | INT-03, INT-06              |
| **Email gửi thật**                | Giả lập: hiển thị nội dung email trên màn hình / ghi log       | INT-02                      |
| **Chat trực tuyến**               | Chưa tích hợp widget chat                                      | FR-62, INT-06               |
| **Tài khoản khách hàng**          | Chưa có đăng ký/đăng nhập, OTP, Google login                   | FR-50 → FR-54               |
| **Mã giảm giá / khuyến mãi**     | Giao diện có nhưng chưa tích hợp logic tính giá                | FR-26                       |
| **Lưu tiến trình & nhắc lại**    | Chưa có email marketing / abandoned cart                        | FR-30                       |
| **Huỷ / đổi tour trực tuyến**    | Chưa có luồng self-service                                     | FR-32                       |
| **Hoá đơn điện tử**               | Chưa tích hợp dịch vụ hoá đơn                                  | FR-33, INT-09               |
| **So sánh tour**                  | Tính năng Could, hoãn                                           | FR-09                       |
| **Upload danh sách đoàn Excel**   | Hoãn                                                            | FR-46                       |
| **Theo dõi trạng thái báo giá**  | Hoãn                                                            | FR-47                       |
| **Ký hợp đồng trực tuyến**       | Hoãn                                                            | FR-48                       |
| **Facebook Pixel**                | Hoãn, cần tài khoản quảng cáo                                  | INT-07                      |
| **Tích hợp CRM / kế toán**       | Hoãn                                                            | INT-08                      |
| **Đa ngôn ngữ**                   | Hoãn                                                            | NFR-32                      |
| **Redis cache**                   | Hoãn, bổ sung khi lên production                                | Mục 9 bản final             |
| **Kiểm thử tải (load test)**     | Chưa cần ở bản demo                                             | NFR-07                      |
| **Kiểm thử bảo mật chuyên sâu** | Chưa cần ở bản demo                                             | NFR-21                      |
| **Sao lưu tự động**               | Chưa cần ở bản demo                                             | NFR-11                      |
| **Giám sát & cảnh báo**           | Chưa cần ở bản demo                                             | NFR-13                      |
| **Đăng ký nhận bản tin**          | Giao diện có, chưa gửi email thật                               | FR-74                       |
| **Mẫu email/SMS chỉnh sửa**     | Hoãn                                                            | FR-90                       |
| **2FA cho admin**                 | Hoãn                                                            | NFR-23                      |

---

## 4. Chi tiết giả lập trong MVP

### 4.1 Thanh toán giả lập

- Bước thanh toán hiển thị **mã QR giả lập** (hình ảnh QR mẫu kiểu VietQR với thông tin chuyển khoản giả).
- Có 2 lựa chọn:
  - **"Thanh toán toàn bộ"**: hiển thị QR + tổng tiền.
  - **"Đặt cọc"**: hiển thị QR + số tiền cọc.
- Nút **"Tôi đã thanh toán"** → hệ thống tự động chuyển đơn sang trạng thái "Đã thanh toán" (demo).
- Không kết nối cổng thanh toán thật (VNPay, MoMo, ZaloPay...).

### 4.2 Email giả lập

- Khi hệ thống cần gửi email (xác nhận đặt tour, nhắc thanh toán, v.v.):
  - **Ghi log** nội dung email trong database / console.
  - **Hiển thị popup / toast** trên giao diện: *"📧 Email xác nhận đã được gửi đến [email]"* kèm nút **"Xem nội dung email"** để xem preview.
  - Admin có thể xem tất cả "email đã gửi" trong mục **Nhật ký thông báo** trên trang quản trị.

### 4.3 SMS giả lập

- Tương tự email: ghi log + hiển thị nội dung SMS giả lập trên giao diện.
- Thông báo: *"📱 SMS đã được gửi đến [SĐT]"* + nút **"Xem nội dung SMS"**.

---

## 5. Chức năng xuất file

| Loại file  | Chức năng                                             | Thư viện gợi ý (backend Go)          |
| ---------- | ----------------------------------------------------- | ------------------------------------- |
| **Excel**  | Xuất báo cáo doanh thu, danh sách đơn, danh sách tour | `excelize` (Go)                       |
| **PDF**    | Xuất lịch trình tour, báo giá đoàn, hoá đơn demo      | `wkhtmltopdf`, `gofpdf`, hoặc `chromedp` |

- Nút **"Tải Excel"** ở trang báo cáo admin + trang danh sách đơn.
- Nút **"Tải PDF"** ở trang chi tiết tour (khách) + trang báo giá (admin).
- File được tạo on-demand, lưu tạm trên R2 hoặc trả stream trực tiếp.

---

## 6. Bản đồ

- Tích hợp **Google Maps JavaScript API** (hoặc **Leaflet + OpenStreetMap** nếu muốn free hoàn toàn).
- Hiển thị tại:
  - **Trang chi tiết tour**: bản đồ hành trình với các điểm dừng (markers + route).
  - **Trang liên hệ**: vị trí văn phòng công ty.
  - **Trang điểm đến**: vị trí điểm đến trên bản đồ.

---

## 7. Analytics

- Tích hợp **Google Analytics 4 (GA4)** vào Next.js bằng `@next/third-parties` hoặc Google Tag Manager.
- Theo dõi:
  - Lượt xem trang, nguồn truy cập.
  - Sự kiện tuỳ chỉnh: xem tour, bắt đầu đặt, hoàn tất đặt, gửi form báo giá.
  - Phễu chuyển đổi cơ bản.
- Tích hợp **Google Search Console** (xác minh qua meta tag hoặc DNS).

---

## 8. Tiêu chí nghiệm thu MVP

1. **Toàn bộ trang giao diện** đã liệt kê ở mục 3.1 hiển thị đúng, responsive mobile → desktop.
2. **Luồng đặt tour** chạy trọn vẹn: chọn tour → chọn ngày/số người → nhập thông tin → thanh toán QR giả lập → xác nhận → tra cứu đơn.
3. **Thanh toán giả lập** hiển thị QR, nhấn xác nhận → đơn chuyển trạng thái thành công.
4. **Email & SMS giả lập** hiển thị đúng nội dung trên giao diện, ghi log trong hệ thống.
5. **Xuất Excel** từ trang báo cáo admin: file `.xlsx` tải về đúng dữ liệu.
6. **Xuất PDF** lịch trình tour & báo giá đoàn: file `.pdf` hiển thị đúng.
7. **Bản đồ** hiển thị trên trang chi tiết tour và trang liên hệ.
8. **Google Analytics** ghi nhận được sự kiện xem trang và sự kiện tuỳ chỉnh.
9. **Trang quản trị admin** CRUD được tour, đơn hàng, nội dung; phân quyền cơ bản.
10. **Chống spam** Cloudflare Turnstile hoạt động trên các form.
11. **Deploy thành công** trên Vercel (frontend) + Render Free (backend) + Neon Free (database) và truy cập được qua URL demo.

---

## 9. Lộ trình từ MVP → Production

| Bước | Hạng mục                                           | Mô tả                                                            |
| ---- | --------------------------------------------------- | ----------------------------------------------------------------- |
| 1    | Khách hàng duyệt MVP                                | Demo, thu thập phản hồi, điều chỉnh giao diện & luồng            |
| 2    | Mua tên miền + setup DNS                             | Trỏ domain chính thức                                             |
| 3    | Nâng cấp hạ tầng                                    | Chuyển sang Cloudflare Workers Paid / VPS, thêm staging           |
| 4    | Tích hợp thanh toán thật                             | Ký hợp đồng VNPay/MoMo, thay module thanh toán giả lập           |
| 5    | Tích hợp email & SMS thật                            | Kết nối SendGrid/SES (email) + eSMS/Zalo ZNS (SMS)               |
| 6    | SSL riêng + CDN                                      | Cloudflare Pro hoặc CDN riêng cho media                           |
| 7    | Tài khoản khách hàng                                 | Đăng ký/đăng nhập, OTP, Google login                              |
| 8    | Redis cache                                          | Thêm cache cho tour, danh mục, trang chi tiết                    |
| 9    | Bảo mật & kiểm thử                                  | Load test, penetration test, 2FA admin                            |
| 10   | Go live                                              | Chuyển DNS, giám sát, sao lưu                                     |

---

## Phụ lục – Mapping ưu tiên MVP

> Bảng dưới đây map lại mức ưu tiên của từng yêu cầu trong bản final, đánh dấu những yêu cầu nào **vào MVP** và những yêu cầu nào **hoãn**.

| ID gốc          | Tên ngắn                      | Ưu tiên final | MVP? | Ghi chú MVP                     |
| ---------------- | ------------------------------ | ------------- | ---- | -------------------------------- |
| FR-01            | Trang chủ                     | M             | ✅   |                                  |
| FR-02            | Danh mục tour                 | M             | ✅   |                                  |
| FR-03            | Tìm kiếm                     | M             | ✅   |                                  |
| FR-04            | Bộ lọc & sắp xếp             | M             | ✅   |                                  |
| FR-05            | Chi tiết tour                 | M             | ✅   | Có bản đồ                       |
| FR-06            | Lịch khởi hành & giá         | M             | ✅   |                                  |
| FR-07            | Sticky bar đặt tour           | M             | ✅   |                                  |
| FR-08            | Tải lịch trình PDF            | S             | ✅   | Xuất PDF                        |
| FR-09            | So sánh tour                  | C             | ❌   |                                  |
| FR-10            | Yêu thích / lưu tour         | S             | ✅   | localStorage                    |
| FR-11            | Tour đã xem gần đây          | S             | ✅   |                                  |
| FR-12            | Chia sẻ tour                  | S             | ✅   | Không Zalo share API            |
| FR-20            | Guest checkout                | M             | ✅   |                                  |
| FR-21            | Luồng đặt tour 3-4 bước      | M             | ✅   |                                  |
| FR-22            | Giá tự động theo đối tượng   | M             | ✅   |                                  |
| FR-23            | Thông tin hành khách          | M             | ✅   |                                  |
| FR-24            | Giữ chỗ có thời hạn          | M             | ✅   |                                  |
| FR-25            | Chống bán vượt chỗ           | M             | ✅   |                                  |
| FR-26            | Mã giảm giá                  | S             | ⚠️   | UI có, logic tính giá hoãn      |
| FR-27            | Thanh toán online              | M             | ✅   | **QR giả lập**                  |
| FR-28            | Toàn bộ / đặt cọc            | M             | ✅   | Giả lập                        |
| FR-29            | Email/SMS xác nhận            | M             | ✅   | **Giả lập**                    |
| FR-30            | Lưu tiến trình                | S             | ❌   |                                  |
| FR-31            | Tra cứu đơn                  | M             | ✅   |                                  |
| FR-32            | Huỷ/đổi tour online           | S             | ❌   |                                  |
| FR-33            | Hoá đơn / chứng từ           | S             | ❌   |                                  |
| FR-40 → FR-44   | Khách đoàn (trường/DN)       | M             | ✅   |                                  |
| FR-45            | Đánh giá khách đoàn          | S             | ✅   | Mock data                       |
| FR-46            | Upload danh sách đoàn        | S             | ❌   |                                  |
| FR-47            | Theo dõi trạng thái báo giá  | C             | ❌   |                                  |
| FR-48            | Tải báo giá PDF               | C             | ✅   | Xuất PDF cơ bản                 |
| FR-50 → FR-54   | Tài khoản khách hàng         | S/C           | ❌   |                                  |
| FR-60            | Hotline, Zalo, email          | M             | ✅   | Hiển thị info, chưa tích hợp OA |
| FR-61            | Yêu cầu gọi lại              | M             | ✅   |                                  |
| FR-62            | Chat trực tuyến               | S             | ❌   |                                  |
| FR-63            | FAQ                           | M             | ✅   |                                  |
| FR-64            | Cẩm nang chuẩn bị            | S             | ✅   | Nội dung tĩnh                   |
| FR-65            | Chuỗi email chăm sóc         | S             | ❌   |                                  |
| FR-66            | Trang giới thiệu              | M             | ✅   |                                  |
| FR-67            | Trang chính sách              | M             | ✅   |                                  |
| FR-70            | Blog / cẩm nang               | M             | ✅   |                                  |
| FR-71            | Trang điểm đến                | S             | ✅   |                                  |
| FR-72            | Thư viện ảnh/video            | S             | ✅   |                                  |
| FR-73            | SEO kỹ thuật                  | M             | ✅   |                                  |
| FR-74            | Đăng ký bản tin               | S             | ⚠️   | UI có, chưa gửi email thật      |
| FR-80 → FR-89   | Quản trị back office          | M/S           | ✅   |                                  |
| FR-90            | Mẫu email/SMS chỉnh sửa     | S             | ❌   |                                  |
| FR-95            | Báo cáo doanh thu             | M             | ✅   |                                  |
| FR-96            | Báo cáo lấp đầy chỗ         | S             | ✅   | Cơ bản                         |
| FR-97            | Phễu chuyển đổi              | S             | ❌   |                                  |
| FR-98            | Báo cáo báo giá đoàn        | S             | ✅   | Cơ bản                         |
| FR-99            | Xuất Excel/CSV                | M             | ✅   |                                  |
| INT-01           | Cổng thanh toán               | M             | ✅   | **Giả lập**                    |
| INT-02           | Email giao dịch               | M             | ✅   | **Giả lập**                    |
| INT-03           | SMS / Zalo OA                 | S             | ✅   | **Giả lập**                    |
| INT-04           | Google Analytics              | M             | ✅   |                                  |
| INT-05           | Google Maps                   | S             | ✅   |                                  |
| INT-06           | Chat (Zalo, Messenger)        | S             | ❌   |                                  |
| INT-07           | Facebook Pixel                | S             | ❌   |                                  |
| INT-08           | CRM / kế toán                 | C             | ❌   |                                  |
| INT-09           | Hoá đơn điện tử              | C             | ❌   |                                  |

**Chú thích:** ✅ = Vào MVP, ❌ = Hoãn, ⚠️ = Giao diện có nhưng chức năng chưa đầy đủ.

## Quyết định bổ sung ngày 22/11/2026 — Admin và poster

Người dùng cho phép giả định giá/đặt chỗ và thông tin công ty để triển khai demo, với điều kiện **đều sửa được ở Admin**. Quyết định này thay thế các mục đang chờ BA về giá trị ban đầu trong tài liệu; chưa phải duyệt chính sách bán thật. Không có deadline bắt buộc, vẫn thực hiện theo dependency và quality gates.

- Mặc định demo: giữ chỗ20 phút, cọc30%, trả dư trước ngày đi7 ngày, adult≥12/child2–<12/infant<2; adult/child1 chỗ, infant0; tối đa20 khách/đơn. Policy huỷ demo: nhân viên xét duyệt theo chính sách hiển thị, không tự hoàn tiền. SLA demo: phản hồi trong1 ngày làm việc. Tất cả hiển thị nhãn demo khi chưa được thay bằng dữ liệu thật.
- Admin quản lý mặc định đặt chỗ, giá/tuổi/chỗ/cọc/hạn trả theo lịch khởi hành, policy huỷ theo tour, thông tin doanh nghiệp/liên hệ/giấy phép/tài liệu/SLA. Thông tin chưa có dùng placeholder rõ ràng, không bịa giấy phép hay số điện thoại thật; link gọi/Zalo chưa cấu hình phải disabled.
- Thay đổi cấu hình có validation, version, audit; chỉ tác động đơn mới. Quote đang hiệu lực dùng snapshot cấu hình đã chốt; booking cũ giữ nguyên giá/policy/hạn, không cập nhật ngược. Sửa giá/tuổi/chỗ/cọc của departure tăng price_version và yêu cầu báo giá lại.
- Tour do team thiết kế poster. Admin upload/thay poster và nhập dữ liệu tour có cấu trúc; poster không thay trường giá/lịch/itinerary dùng tìm kiếm, SEO và checkout. Chưa nhận được ảnh mẫu trong chat hoặc workspace (chỉ có logo), cần bổ sung để chốt bố cục/tỉ lệ poster.
- Chọn Render Free cho Go API và Neon Free cho PostgreSQL để demo nhỏ, trong quota miễn phí; không tự nâng gói trả tiền. Cần smoke test thực tế, nhất là cold start và PDF, trước khi xác nhận đủ năng lực. Cloudflare vẫn cho Turnstile/R2, không làm Go host/DB. Miễn phí backend/DB không đồng nghĩa đã xác nhận chi phí toàn hệ thống.
