# Kiểm tra UI M2/M3 — 22/11/2026

`GATE: OPEN 19 (rounds used up)` — đã chạy đủ 3 vòng taste-skill, dừng chỉnh thẩm mỹ theo giới hạn skill. Chức năng chạy được không đồng nghĩa đã nghiệm thu T703.

Thiết kế và nguồn ảnh: [.taste-skill/plan.md](../.taste-skill/plan.md), [media-credits](media-credits.md). Logo công ty gốc, hero photo Vĩnh Hy, brand navy/blue/yellow, Be Vietnam Pro tự host; nav-centre/footer-dark/close-type/surface-pairs/solid, picture-develops film. Voice trực tiếp tiếng Việt, nội dung demo có nhãn bắt buộc.

## Finding còn mở

- **effect-kind** — (the page): a service page (--type) moves with the scroll: scroll libraries or timelines (ScrollTrigger): effects fit tech, gaming and software pages and a product with a technical subject, never a place, shop or service page (house rule)
- **hero-enter** — main#main > section.hero.hero--photo: the hero does not enter on load: no head line setting html.m-enter (motion.md, The hero's entrance)
- **section-height** — section#journeys "ĐI GẦN, KHÁM PHÁ NHIỀUTour Ninh Chữ & Vĩ": the section is 804px tall (0.9 screens), 96px short of one screen (900px): every section after the hero is one screen or two, 2px either way (SR-2)
- **section-height** — main#main > section.section.coast-section "Ninh Chữ, từ bờ cát đến làng nghề.Xem lị": the scene's section is 804px tall (0.9 screens), 96px short of one screen (900px): every section after the hero is one screen or two, 2px either way; a pinned scene is (hold + 1) screens tall, so write data-big-scene-opts='{"hold": 1}' (two screens) and keep --h on the section (SR-2)
- **section-height** — footer.closing "Chọn hành trình tiếp theo của bạn.Tìm to": the section starts at 3312px, 288px above the screen grid (4 x 900px): sections start on whole screens (SR-2)
- **one-section** — footer.closing > div.wrap > div.close-type > h2 "Chọn hành trình tiếp theo của bạn.": the screen at scroll 2700px (stride 3) shows 288px of section 5 and the heading of section 5 whole and the content of 2 sections, over 24px each beside section 4: a screen shows one section (another may peek 8px, 24px WARN) (SR-4)
- **one-section** — section.section.programmes > div.wrap > div.section-head > h2 "Tour trường học & doanh nghiệp": the screen at scroll 1800px (stride 2) shows 192px of section 4 and the heading of section 4 whole and the content of 2 sections, over 24px each beside section 3: a screen shows one section (another may peek 8px, 24px WARN) (SR-4)
- **one-section** — section#journeys "ĐI GẦN, KHÁM PHÁ NHIỀUTour Ninh Chữ & Vĩ": the screen at scroll 96px (the top of section 1) shows 96px of section 2 beside section 1: a screen shows one section (another may peek 8px, 24px WARN) (and 3 more like it: scroll 1704px, scroll 2508px, scroll 900px) (SR-4)
- **container-edges** — header.site-header.nav "Bỏ qua đến nội dungBLUE SKY GROUPKhám ph": the nav's first box starts at 16px and its last ends at 1340px, not on its .wrap at 100-1340px (SR-6)
- **container-edges** — main#main > section.hero.hero--photo > div.wrap "BLUE SKY GROUP · DU LỊCH & TRẢI NGHIỆMTo": the hero's .wrap holds 1 children, 0 of them .hero-text: one .hero-text and at most one strip (SR-6)
- **container-edges** — footer.closing > div.wrap "Chọn hành trình tiếp theo của bạn.Tìm to": footer.closing > .wrap holds 3 children: exactly two (the call, then the legal row) (SR-6)
- **container-edges** — header.site-header.nav > a.skip-link "Bỏ qua đến nội dung": a box runs 16-198px, outside the container lines of the nav at 100-1340px: full width is for background layers and grounds only (SR-6)
- **container-edges** — div.wrap > div.group-programmes.grid > div.service-rows > a:nth-of-type(1) "01 / Trường họcNgoại khoá học đườngTìm h": a box over half the container wide runs 627-1340px, not on the container's lines at 100-1340px (SR-6)
- **container-edges** — div.wrap > div.group-programmes.grid > div.service-rows > a:nth-of-type(2) "02 / Doanh nghiệpDu lịch gắn kết đồng độ": a box over half the container wide runs 627-1340px, not on the container's lines at 100-1340px (SR-6)
- **screen-short** — section.hero.hero--photo section.hero.hero--photo "BLUE SKY GROUP · DU LỊCH & TRẢI ": 804px tall; one screen under the nav is 900px
- **screen-short** — section.section.coast-section section.section.coast-section "Ninh Chữ, từ bờ cát đến làng ngh": 804px tall; one screen under the nav is 900px
- **screen-height** — section.section.coast-section section.section.coast-section "Ninh Chữ, từ bờ cát đến làng ngh": 804px, 0.89 screens; a section is exactly one or two (--h)
- **neighbour** — section.section.coast-section section.section.coast-section "Ninh Chữ, từ bờ cát đến làng ngh": framed, 26px of section.section.programmes section.section.programmes "Tour trường học & doanh nghiệp01" shows
- **screen-short** — section.section.programmes section.section.programmes "Tour trường học & doanh nghiệp01": 804px tall; one screen under the nav is 900px
- **screen-height** — section.section.programmes section.section.programmes "Tour trường học & doanh nghiệp01": 804px, 0.89 screens; a section is exactly one or two (--h)
- **neighbour** — section.section.programmes section.section.programmes "Tour trường học & doanh nghiệp01": framed, 96px of the footer shows

## Ngoại lệ đã ghi nhận

- hero-enter: Yêu cầu repo dùng GSAP: hero dùng useGSAP có matchMedia/cleanup và SSR nội dung hiện sẵn; không thêm head script m-enter để giấu toàn hero trước hydration.
- effect-kind: GSAP core bundle có chuỗi ScrollTrigger nhưng trang không import/register ScrollTrigger, không scroll tween/pin. picture-develops là moment film hợp lệ cho service và chỉ chạy một lần.

## Giới hạn và evidence

- Playwright 5/5 PASS: menu focus/Escape, reduced motion, không tràn ngang ở320/390/768/820/1024/1440, catalog/checkout/lookup/Admin. Không thay taste review bằng browser smoke.
- Gate có các finding yêu cầu cấu trúc kit và screen grid cứng, gồm skip link và footer3 hàng; sản phẩm đang giữ yêu cầu accessibility/chính sách/demo và SSR. Chưa xin waiver thêm hoặc tuyên bố các finding đều false.
- Ảnh điểm đến do Phương Huy công bố public domain; chưa có bộ ảnh riêng cho từng tour từ team nên cover cùng điểm đến có thể trùng. Poster do user cung cấp, không crop chữ.
- Các giả định giá, lịch, tuổi, cọc, policy, nội dung tour khác mẫu đều có nhãn demo và sửa được qua Admin. Canonical localhost dành local; SITE_URL cần cấu hình khi deploy. Không có form nhận .example email hay cam kết gửi thật.
- Chưa có R2/Render/Neon deployment thực, chưa kiểm trên thiết bị vật lý hay UAT khách hàng.

## Kiểm chứng bổ sung

- Browser mới6/6 PASS15.0s: có Leaflet marker/đổi điểm/đóng map (tile mock), home search, advanced facets, blog category và gallery empty/image. Không dùng kết quả tile mock để khẳng định tọa độ đã xác minh hoặc provider luôn hoạt động.
- Sau tối ưu Next Image: perf1440/390 **0 FAIL,4 WARN**. Transfer801.7/534KiB; images472/209.3KiB; fonts4files66.5KiB; CLS0.015/0.000. Cảnh báo JS parsed622.5KiB (203.5KiB gzip) và logo fixed nav eager bị đo như dưới fold. Logo eager là chủ ý, không lazy thương hiệu trên nav. Timing CPU chỉ tương đối, không phải số đo thiết bị thật.
- Hai FAIL oversized-image của lần đo đầu đã hết sau resize bằng Next Image; signed private URL có query bị image optimizer từ chối để không kéo dài TTL qua cache công khai. Tệp gốc giữ nguyên.
- Bổ sung chức năng home/blog/media/surcharge diễn ra sau vòng taste3. Không chạy lại gate để vượt giới hạn3 vòng; 19 finding là bản ghi lần3, chưa phải danh sách nghiệm thu thẩm mỹ đầy đủ của giao diện cuối. T703 tiếp tục ở M7.
- Một warning ESLint stylesheet là link tới CSS vendor taste-skill được giữ nguyên theo hướng dẫn skill; không sửa vendor để che warning. Native dialog, font Việt, focus và reduced motion có test browser. Chưa có visual UAT trên thiết bị vật lý.

## Rà lại ở M7

Đã xem ảnh trang chủ 390/768/1440 từ Chromium CI run 45024473766. Đã điều chỉnh tương phản chữ gợi ý ô nhập, cân dòng tiêu đề và canh skip link khi focus. Ca release-ui kiểm màn hình, phông hiển thị, ảnh hero, kiểm tra tương phản ô tìm tour, bàn phím, reduced motion và bản đồ liên hệ. Ảnh sinh trong CI chỉ là bằng chứng responsive. Giữ nguyên kết quả gate trước và đưa nghiệm thu thẩm mỹ độc lập vào phiếu UAT.
