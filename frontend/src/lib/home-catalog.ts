import type { Content, Tour } from "./catalog";

export type HomeData = {
  featured_tours?: Tour[];
  destinations?: Content[];
  latest_posts?: Content[];
  testimonials?: { id: string; display_name: string; body: string; rating: number; is_demo: boolean }[];
  banners?: { id: string; title: string; media_id: string; target_url: string | null }[];
};

export function homeCatalog(data: HomeData | null) {
  const tours = data?.featured_tours?.slice(0, 6) ?? [];
  return {
    state: data === null ? "error" : tours.length ? "ready" : "empty",
    tours,
    destinations: data?.destinations?.slice(0, 6) ?? [],
    banners: data?.banners?.slice(0, 6) ?? [],
    posts: data?.latest_posts ?? [],
    reviews: data?.testimonials ?? [],
  };
}
