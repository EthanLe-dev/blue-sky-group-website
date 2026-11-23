"use client";

import { useEffect, useState } from "react";
import ContactIcon from "./ContactIcon";
import { companyContact } from "@/lib/contact";

export default function HomeActions({ phoneHref, zaloHref }: { phoneHref: string; zaloHref: string }) {
  const [pastIntroduction, setPastIntroduction] = useState(false);
  useEffect(() => {
    const end = document.getElementById("home-about-end");
    if (!end) return;
    let frame = 0;
    const update = () => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => setPastIntroduction(end.getBoundingClientRect().bottom <= window.innerHeight));
    };
    update();
    window.addEventListener("scroll", update, { passive: true });
    window.addEventListener("resize", update);
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener("scroll", update);
      window.removeEventListener("resize", update);
    };
  }, []);
  return (
    <nav className="home-actions" aria-label="Hỗ trợ nhanh">
      {pastIntroduction && <a className="home-back-top" href="#home-top" aria-label="Lên đầu trang">↑</a>}
      <a className="contact-float contact-float-zalo" href={zaloHref} target="_blank" rel="noopener noreferrer" aria-label="Nhắn Zalo Blue Sky Group (mở tab mới)">
        <ContactIcon channel="zalo" />
        <span className="contact-tooltip">Nhắn Zalo</span>
      </a>
      <a className="contact-float contact-float-facebook" href={companyContact.facebookHref} target="_blank" rel="noopener noreferrer" aria-label="Facebook Blue Sky Group (mở tab mới)">
        <ContactIcon channel="facebook" />
        <span className="contact-tooltip">Facebook</span>
      </a>
      <a className="contact-float contact-float-phone" href={phoneHref} aria-label={"Gọi hotline " + phoneHref.slice(4)}>
        <ContactIcon channel="phone" />
        <span className="contact-tooltip">Gọi hotline</span>
      </a>
    </nav>
  );
}
