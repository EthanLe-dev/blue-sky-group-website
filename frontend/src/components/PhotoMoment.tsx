"use client";
import Image from "next/image";
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
    <figure ref={ref} className="home-about-photo">
      <Image src="/images/ninh-chu.jpg" sizes="(max-width:850px) 100vw, 50vw"
        alt="Bờ biển Ninh Chữ trải dài dưới bầu trời xanh" width="1280" height="960"
        loading="lazy" data-source="Wikimedia Commons / Phương Huy / Public domain" />
    </figure>
  );
}
