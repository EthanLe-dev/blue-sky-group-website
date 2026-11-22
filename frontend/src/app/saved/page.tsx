"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
type Item = { slug: string; title: string };
export default function Saved() {
  const [data, setData] = useState<{ saved: Item[]; recent: Item[] }>({
    saved: [],
    recent: [],
  });
  const [message, setMessage] = useState("Đang đọc danh sách…");
  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const load = async (key: string) => {
          const raw = JSON.parse(localStorage.getItem(key) || "[]");
          if (!Array.isArray(raw)) return [];
          const slugs = raw
            .filter((x) => typeof x === "string" && /^[a-z0-9-]+$/.test(x))
            .slice(0, 50);
          const results = await Promise.allSettled(
            slugs.map((s) => api<Item>("/tours/" + s)),
          );
          return results.flatMap((r) =>
            r.status === "fulfilled" ? [r.value] : [],
          );
        };
        const [saved, recent] = await Promise.all([
          load("favourite-tours"),
          load("recent-tours"),
        ]);
        if (active) {
          setData({ saved, recent });
          setMessage("");
        }
      } catch {
        if (active)
          setMessage(
            "Không đọc được lưu trữ trình duyệt. Bạn vẫn có thể xem danh mục tour.",
          );
      }
    })();
    return () => {
      active = false;
    };
  }, []);
  return (
    <main id="main" className="page wrap">
      <h1>Hành trình đã lưu</h1>
      <p role="status">{message}</p>
      {(["saved", "recent"] as const).map((key) => (
        <section className="detail-section" key={key}>
          <h2>{key === "saved" ? "Tour yêu thích" : "Tour vừa xem"}</h2>
          {data[key].length ? (
            <ul>
              {data[key].map((t) => (
                <li key={t.slug}>
                  <Link href={"/tours/" + t.slug}>{t.title}</Link>
                </li>
              ))}
            </ul>
          ) : (
            <p>Chưa có tour trong danh sách.</p>
          )}
        </section>
      ))}
      <Link className="btn" href="/tours">
        Khám phá tour
      </Link>
    </main>
  );
}
