import "server-only";
export type Media = { id: string; url: string; alt_text: string; kind: string };
export type Tour = {
  seo_title: string | null;
  seo_description: string | null;
  id: string;
  slug: string;
  title: string;
  summary: string;
  description: string;
  duration_days: number;
  duration_nights: number;
  origin_name?: string | null;
  price_from: string | null;
  cover: Media | null;
  poster: Media | null;
  is_demo: boolean;
  included: string;
  excluded: string;
  cancellation_policy: string;
  days: { id: string; day_number: number; title: string; body: string }[];
  gallery: Media[];
  stops: {
    id: string;
    name: string;
    latitude: number | null;
    longitude: number | null;
  }[];
  faqs: { id: string; question: string; answer: string }[];
};
export type Content = {
  seo_title?: string | null;
  seo_description?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  id: string;
  slug: string;
  title: string;
  name: string;
  body: string;
  description: string;
  excerpt: string;
  question: string;
  answer: string;
  cover_media_id: string | null;
};
export type Surcharge = {
  code: string;
  label: string;
  amount: string;
  unit: string;
  required: boolean;
};
export type Departure = {
  surcharges: Surcharge[];
  id: string;
  starts_at: string;
  ends_at: string;
  remaining_seats: number;
  status: string;
  deposit_bps: number;
  price_version: number;
  prices: {
    passenger_type: string;
    unit_price: string;
    min_age: number;
    max_age_exclusive: number | null;
    seats_per_person: number;
  }[];
};
export class DataError extends Error {
  status: number;
  constructor(status: number) {
    super("Không tải được dữ liệu.");
    this.status = status;
  }
}
export async function read<T>(path: string): Promise<T> {
  const response = await fetch(
    (process.env.API_UPSTREAM_URL || "http://localhost:8080") +
      "/api/v1" +
      path,
    { cache: "no-store", signal: AbortSignal.timeout(10000) },
  );
  if (!response.ok) throw new DataError(response.status);
  return (await response.json()).data as T;
}
export const money = (v: string | number | null | undefined) =>
  v == null
    ? "Liên hệ"
    : new Intl.NumberFormat("vi-VN", {
        style: "currency",
        currency: "VND",
        maximumFractionDigits: 0,
      }).format(Number(v));
export const date = (v: string) =>
  new Intl.DateTimeFormat("vi-VN", {
    dateStyle: "medium",
    timeZone: "Asia/Ho_Chi_Minh",
  }).format(new Date(v));
export const site = process.env.SITE_URL || "http://localhost:3000";
