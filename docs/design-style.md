# Thiết kế UI/UX Blue Sky Group

Loại hình: **service / site-kind other — dịch vụ du lịch**, khách lẻ/gia đình từ 30 tuổi, trường học và doanh nghiệp. Hành động chính: tìm lịch khởi hành và đặt tour; hành động hỗ trợ: yêu cầu tư vấn/báo giá đoàn.

Phiên bản 0.1, 22/11/2026. Đây là đặc tả thiết kế, chưa phải UI đã dựng/duyệt. Nguồn: [MVP](mvp-requirements.md), [Final](final-requirements.md), [API](api-spec.md). Áp dụng **taste-skill:design-taste-frontend 2.5.1** (tasteskill); tìm skill trong catalog phiên làm việc, không hardcode đường dẫn máy cá nhân vào source app.

## 1. Art direction

Ấm áp, thoáng, đáng tin và giàu hình ảnh hành trình. Bố cục có điểm nhấn editorial bằng ảnh điểm đến/đoàn thực, typography rõ và khoảng trắng đều. Một hệ thống màu/type/spacing xuyên suốt public, checkout và admin; admin giảm yếu tố trang trí để đọc dữ liệu nhanh.

- Hero nói rõ bán tour gì/cho ai, một ảnh chủ đạo có nội dung liên quan và ô tìm tour dễ thấy. CTA “Xem lịch khởi hành”, “Đặt tour”, “Nhận tư vấn”; không dùng slogan trừu tượng làm tiêu đề duy nhất.
- Một ý chính mỗi section marketing, lề trái/phải cân bằng, nhịp xen ảnh rộng và danh sách tour. Không lặp một mẫu ba card cho mọi nội dung.
- Ảnh chụp thật, nền màu phẳng và một dải màu thương hiệu là chính; không shader/WebGL, hạt bay, glassmorphism khắp trang, gradient tím hoặc 3D trang trí cho du lịch.
- Card có ý nghĩa: ảnh/tiêu đề/lịch/giá/CTA. Không biến mọi đoạn văn thành hộp bo góc. Giá và chỗ trống lấy dữ liệu thật; seed/review mẫu có nhãn demo.

## 2. Áp dụng taste-skill và xử lý xung đột

Khi bắt đầu dựng trang, đọc đầy đủ SKILL.md và các reference `copy.md`, `copy-voice.md`, `system.md`, `layout-v4.md`, `sections.md`; đọc thêm `motion.md`, `photos.md`, `moments.md`, `perf.md` theo asset dùng. Khởi tạo gate, chạy dealer cho layout/moment phù hợp service, lưu lựa chọn và lý do trong hồ sơ thiết kế. Khi có màu logo được xác minh, dùng brand path thay palette ngẫu nhiên. Tài liệu này chưa chạy dealer/gate vì chưa có trang.

| Quy tắc skill cần điều chỉnh | Quyết định của dự án và căn cứ |
| --- | --- |
| Next.js static export | Giữ SSR/SSG theo architecture, API proxy và SEO; không bật `output: export` chỉ để chạy checker. Kiểm trên URL local của bản build chạy được. |
| Mỗi section đúng 1–2 viewport | Chỉ là định hướng nhịp trang marketing. Tour list/detail, checkout, form/admin theo nội dung; không tạo khoảng trống, cắt lịch trình hoặc đẩy CTA khỏi màn hình (UX-04/05/27). |
| Không caption dưới visual | Tour name/giá/lịch, chú thích điểm đến cần thiết vẫn có để khách hiểu sản phẩm; credit gom footer nếu phù hợp giấy phép. Không xoá dữ liệu mua tour để qua checker. |
| Không số điện thoại/địa chỉ/giấy phép thật | FR-43/60/66/67 cần thông tin doanh nghiệp đã duyệt. Hiển thị đầy đủ thông tin thật được phép công bố; demo thiếu dữ liệu phải ghi rõ, không bịa. |
| Không sample label | Giữ nhãn dữ liệu/thanh toán/email/SMS giả lập theo yêu cầu MVP và AGENTS. |
| Mỗi trang phải có moment | Marketing có 1 moment ảnh/type, tối đa 2 cho trang dài. Checkout/admin ưu tiên phản hồi trạng thái, không ép animation lớn làm gián đoạn thao tác. |
| Không carousel | Tránh carousel tự chạy; gallery tour vẫn cho vuốt/prev-next/lightbox do UX-13 yêu cầu. |

Ghi ngoại lệ áp dụng theo route/check trong báo cáo và `.taste-skill/waivers.json` theo reference gate, không sửa code checker hoặc ẩn nội dung để đạt PASS. Không coi waiver là đã đạt yêu cầu accessibility/performance.

## 3. Màu, typography và tokens

Palette **đề xuất ban đầu**, cần đối chiếu logo `blue-sky-group-logo.jpg` khi dựng UI; chưa khẳng định đây là mã màu gốc logo.

| Token | Giá trị | Dùng cho |
| --- | --- | --- |
| background | `#F7FAFC` | Nền trang sáng |
| surface | `#FFFFFF` | Form/card/menu |
| ink | `#142B3A` | Tiêu đề, nội dung |
| muted | `#506575` | Metadata vẫn cần đủ tương phản |
| brand | `#075985` | CTA/link/focus |
| brand-hover | `#0C4A6E` | Hover CTA |
| brand-soft | `#E8F4FA` | Nền chọn/filter, không chữ nhỏ trên trắng |
| border | `#D4E0E7` | Phân tách; viền input cần đo và tăng tối nếu chưa đạt 3:1 |
| success / danger | `#166534` / `#B91C1C` | Trạng thái kèm text/icon |

Đo contrast trên giao diện thật: chữ thường ≥4.5:1, chữ lớn ≥3:1, control/focus ≥3:1. Màu không phải tín hiệu duy nhất. Text trên ảnh dùng vùng nền đặc/scrim trung tính đủ tương phản, tránh chữ đè lên chi tiết ảnh.

Font đề xuất: **Be Vietnam Pro** cho heading/body; fallback `Noto Sans`, `system-ui`, `sans-serif`. Khi tích hợp phải lấy bản font có glyph/subset Vietnamese và giấy phép, self-host WOFF2 hoặc dùng cơ chế Next font tương đương; kiểm tra thực tế thay vì tin tên font. Không dùng font display thiếu dấu, chữ body quá mảnh hoặc quá nhiều family. Chỉ tải weight 400/500/600/700 cần dùng, `font-display: swap`; preload font quan trọng có chủ đích, kiểm CLS khi đổi font.

- Body 17–18px mobile, 16–18px desktop, line-height 1.6; label tối thiểu16px, metadata tối thiểu14px không dùng cho giá/điều khoản chính.
- H1 mobile 32–40px, desktop 48–64px bằng clamp; H2 26–40px, H3 20–24px. Line-height heading 1.2–1.3 để không cắt dấu. Không letter-spacing âm mạnh, không viết HOA cả đoạn tiếng Việt.
- Test chuỗi: “Đặt tour Đà Nẵng – Hội An”, “Nguyễn Thị Ánh”, “Ắ Ằ Ẳ Ẵ Ặ Ấ Ầ Ẩ Ẫ Ậ Ế Ề Ể Ễ Ệ Ớ Ờ Ở Ỡ Ợ Ứ Ừ Ử Ữ Ự”, “1.250.000 ₫”; kiểm cả mobile, bold, input và PDF (PDF embed font riêng).
- Spacing scale 4/8/12/16/24/32/48/64/96px; wrap tối đa1200px; gutter mobile20px (16px ở320), tablet32px, desktop48–64px. Grid desktop12 cột, tablet6, mobile1 luồng chính.
- Radius button/input8px, card12px, dialog16px; shadow nhẹ chỉ cho phần nổi/menu/sticky. Button tối thiểu44×44px, primary thường48px; khoảng cách target ≥8px. Chỉ một primary CTA trong một nhóm hành động.

## 4. Luồng và cấu trúc màn hình

| Trang | Nội dung và tương tác cần có |
| --- | --- |
| Home | Hero ảnh/search; tour nổi bật; điểm đến; trường/doanh nghiệp; bằng chứng uy tín có nguồn; cẩm nang; CTA tư vấn. Hero không auto-play âm thanh. |
| Tour list | Search/filter dễ sửa, số kết quả, sort; filter mobile mở drawer có Apply/Reset; giữ URL và Back. Card có giá từ/ngày/thời lượng, hết chỗ và loading rõ. |
| Tour detail | Gallery, tóm tắt/giá/lịch, sticky CTA; itinerary theo ngày, gồm/không gồm, policy, map, FAQ, reviews. Link anchor có scroll margin tránh header che. |
| Checkout | Chọn ngày/số người → liên hệ/hành khách → xác nhận/thanh toán. Desktop summary cạnh form, mobile tổng tiền và nút tiếp tục rõ. Cọc/dư nợ/hạn/expiry hiển thị riêng; step back không mất dữ liệu. |
| Payment | Nhãn giả lập nổi bật, QR không nhận tiền thật; trạng thái đang xử lý/thành công/hết hạn/lỗi. Đặt cọc thành công vẫn hiện số còn thiếu. |
| Lookup | Mã+email/SĐT, consent/captcha phù hợp; chỉ tóm tắt đã che thông tin. Không hiện form lỗi khác nhau để lộ khách có tồn tại. |
| Group/contact | Đề xuất dành riêng trường/doanh nghiệp, hồ sơ/cam kết có nguồn, form tối giản và SLA được duyệt. Map có địa chỉ text dự phòng. |
| Admin | Sidebar rõ, page title/actions, filter/table/pagination; mobile dùng summary row hoặc vùng bảng scroll có nhãn. Confirm huỷ/refund nêu tác động, lỗi stale version cho reload, không optimistic update tiền/chỗ. |

Mỗi component có default/hover/focus/active/disabled/loading/error/success và empty nếu liên quan. Loading giữ kích thước để tránh nhảy; disabled có lý do. Error đặt cạnh field và summary có focus, giữ input hợp lệ. Toast không thay thông báo lỗi quan trọng; success message dùng aria-live polite. Modal/menu quản lý focus, Escape, restore focus; mọi chức năng dùng được bằng keyboard. Sticky booking/hotline/chat không đè nhau, có safe-area padding và khoảng trống dưới nội dung.

## 5. Motion có mục đích

| Tình huống | Thời gian/biên độ đề xuất | Quy tắc |
| --- | --- | --- |
| Button hover/press | 120–180ms; translateY -1px / scale .98 | Chỉ hover trên thiết bị có hover; focus rõ độc lập animation |
| Linked card/photo | 180–240ms; lift2px, ảnh scale1.02 | Clip trong frame; không tilt, không dịch layout |
| Menu/dialog | 180–240ms; opacity + translateY8px | Focus và thao tác không phải chờ kết thúc motion |
| Hero entrance | 450–650ms, translateY12px; stagger≤60ms | Một lần/load, không giấu LCP image; tổng≤900ms, không split từng ký tự tiếng Việt |
| Marketing moment | 500–800ms; ảnh reveal hoặc type composition một lần | 1–2 điểm nhấn có lý do; không fade mọi card khi scroll |
| Loading/status | Spinner nhỏ hoặc skeleton tĩnh/pulse nhẹ | Không loop trang trí vô hạn; không dùng animation làm timer nghiệp vụ |

Easing mặc định `cubic-bezier(.22,1,.36,1)`, exit `cubic-bezier(.4,0,1,1)`. Ưu tiên CSS transform/opacity; không animate width/height/top/left/blur lớn. Không scroll hijack, parallax chữ, auto-carousel, hiệu ứng bật nhấp nháy hoặc sound. Dùng thư viện animation chỉ khi interaction cần và không có giải pháp sẵn; không kéo cả motion kit vào checkout.

`prefers-reduced-motion: reduce`: bỏ dịch chuyển/scale/reveal/loop, nội dung luôn hiện, đổi trạng thái tức thì hoặc fade≤100ms. JavaScript tắt vẫn đọc được nội dung SSR. Nếu media chuyển động kéo dài >5s phải có pause và poster, dừng ngoài viewport/tab ẩn. Mục tiêu animation mượt gần60fps trên thiết bị kiểm thử; xem Performance trace khi có giật, không tuyên bố 60fps chỉ bằng mắt.

## 6. Media, copy và nghiệm thu

Ảnh responsive có dimensions/aspect-ratio; hero ưu tiên tải, ảnh dưới fold lazy load; WebP/AVIF khi phù hợp. Không upscale ảnh nhỏ. Nếu chưa có ảnh công ty, dùng ảnh được cấp quyền phù hợp điểm đến, ghi nguồn và stand-in trong hồ sơ; không gắn ảnh stock thành khách/đoàn thật. Alt mô tả nội dung hữu ích, ảnh trang trí alt rỗng.

Giọng văn rõ, lịch sự, thân thiện: “Chọn ngày khởi hành”, “Còn 3 chỗ” khi DB thật, “Giữ chỗ đến 14:20”. Không khẩn cấp giả, khen thương hiệu bằng số liệu bịa. Lỗi có hướng xử lý/hotline; tính năng hoãn disabled với giải thích.

Trước DONE UI: build rồi chạy checker/consistency/layout/copy/perf/gate theo skill trên trang local thực; lưu dòng GATE thật, waiver, FAIL còn mở và screenshot. Không sửa kit để qua check. Kiểm mắt và tương tác ở1440,1024,820,768,390 và stress320; keyboard/zoom/reduced motion theo [testing-strategy](testing-strategy.md). Gate chưa chạy hoặc ERROR không đồng nghĩa PASS. Hiện tài liệu này chưa có browser evidence vì chưa có ứng dụng.
