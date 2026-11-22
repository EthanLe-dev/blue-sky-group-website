# Page map và wireframe — T004

Bản khung thông tin, không phải UI nghiệm thu. Dùng dữ liệu demo inventory, hướng taste/GSAP theo design-style; triển khai visual ở M2.

| Route | Luồng / wireframe từ trên xuống |
| --- | --- |
| / | Header → hero ảnh/search/CTA → tour nổi bật → điểm đến → trường/doanh nghiệp → trust có nguồn → cẩm nang → footer |
| /tours | Breadcrumb/search → filter/sort/result count → tour grid → pagination |
| /tours/[slug] | Tên/thời lượng → gallery/poster viewer → giá/lịch/CTA → itinerary → gồm/không gồm → policy → map/FAQ/review |
| /booking | Stepper → ngày/số người → liên hệ/hành khách → review/consent → payment mock → confirmation |
| /booking/lookup | Mã+contact+captcha → summary che PII, không passenger |
| /schools,/companies | Nhu cầu → tour mẫu → hồ sơ/cam kết → form báo giá/SLA |
| /blog,/destinations,/faq,/about,/contact,/policies | Nội dung CMS có heading/SEO/links, map lazy nếu có |
| Admin /login | Email/password/captcha → errors → login, reset flow |
| Admin / | Navigation → heading/actions → filter/table/detail theo quyền |
| Admin /settings | Nhóm đặt tour/công ty/SLA → form đơn vị rõ → preview → save/version conflict |

Desktop detail/checkout chia nội dung chính + summary sticky. Mobile một cột: tên/giá/lịch lên trước, filter thành drawer, CTA dưới có safe area, nội dung chừa khoảng để không bị che. 320px vẫn không tràn; text tiếng Việt16–18px, touch44px. Error/loading/empty/sold-out/expired/reduced-motion có state riêng. Không bắt khách đọc poster để lấy giá và itinerary.

Review cấu trúc: có đường từ khám phá đến đặt, từ khách đoàn đến lead, hỗ trợ hotline khi có cấu hình; max3–4 bước checkout. Keyboard/focus và thực nghiệm người dùng cần kiểm khi có UI (T703/T706).
