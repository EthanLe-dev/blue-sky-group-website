import Map from "@/components/Map";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import {
  read,
  DataError,
  money,
  date,
  site,
  type Tour,
  type Departure,
} from "@/lib/catalog";
import { Gallery, TourTools } from "@/components/TourTools";
async function get(slug: string) {
  try {
    return await read<Tour>("/tours/" + encodeURIComponent(slug));
  } catch (e) {
    if (e instanceof DataError && e.status === 404) notFound();
    throw e;
  }
}
export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const t = await get(slug);
  return {
    title: t.seo_title || t.title,
    description: t.seo_description || t.summary,
    alternates: { canonical: "/tours/" + t.slug },
    openGraph: {
      title: t.title,
      description: t.summary,
      images: t.cover ? [t.cover.url] : [],
    },
  };
}
export default async function Detail({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const t = await get(slug);
  const departures = await read<Departure[]>(
    "/tours/" + encodeURIComponent(slug) + "/departures",
  );
  const reviews = await read<
    {
      id: string;
      display_name: string;
      rating: number;
      body: string;
      is_demo: boolean;
    }[]
  >("/tours/" + encodeURIComponent(slug) + "/reviews");
  const schema = {
    "@context": "https://schema.org",
    "@type": "TouristTrip",
    name: t.title,
    description: t.summary,
    url: site + "/tours/" + t.slug,
    image: t.cover ? site + t.cover.url : undefined,
    itinerary: t.days?.map((d) => ({
      "@type": "TouristAttraction",
      name: d.title,
    })),
  };
  return (
    <main id="main" className="page wrap">
      <nav className="breadcrumb" aria-label="Đường dẫn">
        <Link href="/tours">Khám phá tour</Link>
        <span> / {t.title}</span>
      </nav>
      <p className="eyebrow">
        {t.duration_days} NGÀY · {t.duration_nights} ĐÊM{" "}
        {t.is_demo ? " / NỘI DUNG DEMO" : ""}
      </p>
      <h1>{t.title}</h1>
      <p className="lead">{t.summary}</p>
      <TourTools slug={t.slug} title={t.title} />
      <Gallery pictures={t.gallery || []} poster={t.poster} />
      <div className="detail-grid">
        <div>
          <section className="detail-section">
            <h2>Tổng quan hành trình</h2>
            <div
              className="prose"
              dangerouslySetInnerHTML={{ __html: t.description }}
            />
          </section>
          <section className="detail-section">
            <h2>Lịch trình từng ngày</h2>
            {t.days?.map((d) => (
              <article className="itinerary" key={d.id}>
                <p className="eyebrow">NGÀY {d.day_number}</p>
                <h3>{d.title}</h3>
                <div
                  className="prose"
                  dangerouslySetInnerHTML={{ __html: d.body }}
                />
              </article>
            ))}
          </section>
          <section className="detail-section">
            <h2>Giá tour bao gồm</h2>
            <div
              className="prose"
              dangerouslySetInnerHTML={{ __html: t.included }}
            />
            <h3>Chi phí chưa bao gồm</h3>
            <div
              className="prose"
              dangerouslySetInnerHTML={{ __html: t.excluded }}
            />
          </section>
          <section className="detail-section">
            <h2>Chính sách huỷ tour</h2>
            <div
              className="prose"
              dangerouslySetInnerHTML={{ __html: t.cancellation_policy }}
            />
          </section>
          {t.faqs?.length > 0 && (
            <section className="detail-section">
              <h2>Câu hỏi về hành trình</h2>
              {t.faqs.map((f) => (
                <article key={f.id}>
                  <h3>{f.question}</h3>
                  <div
                    className="prose"
                    dangerouslySetInnerHTML={{ __html: f.answer }}
                  />
                </article>
              ))}
            </section>
          )}
          <section className="detail-section">
            <h2>Điểm dừng trong chuyến đi</h2>
            <Map
              points={(t.stops || [])
                .filter((s) => s.latitude != null && s.longitude != null)
                .map((s) => ({
                  name: s.name,
                  latitude: s.latitude!,
                  longitude: s.longitude!,
                }))}
            />
            <ul>
              {t.stops?.map((s) => (
                <li key={s.id}>
                  {s.name}
                  {s.latitude != null && s.longitude != null && (
                    <>
                      {" "}
                      ·{" "}
                      <a
                        target="_blank"
                        rel="noreferrer"
                        href={
                          "https://www.openstreetmap.org/?mlat=" +
                          s.latitude +
                          "&mlon=" +
                          s.longitude +
                          "#map=13/" +
                          s.latitude +
                          "/" +
                          s.longitude
                        }
                      >
                        Xem vị trí trên OpenStreetMap
                      </a>{" "}
                      (vị trí demo cần xác minh)
                    </>
                  )}
                </li>
              ))}
            </ul>
          </section>
          <section className="detail-section">
            <h2>Đánh giá hành trình</h2>
            {reviews.length ? (
              reviews.map((r) => (
                <article key={r.id}>
                  <h3>{r.display_name}</h3>
                  <p>
                    {r.rating}/5 {r.is_demo ? "· Đánh giá demo" : ""}
                  </p>
                  <div
                    className="prose"
                    dangerouslySetInnerHTML={{ __html: r.body }}
                  />
                </article>
              ))
            ) : (
              <p>Chưa có đánh giá được xác thực cho hành trình này.</p>
            )}
          </section>
          <a className="text-action" href={`/api/v1/tours/${encodeURIComponent(t.slug)}/itinerary.pdf`}>
            Tải PDF lịch trình
          </a>
        </div>
        <aside className="booking-panel" id="departures">
          <p className="eyebrow">CHỌN LỊCH KHỞI HÀNH</p>
          <h2>Từ {money(t.price_from)}</h2>
          <p>/ khách người lớn · Dữ liệu demo</p>
          {departures.length ? (
            departures.map((d) => (
              <article className="departure" key={d.id}>
                <h3>{date(d.starts_at)}</h3>
                <p>
                  {d.remaining_seats} chỗ còn lại · Cọc {d.deposit_bps / 100}%
                </p>
                <ul>
                  {d.prices?.map((p) => (
                    <li key={p.passenger_type}>
                      {{ adult: "Người lớn", child: "Trẻ em", infant: "Em bé" }[
                        p.passenger_type
                      ] || p.passenger_type}
                      : {money(p.unit_price)}
                      <small>
                        {p.min_age}–
                        {p.max_age_exclusive == null
                          ? "trở lên"
                          : p.max_age_exclusive - 1}{" "}
                        tuổi · {p.seats_per_person} chỗ
                      </small>
                    </li>
                  ))}
                </ul>
                {d.status === "open" && d.remaining_seats > 0 ? (
                  <Link
                    className="btn"
                    href={
                      "/checkout?tour=" +
                      encodeURIComponent(t.slug) +
                      "&departure=" +
                      d.id
                    }
                  >
                    Đặt tour ngày này ↗
                  </Link>
                ) : (
                  <button disabled>
                    {d.remaining_seats <= 0 ? "Đã hết chỗ" : "Đã đóng bán"}
                  </button>
                )}
              </article>
            ))
          ) : (
            <p>Chưa có lịch mở bán. Hãy liên hệ để được tư vấn.</p>
          )}
          <Link href="/contact" className="text-action">
            Liên hệ hỗ trợ
          </Link>
          <Link href={`/contact?tour=${t.id}#goi-lai`} className="text-action">
            Yêu cầu gọi lại
          </Link>
        </aside>
      </div>
      <div className="mobile-booking">
        <span>
          Từ <strong>{money(t.price_from)}</strong>
        </span>
        <a className="btn" href="#departures">
          Chọn ngày đi
        </a>
      </div>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(schema).replace(/</g, "\\u003c"),
        }}
      />
    </main>
  );
}
