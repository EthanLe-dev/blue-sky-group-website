import Image from "next/image";
import ContactIcon from "./ContactIcon";
import type { ContactLinks } from "@/lib/contact";

export default function ContactClosing({ contact }: { contact: ContactLinks }) {
  return (
    <section className="contact-closing wrap" aria-labelledby="contact-closing-title">
      <div className="contact-closing-photo">
        <Image src="/images/about-hoi-an.webp" width={1200} height={800} alt="Đèn lồng rực sáng trên phố Hội An" loading="lazy" sizes="(max-width:680px) 100vw, 40vw" />
      </div>
      <div className="contact-closing-copy">
        <p className="eyebrow">CÙNG BLUE SKY GROUP</p>
        <h2 id="contact-closing-title">Bạn đang lên kế hoạch cho chuyến đi?</h2>
        <p>Chia sẻ điểm đến, ngày đi và số người. Blue Sky Group sẽ cùng bạn tìm chương trình phù hợp.</p>
        <a className="cta-zalo" href={contact.zaloHref} target="_blank" rel="noopener noreferrer" data-contact-channel="zalo" data-contact-position="closing"><ContactIcon channel="zalo" />Nhắn Zalo tư vấn <span aria-hidden="true">↗</span></a>
        <a className="closing-hotline" href={contact.phoneHref} data-contact-channel="phone" data-contact-position="closing">Hoặc gọi {contact.hotline}</a>
      </div>
    </section>
  );
}
