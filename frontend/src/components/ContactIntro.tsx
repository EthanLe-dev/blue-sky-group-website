import Image from "next/image";
import { read } from "@/lib/catalog";
import { resolveContact } from "@/lib/contact";
import type { CompanyProfile } from "./HomeFooter";
import ContactIcon from "./ContactIcon";

export default async function ContactIntro({ kind }: { kind: "contact" | "schools" | "companies" }) {
  const settings = await read<{ company_profile?: CompanyProfile }>("/site-settings").catch(() => null);
  const contact = resolveContact(settings?.company_profile);
  const image = kind === "schools" ? "group-school.webp" : kind === "companies" ? "group-company.webp" : "about-ninh-binh.webp";
  const text = kind === "schools" ? "Trao đổi chương trình ngoại khóa theo độ tuổi, số học sinh và lịch của nhà trường."
    : kind === "companies" ? "Cùng lên chương trình du lịch và hoạt động đội nhóm theo ngân sách của doanh nghiệp."
    : "Bạn có thể nhắn Zalo, gọi hotline hoặc để lại yêu cầu gọi lại bên dưới.";
  return <section className="contact-intro" aria-label="Kết nối với Blue Sky Group">
    <div className="contact-intro-photo"><Image src={"/images/" + image} alt={kind === "contact" ? "Phong cảnh sông núi Ninh Bình" : "Ảnh minh họa chuyến đi theo đoàn"} width={1200} height={800} sizes="(max-width:680px) 100vw, 40vw" loading="lazy" />{kind !== "contact" && <span>Ảnh minh họa</span>}</div>
    <div className="contact-intro-copy"><h2>Cùng bạn lên kế hoạch</h2><p>{text}</p>
      <a className="cta-zalo" href={contact.zaloHref} target="_blank" rel="noopener noreferrer" data-contact-channel="zalo" data-contact-position="contact_page"><ContactIcon channel="zalo" />Nhắn Zalo tư vấn ↗</a>
      <a className="closing-hotline" href={contact.phoneHref} data-contact-channel="phone" data-contact-position="contact_page">Gọi {contact.hotline}</a>
    </div>
  </section>;
}
