"use client";
import Image from "next/image";
import { useRef } from "react";
import { useGSAP } from "@gsap/react";
import gsap from "gsap";

gsap.registerPlugin(useGSAP);

const photos = [
  { name: "Ninh Bình", image: "/images/about-ninh-binh.webp", width: 1200, height: 800, alt: "Thuyền chèo trên dòng sông giữa núi đá vôi ở Ninh Bình", source: "Pexels / Vishal Chokkala" },
  { name: "Hội An", image: "/images/about-hoi-an.webp", width: 800, height: 543, alt: "Đèn lồng nhiều màu trong phố cổ Hội An", source: "Pexels / Võ Văn Tiến" },
  { name: "Ninh Chữ", image: "/images/ninh-chu.jpg", width: 1280, height: 960, alt: "Mặt trời chiếu xuống bờ biển Ninh Chữ", source: "Wikimedia Commons / Phương Huy / Public domain" },
];

export default function PhotoMoment() {
  const ref = useRef<HTMLElement>(null);
  useGSAP(() => {
    const photo = ref.current;
    if (!photo || typeof IntersectionObserver === "undefined") return;
    const media = gsap.matchMedia();
    media.add("(prefers-reduced-motion: no-preference)", () => {
      const reveal = gsap.timeline({ paused: true })
        .from(".home-about-frame", {
          y: 15,
          opacity: 0.72,
          duration: 0.65,
          stagger: 0.13,
          ease: "power3.out",
          immediateRender: false,
          clearProps: "transform,opacity",
        })
        .from("img", {
          scale: 1.055,
          duration: 0.85,
          stagger: 0.13,
          ease: "power3.out",
          immediateRender: false,
          clearProps: "transform",
        }, 0);
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
    <figure ref={ref} className="home-about-photo" aria-label="Phong cảnh Việt Nam">
      {photos.map(({ name, image, width, height, alt, source }, index) => (
        <div className="home-about-frame" key={name}>
          <Image src={image} alt={alt} width={width} height={height}
            sizes={index === 0 ? "(max-width:680px) 100vw, 52vw" : "(max-width:680px) 50vw, 26vw"}
            loading="lazy" data-source={source} data-subject={name} />
          <span aria-hidden="true">{name}</span>
        </div>
      ))}
    </figure>
  );
}
