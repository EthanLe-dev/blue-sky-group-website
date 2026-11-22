import Link from "next/link";
import type { Tour } from "@/lib/catalog";
import { money } from "@/lib/catalog";
export default function TourCard({ tour }: { tour: Tour }) {
  return (
    <article className="card tour-card">
      <Link className="m-zoom" href={"/tours/" + tour.slug}>
        <div className="card-image">
          {tour.cover ? (
            <img
              src={tour.cover.url}
              alt={tour.cover.alt_text || tour.title}
              width="640"
              height="440"
              loading="lazy"
            />
          ) : (
            <span>Ảnh tour đang cập nhật</span>
          )}
        </div>
        <div className="card-body">
          <p className="eyebrow">
            {tour.duration_days} ngày · {tour.duration_nights} đêm{" "}
            {tour.is_demo ? " / Tour demo" : ""}
          </p>
          <h3>{tour.title}</h3>
          <p>{tour.summary}</p>
          <div className="card-price">
            <span>
              Từ <strong>{money(tour.price_from)}</strong> / khách
            </span>
            <span aria-hidden="true">↗</span>
          </div>
        </div>
      </Link>
    </article>
  );
}
