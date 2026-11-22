"use client";
import Image from "next/image";
import { useRef } from "react";
import { useGSAP } from "@gsap/react";
import gsap from "gsap";
gsap.registerPlugin(useGSAP);
export default function Hero() {
  const scope = useRef<HTMLElement>(null);
  useGSAP(
    () => {
      const mm = gsap.matchMedia();
      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.timeline({ defaults: { ease: "power3.out" } })
          .from(".hero-photo", { scale: 1.045, duration: 1.1 })
          .from("[data-enter-text]", {
            y: 14,
            opacity: 0.25,
            duration: 0.58,
            stagger: 0.11,
          }, 0);
      });
      return () => mm.revert();
    },
    { scope },
  );
  return (
    <section ref={scope} className="hero hero--photo" data-enter="">
      <Image
        sizes="100vw"
        className="hero-photo"
        src="/images/vinh-hy.jpg"
        width="1920"
        height="1280"
        alt="Vịnh Vĩnh Hy với biển xanh, thuyền neo và núi ven bờ"
        fetchPriority="high"
        loading="eager"
      />
      <div className="wrap">
        <div className="hero-copy">
          <h1 data-enter-text="">
            Khám phá những hành trình cùng Blue Sky Group
          </h1>
          <p data-enter-text="">
            Tìm chuyến đi cho gia đình hoặc khám phá các chương trình dành cho
            trường học và doanh nghiệp.
          </p>
          <form className="hero-search" action="/tours">
            <label htmlFor="hero-q">Bạn muốn đi đâu?</label>
            <input
              id="hero-q"
              name="q"
              maxLength={100}
              placeholder="Nhập điểm đến bạn muốn khám phá"
            />
            <button className="btn">Tìm tour ↗</button>
          </form>
        </div>
      </div>
    </section>
  );
}
