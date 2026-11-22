"use client";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import Image from "next/image";
type Picture = { id: string; url: string; alt_text: string; kind: string };
export function TourTools({ slug, title }: { slug: string; title: string }) {
  const [saved, setSaved] = useState(false);
  const [msg, setMsg] = useState("");
  useEffect(() => {
    try {
      const raw = JSON.parse(localStorage.getItem("recent-tours") || "[]");
      const recent = Array.isArray(raw)
        ? raw.filter((x) => typeof x === "string")
        : [];
      localStorage.setItem(
        "recent-tours",
        JSON.stringify(
          [slug, ...recent.filter((x) => x !== slug)].slice(0, 12),
        ),
      );
      const f = JSON.parse(localStorage.getItem("favourite-tours") || "[]");
      queueMicrotask(() => setSaved(Array.isArray(f) && f.includes(slug)));
    } catch {}
  }, [slug]);
  return (
    <div className="tools">
      <button
        aria-pressed={saved}
        onClick={() => {
          try {
            const raw = JSON.parse(
              localStorage.getItem("favourite-tours") || "[]",
            );
            const a: string[] = Array.isArray(raw)
              ? raw.filter((x) => typeof x === "string")
              : [];
            localStorage.setItem(
              "favourite-tours",
              JSON.stringify(
                saved ? a.filter((x) => x !== slug) : [slug, ...a].slice(0, 50),
              ),
            );
            setSaved(!saved);
          } catch {
            setMsg(
              "Trình duyệt đang chặn lưu tour. Bạn vẫn có thể chia sẻ liên kết.",
            );
          }
        }}
      >
        {saved ? "♥ Đã lưu tour" : "♡ Lưu tour"}
      </button>
      <button
        onClick={async () => {
          try {
            if (navigator.share)
              await navigator.share({ title, url: location.href });
            else {
              await navigator.clipboard.writeText(location.href);
              setMsg("Đã sao chép liên kết.");
            }
          } catch {
            setMsg(
              "Chưa chia sẻ được. Bạn có thể sao chép URL trên thanh địa chỉ.",
            );
          }
        }}
      >
        Chia sẻ tour ↗
      </button>
      <Link href="/saved">Tour đã lưu</Link>
      <p role="status">{msg}</p>
    </div>
  );
}
export function Gallery({
  pictures,
  poster,
}: {
  pictures: Picture[];
  poster?: Picture | null;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const [index, setIndex] = useState(0);
  const trigger = useRef<HTMLButtonElement | null>(null);
  const all = [...pictures, ...(poster ? [poster] : [])];
  function show(i: number, e: React.MouseEvent<HTMLButtonElement>) {
    trigger.current = e.currentTarget;
    setIndex(i);
    dialog.current?.showModal();
  }
  function close() {
    dialog.current?.close();
    trigger.current?.focus();
  }
  return (
    <>
      <div className="gallery-grid">
        {all.map((p, i) => (
          <button
            key={p.id + "-" + i}
            onClick={(e) => show(i, e)}
            className={p.id === poster?.id ? "poster-thumb" : ""}
          >
            {p.kind === "video" ? (
              <span>Xem video</span>
            ) : (
              <Image
                sizes="(max-width:540px) 90vw, 46vw"
                src={p.url}
                alt={p.alt_text || "Ảnh hành trình"}
                loading="lazy"
                width="640"
                height="440"
              />
            )}
            {p.id === poster?.id && <span>Xem poster đầy đủ ↗</span>}
          </button>
        ))}
      </div>
      <dialog
        ref={dialog}
        className="lightbox"
        onClose={() => trigger.current?.focus()}
        onKeyDown={(e) => {
          if (e.key === "ArrowRight") setIndex((index + 1) % all.length);
          if (e.key === "ArrowLeft")
            setIndex((index + all.length - 1) % all.length);
        }}
      >
        <div className="lightbox-controls">
          <button onClick={close}>Đóng ×</button>
          <button
            onClick={() => setIndex((index + all.length - 1) % all.length)}
          >
            Ảnh trước ←
          </button>
          <button onClick={() => setIndex((index + 1) % all.length)}>
            Ảnh tiếp →
          </button>
          <span>
            {index + 1}/{all.length}
          </span>
        </div>
        {all[index] &&
          (all[index].kind === "video" ? (
            <video controls src={all[index].url} />
          ) : (
            <a href={all[index].url} target="_blank" rel="noreferrer">
              <img
                src={all[index].url}
                alt={all[index].alt_text || "Ảnh phóng lớn"}
              />
            </a>
          ))}
        <p>
          Chạm ảnh để mở kích thước gốc và phóng to. Dùng phím ← → để chuyển
          ảnh.
        </p>
      </dialog>
    </>
  );
}
