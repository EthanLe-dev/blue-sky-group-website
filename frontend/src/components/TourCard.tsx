import Image from "next/image";
import Link from "next/link";
import type { Tour } from "@/lib/catalog";
import { money } from "@/lib/catalog";
export default function TourCard({ tour, imageSizes, home = false }: { tour: Tour; imageSizes?: string; home?: boolean }) {
  return (
    <article className="card tour-card">
      <Link prefetch={false} className="m-zoom" href={"/tours/" + tour.slug}>
        <div className="card-image">
          {tour.cover ? (
            <Image
              src={tour.cover.url}
              sizes={imageSizes ?? "(max-width:540px) 100vw, (max-width:850px) 50vw, 33vw"}
              alt={tour.cover.alt_text || tour.title}
              width="640"
              height="440"
              loading="lazy"
            />
          ) : (
            <span className="tour-cover-empty">Ảnh tour đang cập nhật</span>
          )}
          {home && (
            <div className="home-tour-badges">
              <span className="home-tour-duration">
                {tour.duration_days} ngày{tour.duration_nights > 0 && ` · ${tour.duration_nights} đêm`}
              </span>
              {tour.is_demo && <span className="home-tour-demo">Tour demo</span>}
            </div>
          )}
        </div>
        <div className="card-body">
          {!home && (
            <p className="eyebrow">
              {tour.duration_days} ngày · {tour.duration_nights} đêm{" "}
              {tour.is_demo ? " / Tour demo" : ""}
            </p>
          )}
          <h3>{tour.title}</h3>
          {home && tour.origin_name && (
            <p className="home-tour-origin"><span aria-hidden="true" />Khởi hành: {tour.origin_name}</p>
          )}
          <p>{tour.summary}</p>
          {home ? (
            <div className="home-tour-action">
              <div className="home-tour-price">
                {tour.price_from == null ? <strong>Liên hệ</strong> : (
                  <>
                    <span>Giá từ</span>
                    <strong>{money(tour.price_from)}</strong>
                    <span>/ khách</span>
                  </>
                )}
              </div>
              <span className="home-tour-cta">Xem chi tiết <span aria-hidden="true">↗</span></span>
            </div>
          ) : (
            <div className="card-price">
              <span>Từ <strong>{money(tour.price_from)}</strong> / khách</span>
              <span aria-hidden="true">↗</span>
            </div>
          )}
        </div>
      </Link>
    </article>
  );
}
