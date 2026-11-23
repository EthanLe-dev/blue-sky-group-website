import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import Hero from "@/components/Hero";
import HomeActions from "@/components/HomeActions";
import { HomeFooter, type CompanyProfile } from "@/components/HomeFooter";
import HomeGroupTours from "@/components/HomeGroupTours";
import PhotoMoment from "@/components/PhotoMoment";
import { Header } from "@/components/Shell";
import TourCard from "@/components/TourCard";
import { read, type Content } from "@/lib/catalog";
import { homeCatalog, type HomeData } from "@/lib/home-catalog";
import ContactClosing from "@/components/ContactClosing";
import { resolveContact } from "@/lib/contact";
import "./home.css";

export const dynamic = "force-dynamic";
const title = "Blue Sky Group | Du lịch & Trải nghiệm";
const description = "Tìm tour du lịch và khám phá các chương trình dành cho trường học, doanh nghiệp cùng Blue Sky Group.";
export const metadata: Metadata = {
  title: { absolute: title },
  description,
  alternates: { canonical: "/" },
  openGraph: { title, description, locale: "vi_VN", siteName: "Blue Sky Group", type: "website", images: ["/images/vinh-hy.jpg"] },
};

export default async function Home() {
  const [data, homeCopy, settings] = await Promise.all([
    read<HomeData>("/home").catch(() => null),
    read<Content>("/pages/home").catch(() => null),
    read<{ company_profile?: CompanyProfile }>("/site-settings").catch(() => null),
  ]);
  const { state, tours, destinations, banners, reviews } = homeCatalog(data);
  const company = settings?.company_profile;
  const contact = resolveContact(company);
  const zaloHref = contact.zaloHref;

  return (
    <div className="home-shell" id="home-top" tabIndex={-1}>
      <Header home contact={{ ...contact, zaloHref }} />
      <noscript><style>{`
        .home-shell .home-header { position: relative; }
        .home-shell .home-header > .wrap { flex-wrap: wrap; }
        .home-shell .home-header > .wrap > nav { display: flex; flex-wrap: wrap; gap: 4px 16px; }
        .home-shell .home-header .menu-toggle { display: none; }
      `}</style></noscript>
      <main id="main" className="home" tabIndex={-1}>
        <Hero zaloHref={zaloHref} destinations={destinations.map(({ slug, name }) => ({ slug, name }))} />

        {banners.length > 0 && (
          <section className="wrap home-block" aria-labelledby="campaign-heading">
            <h2 id="campaign-heading">Tin và ưu đãi</h2>
            <div className="home-banner-grid">
              {banners.map((banner) => {
                const src = `/api/v1/media/${banner.media_id}`;
                const href = banner.target_url;
                const picture = (
                  <Image src={src} alt={banner.title} width={800} height={360}
                    sizes="(max-width:680px) 100vw, 50vw" loading="lazy" />
                );
                return (
                  <article className="home-banner-tile" key={banner.id}>
                    {href && /^(?:\/(?![\/\\])|https?:\/\/)/i.test(href)
                      ? <Link href={href} prefetch={false}>{picture}</Link>
                      : picture}
                  </article>
                );
              })}
            </div>
          </section>
        )}

        <section className="wrap home-block home-tour-section" id="journeys" aria-labelledby="journeys-heading">
          <div className="home-section-heading">
            <div>
              <h2 id="journeys-heading">Khám phá các tour</h2>
              <p>Chọn điểm đến và xem giá, lịch khởi hành của từng hành trình.</p>
            </div>
            <Link className="home-outline" prefetch={false} href="/tours">Xem tất cả tour <span aria-hidden="true">↗</span></Link>
          </div>
          {destinations.length > 0 && (
            <nav className="home-destination-links" aria-label="Lọc tour theo điểm đến">
              {destinations.map(({ slug, name }) => (
                <Link prefetch={false} key={slug} href={"/tours?" + new URLSearchParams({ destination: slug })}>{name}</Link>
              ))}
            </nav>
          )}
          {state === "ready" ? (
            <div className="home-tour-grid">
              {tours.map((tour) => (
                <TourCard home key={tour.id} tour={tour}
                  imageSizes="(max-width:680px) 100vw, (max-width:1100px) 50vw, 25vw" />
              ))}
            </div>
          ) : (
            <div className="home-empty">
              <h3>{state === "empty" ? "Chưa có tour để hiển thị" : "Chưa tải được danh mục tour"}</h3>
              <p>{state === "empty" ? "Bạn có thể liên hệ để được tư vấn chuyến đi." : "Vui lòng thử mở lại danh mục hoặc liên hệ để được hỗ trợ."}</p>
              <Link className="home-outline" prefetch={false} href={state === "empty" ? "/contact" : "/tours"}>{state === "empty" ? "Liên hệ Blue Sky Group" : "Mở danh mục tour"}</Link>
            </div>
          )}
        </section>

        <HomeGroupTours />

        <section className="wrap home-block home-about" aria-labelledby="about-heading">
          <div className="home-about-copy">
            <h2 id="about-heading">Về Blue Sky Group</h2>
            <p>Blue Sky Group tổ chức tour du lịch và các chuyến đi theo yêu cầu cho gia đình, trường học và doanh nghiệp.</p>
            {homeCopy?.body && (
              <div className="prose home-about-note" dangerouslySetInnerHTML={{ __html: homeCopy.body }} />
            )}
            <Link className="btn home-button" prefetch={false} href="/about">Tìm hiểu về Blue Sky Group <span aria-hidden="true">↗</span></Link>
          </div>
          <PhotoMoment />
          <span id="home-about-end" aria-hidden="true" />
        </section>

        {reviews.length > 0 && (
          <section className="wrap home-block" aria-labelledby="reviews-heading">
            <h2 id="reviews-heading">Đánh giá được chia sẻ</h2>
            <div className="home-group-grid home-review-grid">
              {reviews.map((review) => (
                <article key={review.id}>
                  <h3>{review.display_name}</h3>
                  <p>{review.rating}/5{review.is_demo ? " · Đánh giá demo" : ""}</p>
                  <div className="prose" dangerouslySetInnerHTML={{ __html: review.body }} />
                </article>
              ))}
            </div>
          </section>
        )}
        <ContactClosing contact={contact} />
      </main>
      <HomeFooter company={company} contact={contact} />
      <HomeActions phoneHref={contact.phoneHref} zaloHref={zaloHref} />
    </div>
  );
}
