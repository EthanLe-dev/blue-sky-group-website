"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

export default function HomeActions({ phoneHref }: { phoneHref?: string }) {
  const [pastIntroduction, setPastIntroduction] = useState(false);
  useEffect(() => {
    const end = document.getElementById("home-about-end");
    if (!end) return;
    let frame = 0;
    const update = () => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => setPastIntroduction(end.getBoundingClientRect().bottom <= 0));
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
      {phoneHref
        ? <a className="home-button" href={phoneHref}>Gọi hỗ trợ</a>
        : <Link className="home-button" href="/contact" prefetch={false}>Hỗ trợ</Link>}
    </nav>
  );
}
