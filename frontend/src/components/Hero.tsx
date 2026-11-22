"use client";
import Image from "next/image";
import { useRef } from "react";
import { useGSAP } from "@gsap/react";
import gsap from "gsap";
import HeroVideo from "./HeroVideo";
gsap.registerPlugin(useGSAP);
export default function Hero({ destinations }: { destinations: { slug: string; name: string }[] }) {
  const scope = useRef<HTMLElement>(null);
  useGSAP(
    () => {
      const mm = gsap.matchMedia();
      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.from("[data-enter-text]", {
          y: 14,
          opacity: 0.25,
          duration: 0.58,
          stagger: 0.11,
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
          <h1 id="hero-title" data-enter-text="">Kiến tạo hành trình</h1>
          <p data-enter-text="">Khám phá Việt Nam cùng Blue Sky Group và tìm chuyến đi dành cho bạn.</p>
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
