import Image from "next/image";
import Link from "next/link";
import { read, type Tour, money } from "@/lib/catalog";
import { resolveContact } from "@/lib/contact";
import type { CompanyProfile } from "./HomeFooter";
import ContactIcon from "./ContactIcon";

export default async function TourCard({ tour, imageSizes, home = false }: { tour: Tour; imageSizes?: string; home?: boolean }) {
  const settings = await read<{ company_profile?: CompanyProfile }>("/site-settings").catch(() => null);
  const contact = resolveContact(settings?.company_profile);
  const href = "/tours/" + tour.slug;
  const tracking = { "data-select-tour": tour.id, "data-select-position": home ? "home" : "catalog" };
  return (
    <article className="card tour-card conversion-tour-card">
      <Link prefetch={false} className="tour-card-photo" href={href} aria-label={"Xem tour " + tour.title} {...tracking}>
        <div className="card-image">
          {tour.cover ? <Image src={tour.cover.url} sizes={imageSizes ?? "(max-width:540px) 100vw, (max-width:850px) 50vw, 33vw"}
            alt={tour.cover.alt_text || tour.title} width={640} height={440} loading="lazy" />
            : <span className="tour-cover-empty">Ảnh tour đang cập nhật</span>}
          <div className="tour-card-badges">
            <span>{tour.duration_days} ngày{tour.duration_nights > 0 && ` · ${tour.duration_nights} đêm`}</span>
            {tour.is_demo && <span className="tour-card-demo">Tour demo</span>}
          </div>
        </div>
      </Link>
      <div className="card-body">
        {tour.origin_name && <p className="tour-card-origin">Khởi hành: {tour.origin_name}</p>}
        <h3><Link prefetch={false} href={href} {...tracking}>{tour.title}</Link></h3>
        <p className="tour-card-summary">{tour.summary}</p>
        <div className="tour-card-price">
          {tour.price_from == null ? <strong>Liên hệ báo giá</strong> : <><span>Giá từ</span><strong>{money(tour.price_from)}</strong><span>/ khách</span></>}
        </div>
        <div className="tour-card-actions">
          <Link className="btn" prefetch={false} href={href + "#departures"} {...tracking}>Xem lịch & giá <span aria-hidden="true">↗</span></Link>
          <a className="tour-card-zalo" href={contact.zaloHref} target="_blank" rel="noopener noreferrer"
            aria-label={"Tư vấn qua Zalo · " + tour.title} data-contact-channel="zalo" data-contact-position="tour_card" data-contact-tour={tour.id}>
            <ContactIcon channel="zalo" />Tư vấn qua Zalo
          </a>
        </div>
      </div>
    </article>
  );
}
