"use client";
import { useEffect, useRef, useState } from "react";
import type { Map as LeafletMap } from "leaflet";
import "leaflet/dist/leaflet.css";
type Point = { name: string; latitude: number; longitude: number };
export default function Map({ points }: { points: Point[] }) {
  const [open, setOpen] = useState(false);
  const [index, setIndex] = useState(0);
  const [error, setError] = useState("");
  const host = useRef<HTMLDivElement>(null);
  const map = useRef<LeafletMap | null>(null);
  useEffect(() => {
    if (!open || !points.length || !host.current) return;
    let active = true;
    let instance: LeafletMap | null = null;
    import("leaflet")
      .then((L) => {
        if (!active || !host.current) return;
        instance = L.map(host.current, {
          scrollWheelZoom: false,
          zoomAnimation: false,
          fadeAnimation: false,
          markerZoomAnimation: false,
          inertia: false,
          zoomControl: false,
        });
        map.current = instance;
        L.control
          .zoom({
            zoomInTitle: "Phóng to bản đồ",
            zoomOutTitle: "Thu nhỏ bản đồ",
          })
          .addTo(instance);
        L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
          maxZoom: 19,
          attribution:
            '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap contributors</a>',
        })
          .on("tileerror", () => {
            if (active)
              setError(
                "Một số ô bản đồ chưa tải được. Bạn có thể mở bản đồ lớn bên dưới.",
              );
          })
          .addTo(instance);
        const icon = L.icon({
          iconUrl: "/map/marker-icon.png",
          iconRetinaUrl: "/map/marker-icon-2x.png",
          shadowUrl: "/map/marker-shadow.png",
          iconSize: [25, 41],
          iconAnchor: [12, 41],
          popupAnchor: [1, -34],
          shadowSize: [41, 41],
        });
        points.forEach((p) => {
          const label = document.createElement("span");
          label.textContent = p.name;
          L.marker([p.latitude, p.longitude], {
            icon,
            title: p.name,
            alt: p.name,
          })
            .bindPopup(label)
            .addTo(instance!);
        });
        instance.fitBounds(
          L.latLngBounds(
            points.map((p) => [p.latitude, p.longitude] as [number, number]),
          ),
          { padding: [32, 32], maxZoom: 13, animate: false },
        );
      })
      .catch(() => {
        if (active)
          setError("Chưa tải được bản đồ. Hãy dùng liên kết bản đồ lớn.");
      });
    return () => {
      active = false;
      instance?.remove();
      map.current = null;
    };
  }, [open, points]);
  if (!points.length) return <p>Vị trí bản đồ đang được xác minh.</p>;
  const p = points[index] || points[0];
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
              onChange={(e) => {
                const n = Number(e.target.value);
                setIndex(n);
                map.current?.setView(
                  [points[n].latitude, points[n].longitude],
                  13,
                  { animate: false },
                );
              }}
            >
              {points.map((p, i) => (
                <option key={i} value={i}>
                  {p.name}
                </option>
              ))}
            </select>
          </label>
          <div
            ref={host}
            className="map-frame"
            role="region"
            aria-label="Bản đồ điểm đến"
          />
          <p role="status">{error}</p>
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
