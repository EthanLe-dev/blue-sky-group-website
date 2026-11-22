"use client";
import { useState } from "react";
export default function Map({
  points,
}: {
  points: { name: string; latitude: number; longitude: number }[];
}) {
  const [open, setOpen] = useState(false);
  const [index, setIndex] = useState(0);
  if (!points.length) return <p>Vị trí bản đồ đang được xác minh.</p>;
  const p = points[index];
  const bbox = [
    p.longitude - 0.03,
    p.latitude - 0.02,
    p.longitude + 0.03,
    p.latitude + 0.02,
  ].join(",");
  return (
    <div>
      <button onClick={() => setOpen(!open)} aria-expanded={open}>
        {open ? "Ẩn bản đồ" : "Mở bản đồ điểm đến"}
      </button>
      {open && (
        <>
          <label>
            Chọn điểm trên bản đồ
            <select
              value={index}
              onChange={(e) => setIndex(Number(e.target.value))}
            >
              {points.map((p, i) => (
                <option key={i} value={i}>
                  {p.name}
                </option>
              ))}
            </select>
          </label>
          <iframe
            className="map-frame"
            title={"Bản đồ " + p.name}
            loading="lazy"
            src={
              "https://www.openstreetmap.org/export/embed.html?bbox=" +
              encodeURIComponent(bbox) +
              "&layer=mapnik&marker=" +
              p.latitude +
              "%2C" +
              p.longitude
            }
            referrerPolicy="no-referrer"
          />
          <p>
            <a
              href={
                "https://www.openstreetmap.org/?mlat=" +
                p.latitude +
                "&mlon=" +
                p.longitude
              }
              target="_blank"
              rel="noreferrer"
            >
              Mở bản đồ lớn
            </a>{" "}
            · ©{" "}
            <a href="https://www.openstreetmap.org/copyright">
              OpenStreetMap contributors
            </a>
          </p>
        </>
      )}
    </div>
  );
}
