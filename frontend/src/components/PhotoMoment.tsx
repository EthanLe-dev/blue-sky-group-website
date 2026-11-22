"use client";
import Image from "next/image";
import Link from "next/link";
import { useRef } from "react";
import { useGSAP } from "@gsap/react";
import gsap from "gsap";

gsap.registerPlugin(useGSAP);

export default function PhotoMoment() {
  const ref = useRef<HTMLElement>(null);
  useGSAP(() => {
    const photo = ref.current;
    if (!photo || typeof IntersectionObserver === "undefined") return;
    const media = gsap.matchMedia();
    media.add("(prefers-reduced-motion: no-preference)", () => {
      const reveal = gsap.from("img", {
        scale: 1.09,
        opacity: 0.75,
        duration: 1.15,
        ease: "power3.out",
        paused: true,
        immediateRender: false,
      });
      const observer = new IntersectionObserver(([entry]) => {
        if (!entry?.isIntersecting) return;
        observer.disconnect();
        reveal.play();
      }, { threshold: 0.25 });
      observer.observe(photo);
      return () => observer.disconnect();
    });
    return () => media.revert();
  }, { scope: ref });
  return (
    <section className="section coast-section">
      <div className="wrap">
        <div className="section-head">
          <h2>Ninh Chữ, từ bờ cát đến làng nghề.</h2>
          <Link prefetch={false} className="text-action" href="/destinations">
            Xem các điểm đến ↗
          </Link>
        </div>
        <figure
          ref={ref}
          className="frame frame--photo coast-photo"
        >
          <Image
            src="/images/ninh-chu.jpg"
            sizes="100vw"
            alt="Bờ biển Ninh Chữ trải dài dưới bầu trời xanh"
            width="1280"
            height="960"
            loading="lazy"
            data-source="Wikimedia Commons / Phương Huy / Public domain"
            data-subject="Bãi biển Ninh Chữ"
          />
        </figure>
      </div>
    </section>
  );
}
