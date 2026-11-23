"use client";
import Image from "next/image";
import ContactIcon from "./ContactIcon";
import { useRef } from "react";
import { useGSAP } from "@gsap/react";
import gsap from "gsap";
import HeroVideo from "./HeroVideo";
gsap.registerPlugin(useGSAP);
export default function Hero({ destinations, zaloHref }: { zaloHref: string; destinations: { slug: string; name: string }[] }) {
  const scope = useRef<HTMLElement>(null);
  useGSAP(
    () => {
      const mm = gsap.matchMedia();
      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.from("[data-enter-text]", {
          y: 14,
          opacity: 0.25,
          duration: 0.65,
          stagger: 0.07,
          ease: "power3.out",
        });
      });
      return () => mm.revert();
    },
    { scope },
  );
  return (
    <section ref={scope} className="hero home-hero" data-enter="off" aria-labelledby="hero-title">
      <div className="home-hero-scene">
        <Image
          sizes="100vw"
          className="hero-photo"
          src="/images/vinh-hy.jpg"
          width="1920"
          height="1280"
          alt=""
          fetchPriority="high"
          loading="eager"
        />
        <HeroVideo />
        <div className="wrap home-hero-copy">
          <p className="hero-kicker" data-enter-text="">Du lịch · Sự kiện · Trải nghiệm</p>
          <h1 id="hero-title" data-enter-text="">Kiến tạo hành trình</h1>
          <p data-enter-text="">Tour du lịch cho gia đình, trường học và doanh nghiệp. Cùng Blue Sky Group lên kế hoạch cho chuyến đi của bạn.</p>
          <div className="hero-actions" data-enter-text="">
            <a className="cta-zalo" href={zaloHref} target="_blank" rel="noopener noreferrer" data-contact-channel="zalo" data-contact-position="hero">
              <ContactIcon channel="zalo" />Nhắn Zalo tư vấn <span aria-hidden="true">↗</span>
            </a>
            <a className="hero-browse" href="#journeys">Khám phá tour <span aria-hidden="true">↓</span></a>
          </div>
        </div>
      </div>
      <div className="wrap home-search-panel">
        <form className="home-search" action="/tours" method="get" aria-label="Tìm tour">
          <div className="home-search-field">
            <label htmlFor="hero-q">Từ khóa</label>
            <input id="hero-q" name="q" maxLength={100} placeholder="Nhập tên tour hoặc điểm đến" />
          </div>
          <div className="home-search-field">
            <label htmlFor="hero-destination">Điểm đến</label>
            <select id="hero-destination" name="destination">
              <option value="">Tất cả điểm đến</option>
              {destinations.map(({ slug, name }) => <option key={slug} value={slug}>{name}</option>)}
            </select>
          </div>
          <div className="home-search-field">
            <label htmlFor="hero-departure">Khởi hành từ ngày</label>
            <input id="hero-departure" name="depart_from" type="date" />
          </div>
          <button className="btn home-button" type="submit">Tìm tour <span aria-hidden="true">↗</span></button>
        </form>
      </div>
    </section>
  );
}
