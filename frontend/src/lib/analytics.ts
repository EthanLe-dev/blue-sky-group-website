"use client";

type Tag = (...args: unknown[]) => void;
type AnalyticsWindow = Window & { dataLayer?: unknown[]; gtag?: Tag };
let measurement = "";
let enabled = false;

export function analyticsLocation(pathname: string) {
  if (["/", "/tours", "/checkout", "/lookup", "/saved", "/school", "/company", "/contact", "/about", "/blog"].includes(pathname)) return pathname;
  if (pathname.startsWith("/tours/")) return "/tours/detail";
  if (pathname.startsWith("/blog/")) return "/blog/article";
  return "/info";
}

export function setAnalytics(id: string, allowed: boolean, debug: boolean) {
  if (!/^G-[A-Z0-9]+$/.test(id)) return;
  const browser: AnalyticsWindow = window;
  const flags = window as unknown as Record<string, unknown>;
  flags["ga-disable-" + id] = !allowed;
  enabled = allowed;
  if (!allowed || measurement === id) return;
  measurement = id;
  browser.dataLayer ||= [];
  browser.gtag = function () {
    browser.dataLayer?.push(arguments);
  };
  const tag = document.createElement("script");
  tag.async = true;
  tag.referrerPolicy = "no-referrer";
  tag.src = "https://www.googletagmanager.com/gtag/js?id=" + id;
  browser.gtag("js", new Date());
  browser.gtag("config", id, {
    send_page_view: false, allow_google_signals: false, allow_ad_personalization_signals: false,
    page_location: location.origin + analyticsLocation(location.pathname), page_referrer: "",
    ...(debug ? { debug_mode: true } : {}),
  });
  document.head.append(tag);
}

function send(event: string, data: Record<string, unknown>) {
  const browser: AnalyticsWindow = window;
  if (!enabled || !browser.gtag) return;
  browser.gtag("event", event, {
    ...data, send_to: measurement, page_location: location.origin + analyticsLocation(location.pathname),
    page_referrer: "", page_title: "Blue Sky Group",
  });
}

export function pageView() {
  let source = "direct";
  try {
    const referrer = new URL(document.referrer);
    if (referrer.origin === location.origin) source = "internal";
    else if (/(^|\.)google\.(com|com\.vn)$/.test(referrer.hostname)) source = "google";
    else if (/(^|\.)bing\.com$/.test(referrer.hostname)) source = "bing";
    else source = "referral";
  } catch { /* A direct visit has no referrer. */ }
  send("page_view", { traffic_source: source });
}

export function tourEvent(event: "view_item" | "begin_checkout", tour: string) {
  if (!/^[a-f0-9-]{36}$/i.test(tour)) return;
  send(event, { items: [{ item_id: tour }], currency: "VND" });
}

export function leadEvent(groupType?: "school" | "company") {
  send("generate_lead", { lead_type: groupType ?? "callback" });
}

const sent = new Set<string>();
export async function demoPurchase(paymentID: string, amount: string, tour?: string) {
  if (!enabled || !crypto.subtle || !/^\d{1,15}$/.test(amount)) return;
  const hash = await crypto.subtle.digest("SHA-256", new TextEncoder().encode("bluesky-demo:" + paymentID));
  const transaction = "demo-" + Array.from(new Uint8Array(hash), x => x.toString(16).padStart(2, "0")).join("");
  if (!enabled || sent.has(transaction)) return;
  sent.add(transaction);
  try {
    const key = "ga-demo-claims";
    const stored: string[] = JSON.parse(localStorage.getItem(key) || "[]");
    if (Array.isArray(stored) && stored.includes(transaction)) return;
    localStorage.setItem(key, JSON.stringify([...(Array.isArray(stored) ? stored.slice(-99) : []), transaction]));
  } catch { /* The transaction ID also lets GA4 deduplicate across tabs or storage failures. */ }
  send("purchase", {
    transaction_id: transaction, currency: "VND", value: Number(amount), is_demo: true,
    items: tour && /^[a-f0-9-]{36}$/i.test(tour) ? [{ item_id: tour, price: Number(amount), quantity: 1 }] : [],
  });
}
