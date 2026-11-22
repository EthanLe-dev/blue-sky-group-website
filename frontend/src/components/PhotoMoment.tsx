"use client";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useRef } from "react";
export default function PhotoMoment() {
  const ref = useRef<HTMLElement>(null);
  useEffect(() => {
    const element = ref.current as
      | (HTMLElement & { __moment?: { destroy?: () => void } })
      | null;
    const script = document.createElement("script");
    script.type = "module";
    script.textContent =
      "import {mount} from '/assets/moments/moments.js';document.querySelectorAll('[data-moment]').forEach(el=>{if(el.isConnected)mount(el).then(api=>{if(!el.isConnected)api?.destroy?.()})});";
    document.body.append(script);
    return () => {
      element?.__moment?.destroy?.();
      script.remove();
    };
  }, []);
  return (
    <section className="section coast-section">
      <div className="wrap">
        <div className="section-head">
          <h2>Ninh Chữ, từ bờ cát đến làng nghề.</h2>
          <Link className="text-action" href="/tours/ninh-chu-2-ngay-1-dem">
            Xem lịch trình 2 ngày ↗
          </Link>
        </div>
        <figure
          ref={ref}
          className="frame frame--photo coast-photo"
          data-moment="picture-develops"
          data-moment-opts='{"look":"film"}'
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
