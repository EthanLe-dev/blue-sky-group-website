import Image from "next/image";
import Link from "next/link";

export type CompanyProfile = {
  display_name?: string;
  address?: string | null;
  hotline?: string | null;
  email?: string | null;
  zalo_url?: string | null;
  is_demo?: boolean;
};

const footerLinks = [
  { title: "Khám phá", links: [["/tours", "Danh mục tour"], ["/destinations", "Điểm đến"], ["/schools", "Tour trường học"], ["/companies", "Tour doanh nghiệp"]] },
  { title: "Hỗ trợ", links: [["/contact", "Liên hệ"], ["/lookup", "Tra cứu đơn"], ["/faq", "Câu hỏi thường gặp"], ["/about", "Về Blue Sky Group"]] },
  { title: "Chính sách", links: [["/policies/privacy", "Bảo mật thông tin"], ["/policies/terms", "Điều khoản sử dụng"], ["/policies/payment", "Thanh toán"], ["/policies/cancellation", "Hủy tour"], ["/policies/preparation", "Chuẩn bị chuyến đi"]] },
];

export function HomeFooter({ company, contact }: {
  company?: CompanyProfile;
  contact: { hotline: string; phoneHref: string } | null;
}) {
  return (
    <footer className="home-footer">
      <div className="wrap">
        <div className="home-footer-grid">
          <div className="home-footer-brand">
            <Link className="brand" href="/" prefetch={false}>
              <Image src="/brand/logo.jpg" alt="" width={48} height={48} />
              <span>{company?.display_name || "Blue Sky Group"}</span>
            </Link>
            {company?.address && <p>{company.address}</p>}
            {company?.is_demo && <p className="home-demo-note">Thông tin công ty dùng cho demo.</p>}
            {contact && <a href={contact.phoneHref} data-contact-channel="phone" data-contact-position="footer">Gọi {contact.hotline}</a>}
            {company?.email && company.is_demo === false && <a href={"mailto:" + company.email}>{company.email}</a>}
            {company?.zalo_url && company.is_demo === false && /^https:\/\/zalo\.me\//i.test(company.zalo_url) && (
              <a href={company.zalo_url} data-contact-channel="zalo" data-contact-position="footer">Liên hệ qua Zalo</a>
            )}
          </div>
          {footerLinks.map(({ title, links }) => (
            <nav aria-label={title} key={title}>
              <h2>{title}</h2>
              {links.map(([href, label]) => <Link href={href} prefetch={false} key={href}>{label}</Link>)}
            </nav>
          ))}
        </div>
        <div className="home-footer-bottom">
          <span>© {new Date().getFullYear()} Blue Sky Group</span>
          <Link href="/credits" prefetch={false}>Nguồn ảnh và quyền sử dụng</Link>
        </div>
        {process.env.APP_ENV !== "production" && (
          <p className="home-demo-note">Bản demo · Vui lòng xác nhận giá và lịch khởi hành trước khi đặt.</p>
        )}
      </div>
    </footer>
  );
}
