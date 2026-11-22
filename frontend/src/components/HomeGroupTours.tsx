"use client";

import Image from "next/image";
import Link from "next/link";
import { useRef } from "react";
import { useGSAP } from "@gsap/react";
import gsap from "gsap";

gsap.registerPlugin(useGSAP);

const services = [
  {
    route: "/schools",
    audience: "trường học",
    title: "Chuyến đi cho trường học",
    text: "Tổ chức ngoại khóa và tham quan phù hợp với lứa tuổi, quy mô lớp và kế hoạch của nhà trường.",
    image: "/images/group-school.webp",
    height: 800,
    alt: "Ảnh minh họa học sinh đi ngoại khóa trong công viên",
  },
  {
    route: "/companies",
    audience: "doanh nghiệp",
    title: "Chuyến đi cho doanh nghiệp",
    text: "Lên chương trình du lịch và hoạt động cho đội ngũ theo số lượng, thời gian và ngân sách của đoàn.",
    image: "/images/group-company.webp",
    height: 798,
    alt: "Ảnh minh họa một nhóm người tham gia trò chơi vận động ngoài trời",
  },
];

export default function HomeGroupTours() {
  const section = useRef<HTMLElement>(null);
  useGSAP(() => {
    if (!section.current || typeof IntersectionObserver === "undefined") return;
    const cards = [...section.current.querySelectorAll(".home-group-card")];
    const media = gsap.matchMedia();
    media.add("(prefers-reduced-motion: no-preference)", () => {
      const reveals = new Map(cards.map((card, index) => [card, gsap.timeline({ paused: true }).from(card, {
        y: 18,
        opacity: 0.72,
        duration: 0.65,
        ease: "power3.out",
        immediateRender: false,
        clearProps: "transform,opacity",
      }, index * 0.12)]));
      const observer = new IntersectionObserver((entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue;
          reveals.get(entry.target)?.play();
          observer.unobserve(entry.target);
        }
      }, { threshold: 0.16 });
      cards.forEach((card) => observer.observe(card));
      return () => observer.disconnect();
    });
    return () => media.revert();
  }, { scope: section });

  return (
    <section ref={section} className="wrap home-block" aria-labelledby="groups-heading">
      <h2 id="groups-heading">Tour theo đoàn</h2>
      <div className="home-group-grid">
        {services.map(({ route, audience, title, text, image, height, alt }) => (
          <article className="home-group-card" key={route}>
            <div className="home-group-photo">
              <Image src={image} alt={alt} width={1200} height={height}
                sizes="(max-width:680px) 100vw, (max-width:1344px) 50vw, 640px"
                loading="lazy" data-source="stand-in" data-subject={audience} />
              <span aria-hidden="true">Ảnh minh họa</span>
            </div>
            <div className="home-group-content">
              <h3><Link prefetch={false} href={route}>{title}</Link></h3>
              <p>{text}</p>
              <Link className="home-button" prefetch={false} href={route + "#bao-gia"}>
                Yêu cầu báo giá <span className="sr-only">{audience}</span>
                <span className="home-group-arrow" aria-hidden="true">↗</span>
              </Link>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
