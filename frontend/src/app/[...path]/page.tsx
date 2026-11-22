import Link from "next/link";
import Image from "next/image";
import Map from "@/components/Map";
import { Gallery } from "@/components/TourTools";
import { notFound } from "next/navigation";
import type { Metadata } from "next";
import {
  read,
  DataError,
  type Content,
  type Media,
  type Tour,
} from "@/lib/catalog";
import TourCard from "@/components/TourCard";
import LeadForm from "@/components/LeadForm";
const names: Record<string, string> = {
  about: "Về Blue Sky Group",
  contact: "Liên hệ",
  schools: "Tour dành cho trường học",
  companies: "Tour dành cho doanh nghiệp",
  blog: "Cẩm nang du lịch",
  destinations: "Điểm đến",
  faq: "Câu hỏi thường gặp",
  gallery: "Thư viện ảnh",
  credits: "Nguồn ảnh và quyền sử dụng",
};
async function pageContent(key: string) {
  try {
    return await read<Content>("/pages/" + encodeURIComponent(key));
  } catch (e) {
    if (e instanceof DataError && e.status === 404) notFound();
    throw e;
  }
}
export async function generateMetadata({
  params,
}: {
  params: Promise<{ path: string[] }>;
}): Promise<Metadata> {
  const { path } = await params;
  let title = names[path[0]];
  let content: Content | null = null;
  if (["about", "contact", "schools", "companies"].includes(path[0])) {
    try {
      content = await read<Content>("/pages/" + path[0]);
      title = content.title;
    } catch {}
  }
  if (
    path[0] === "policies" ||
    (["blog", "destinations"].includes(path[0]) && path[1])
  ) {
    const type =
      path[0] === "blog"
        ? "posts"
        : path[0] === "destinations"
          ? "destinations"
          : "pages";
    try {
      const p = await read<Content>(
        "/" + type + "/" + encodeURIComponent(path[1]),
      );
      content = p;
      title = p.seo_title || p.title || p.name;
    } catch {}
  }
  return {
    title: content?.seo_title || title || "Thông tin",
    description:
      content?.seo_description ||
      content?.excerpt ||
      (content?.description || content?.body || "")
        .replace(/<[^>]*>/g, " ")
        .replace(/\s+/g, " ")
        .trim()
        .slice(0, 160) ||
      undefined,
    alternates: { canonical: "/" + path.join("/") },
  };
}
export default async function ContentPage({
  params,
  searchParams,
}: {
  params: Promise<{ path: string[] }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const { path } = await params;
  const key = path[0];
  const filter = await searchParams;
  const query = new URLSearchParams();
  for (const k of ["q", "category", "tag", "kind", "collection", "page"]) {
    const v = filter[k];
    if (typeof v === "string" && v) query.set(k, v);
  }
  if (path.length > 2) notFound();
  let title = names[key];
  let body: React.ReactNode;
  if (key === "credits") {
    body = (
      <>
        <p>
          Logo và poster Ninh Chữ do chủ dự án cung cấp. Không chỉnh sửa nội
          dung ảnh.
        </p>
        <p>
          Ảnh Vịnh Vĩnh Hy và bãi biển Ninh Chữ: Phương Huy, phạm vi công cộng
          theo công bố của tác giả trên Wikimedia Commons.
        </p>
        <ul>
          <li>
            <a href="https://commons.wikimedia.org/wiki/File:V%E1%BB%8Bnh_V%C4%A9nh_Hy.jpg">
              Nguồn ảnh Vịnh Vĩnh Hy
            </a>
          </li>
          <li>
            <a href="https://commons.wikimedia.org/wiki/File:B%C3%A3i_bi%E1%BB%83n_Ninh_Ch%E1%BB%AF.JPG">
              Nguồn ảnh Ninh Chữ
            </a>
          </li>
          <li>
            Video nền: <a href="https://mixkit.co/free-stock-video/bay-landscape-aerial-shot-1940/">Bay landscape, aerial shot</a> từ Mixkit,
            đã cắt ngắn và bỏ âm thanh. <a href="https://mixkit.co/license/modal/videoFree/">Mixkit Stock Video Free License</a>.
          </li>
        </ul>
      </>
    );
  } else if (key === "gallery") {
    const images = await read<Media[]>("/gallery?" + query);
    body = (
      <>
        <form className="filters" action="/gallery">
          <label>
            Loại media
            <select name="kind" defaultValue={query.get("kind") || ""}>
              <option value="">Ảnh và video</option>
              <option value="image">Ảnh</option>
              <option value="video">Video</option>
            </select>
          </label>
          <label>
            Bộ sưu tập
            <input
              name="collection"
              defaultValue={query.get("collection") || ""}
            />
          </label>
          <button className="btn">Lọc thư viện</button>
        </form>
        {images.length ? (
          <Gallery pictures={images} />
        ) : (
          <p>Chưa có media phù hợp bộ lọc.</p>
        )}
        {images.length === 20 && (
          <Link prefetch={false}
            className="text-action"
            href={
              "/gallery?" +
              new URLSearchParams({
                ...Object.fromEntries(query),
                page: String(Number(query.get("page") || 1) + 1),
              })
            }
          >
            Xem trang tiếp
          </Link>
        )}
      </>
    );
  } else if (key === "faq") {
    const list = await read<Content[]>("/faqs");
    body = (
      <div className="content-list">
        {list.length ? (
          list.map((i) => (
            <article key={i.id}>
              <h2>{i.question}</h2>
              <div
                className="prose"
                dangerouslySetInnerHTML={{ __html: i.answer }}
              />
            </article>
          ))
        ) : (
          <p>Chưa có câu hỏi được đăng.</p>
        )}
      </div>
    );
  } else if (key === "blog" || key === "destinations") {
    const type = key === "blog" ? "posts" : "destinations";
    if (path[1]) {
      let item: Content;
      try {
        item = await read("/" + type + "/" + encodeURIComponent(path[1]));
      } catch (e) {
        if (e instanceof DataError && e.status === 404) notFound();
        throw e;
      }
      title = item.title || item.name;
      body = (
        <>
          {item.cover_media_id && (
            <Image
              sizes="100vw"
              loading="eager"
              className="content-image"
              src={"/api/v1/media/" + item.cover_media_id}
              alt={title}
              width="900"
              height="540"
            />
          )}
          <div
            className="prose"
            dangerouslySetInnerHTML={{ __html: item.body || item.description }}
          />
          {key === "blog" && (
            <>
              <p>
                {item.category && (
                  <Link prefetch={false} href={"/blog?category=" + item.category.slug}>
                    {item.category.name}
                  </Link>
                )}
              </p>
              <div className="tools">
                {item.tags?.map((tag) => (
                  <Link prefetch={false} href={"/blog?tag=" + tag.slug} key={tag.id}>
                    #{tag.name}
                  </Link>
                ))}
              </div>
              {(item.related_posts?.length || 0) > 0 && (
                <section className="detail-section">
                  <h2>Bài viết liên quan</h2>
                  <ul>
                    {item.related_posts?.map((post) => (
                      <li key={post.id}>
                        <Link prefetch={false} href={"/blog/" + post.slug}>{post.title}</Link>
                      </li>
                    ))}
                  </ul>
                </section>
              )}
            </>
          )}
          {key === "destinations" &&
            item.latitude != null &&
            item.longitude != null && (
              <Map
                points={[
                  {
                    name: item.name,
                    latitude: item.latitude,
                    longitude: item.longitude,
                  },
                ]}
              />
            )}
          {key === "destinations" && (
            <Link prefetch={false} className="btn" href={"/tours?destination=" + item.slug}>
              Xem tour tại điểm đến
            </Link>
          )}
        </>
      );
    } else {
      const list = await read<Content[]>("/" + type + "?" + query);
      const categories =
        key === "blog" ? await read<Content[]>("/post-categories") : [];
      const tags = key === "blog" ? await read<Content[]>("/tags") : [];
      body = (
        <>
          {key === "blog" && (
            <form className="filters" action="/blog">
              <label>
                Tìm bài viết
                <input
                  name="q"
                  defaultValue={query.get("q") || ""}
                  maxLength={100}
                />
              </label>
              <label>
                Danh mục
                <select
                  name="category"
                  defaultValue={query.get("category") || ""}
                >
                  <option value="">Tất cả danh mục</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.slug}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Thẻ
                <select name="tag" defaultValue={query.get("tag") || ""}>
                  <option value="">Tất cả thẻ</option>
                  {tags.map((c) => (
                    <option key={c.id} value={c.slug}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </label>
              <button className="btn">Lọc bài viết</button>
            </form>
          )}
          <div className="content-list">
            {list.length ? (
              list.map((i) => (
                <article key={i.id}>
                  <h2>
                    <Link prefetch={false} href={"/" + key + "/" + i.slug}>
                      {i.title || i.name} ↗
                    </Link>
                  </h2>
                  <p>{i.excerpt}</p>
                </article>
              ))
            ) : (
              <p>Nội dung đang được cập nhật.</p>
            )}
          </div>
          {list.length === 20 && (
            <Link prefetch={false}
              className="text-action"
              href={
                "/" +
                key +
                "?" +
                new URLSearchParams({
                  ...Object.fromEntries(query),
                  page: String(Number(query.get("page") || 1) + 1),
                })
              }
            >
              Xem trang tiếp
            </Link>
          )}
        </>
      );
    }
  } else if (
    ["about", "contact", "schools", "companies", "policies"].includes(key)
  ) {
    const page = await pageContent(
      key === "policies" ? path[1] || "terms" : key,
    );
    title = page.title;
    const hasLeadForm = ["schools", "companies", "contact"].includes(key);
    const settings = hasLeadForm
      ? await read<{ response_sla?: { message: string } }>("/site-settings")
      : undefined;
    const tourId = typeof filter.tour === "string" &&
      /^[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(filter.tour)
      ? filter.tour : undefined;
    body = (
      <>
        <div
          className="prose"
          dangerouslySetInnerHTML={{ __html: page.body }}
        />
        {["schools", "companies"].includes(key) && (
          <>
            <GroupTours audience={key === "schools" ? "school" : "company"} />
            <Link prefetch={false} className="btn" href="#bao-gia">
              Yêu cầu báo giá
            </Link>
          </>
        )}
        {key === "contact" && <Company />}
        {hasLeadForm && (
          <LeadForm
            groupType={key === "contact" ? undefined : key === "schools" ? "school" : "company"}
            tourId={key === "contact" ? tourId : undefined}
            sla={settings?.response_sla?.message}
            mock={["local", "test"].includes(process.env.APP_ENV || "local") &&
              process.env.TURNSTILE_MODE !== "real"}
            siteKey={process.env.TURNSTILE_SITE_KEY || ""}
          />
        )}
      </>
    );
  } else notFound();
  return (
    <main
      id="main"
      className={
        "page wrap " +
        (!["gallery", "schools", "companies"].includes(key) ? "prose-page" : "")
      }
    >
      <p className="eyebrow">BLUE SKY GROUP</p>
      <h1>{title}</h1>
      {body}
    </main>
  );
}
async function GroupTours({ audience }: { audience: string }) {
  const tours = await read<Tour[]>("/tours?audience=" + audience);
  return (
    <div className="tour-grid">
      {tours.map((t) => (
        <TourCard key={t.id} tour={t} />
      ))}
    </div>
  );
}
async function Company() {
  const company = await read<{
    company_profile: {
      display_name: string;
      address: string | null;
      hotline: string | null;
      email: string | null;
      zalo_url: string | null;
      is_demo: boolean;
      document_media_ids: string[];
      office_coordinates: { latitude: number; longitude: number } | null;
    };
  }>("/site-settings");
  const p = company.company_profile;
  return (
    <section className="detail-section">
      <h2>{p.display_name}</h2>
      {p.is_demo && (
        <p className="notice">
          Thông tin công ty đang ở chế độ demo. Cần xác minh trước khi sử dụng.
        </p>
      )}
      <p>{p.address || "Địa chỉ đang được cập nhật."}</p>
      {p.hotline && (
        <p>
          <a href={"tel:" + p.hotline.replace(/[^+0-9]/g, "")}>
            Gọi {p.hotline}
          </a>
        </p>
      )}
      {p.email && (
        <p>
          <a href={"mailto:" + p.email}>{p.email}</a>
        </p>
      )}
      {p.zalo_url && (
        <p>
          <a href={p.zalo_url}>Liên hệ qua Zalo</a>
        </p>
      )}
      {p.document_media_ids?.length > 0 && (
        <ul>
          {p.document_media_ids.map((id) => (
            <li key={id}>
              <a href={"/api/v1/media/" + id} target="_blank" rel="noreferrer">
                Xem tài liệu công ty
              </a>
            </li>
          ))}
        </ul>
      )}
      {p.office_coordinates && (
        <div>
          {p.is_demo && <p>Vị trí minh họa. Vui lòng xác nhận địa chỉ trước khi đến.</p>}
          <Map points={[{ name: p.is_demo ? "Vị trí minh họa" : p.display_name, ...p.office_coordinates }]} />
        </div>
      )}
      {p.office_coordinates && (
        <a
          target="_blank"
          rel="noreferrer"
          href={
            "https://www.openstreetmap.org/?mlat=" +
            p.office_coordinates.latitude +
            "&mlon=" +
            p.office_coordinates.longitude
          }
        >
          {p.is_demo ? "Xem vị trí minh họa trên OpenStreetMap" : "Xem văn phòng trên OpenStreetMap"} ↗
        </a>
      )}
    </section>
  );
}
