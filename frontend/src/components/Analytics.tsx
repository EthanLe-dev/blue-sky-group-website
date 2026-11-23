"use client";

import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { usePathname } from "next/navigation";
import { pageView, setAnalytics, tourEvent, contactEvent, selectTourEvent } from "@/lib/analytics";

const key = "bluesky-analytics-choice";
let current = "";
function subscribe(callback: () => void) {
  window.addEventListener(key, callback);
  window.addEventListener("storage", callback);
  return () => {
    window.removeEventListener(key, callback);
    window.removeEventListener("storage", callback);
  };
}
function choice() {
  if (current) return current;
  try { return localStorage.getItem(key) || ""; } catch { return ""; }
}
function choose(value: string) {
  try {
    localStorage.setItem(key, value);
    current = "";
  } catch { current = value; }
  window.dispatchEvent(new Event(key));
}

export default function Analytics() {
  const [config, setConfig] = useState<{ id: string; debug: boolean }>();
  const value = useSyncExternalStore(subscribe, choice, () => "");
  const [open, setOpen] = useState(false);
  const pathname = usePathname();
  const lastPage = useRef("");
  useEffect(() => {
    const controller = new AbortController();
    fetch("/measurement", { signal: controller.signal, cache: "no-store" })
      .then(response => response.ok ? response.json() : null)
      .then(data => { if (!controller.signal.aborted && data?.id) setConfig(data); })
      .catch(() => {});
    return () => controller.abort();
  }, []);
  useEffect(() => {
    if (!config) return;
    setAnalytics(config.id, value === "yes", config.debug);
    if (value !== "yes") lastPage.current = "";
    else if (lastPage.current !== pathname) {
      lastPage.current = pathname;
      pageView();
      const tour = document.querySelector<HTMLElement>("[data-analytics-tour]")?.dataset.analyticsTour;
      if (tour) tourEvent("view_item", tour);
    }
  }, [config, value, pathname]);
  useEffect(() => {
    if (!config || value !== "yes") return;
    const clicked = (event: MouseEvent) => {
      if (!(event.target instanceof Element)) return;
      const link = event.target.closest<HTMLElement>("a[data-contact-channel], a[data-select-tour]");
      if (!link) return;
      const { contactChannel, contactPosition, contactTour, selectTour, selectPosition } = link.dataset;
      if (contactChannel && contactPosition) contactEvent(contactChannel, contactPosition, contactTour);
      else if (selectTour && selectPosition) selectTourEvent(selectTour, selectPosition);
    };
    document.addEventListener("click", clicked);
    return () => document.removeEventListener("click", clicked);
  }, [config, value]);
  if (!config) return null;
  return (
    <aside className="wrap analytics-preference" aria-label="Tùy chọn phân tích">
      {(!value || open) ? <>
        <p>Cho phép Google Analytics đo lượt xem và thao tác đặt tour? Bạn có thể đổi lựa chọn bất cứ lúc nào.</p>
        <div className="tools">
          <button onClick={() => { choose("yes"); setOpen(false); }}>Cho phép phân tích</button>
          <button onClick={() => { setAnalytics(config.id, false, config.debug); choose("no"); setOpen(false); }}>Từ chối phân tích</button>
        </div>
      </> : <button onClick={() => setOpen(true)}>Tùy chọn phân tích</button>}
    </aside>
  );
}
