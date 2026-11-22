import Link from "next/link";
import Image from "next/image";
import Hero from "@/components/Hero";
import PhotoMoment from "@/components/PhotoMoment";
import TourCard from "@/components/TourCard";
import { read, type Tour, type Content } from "@/lib/catalog";
export const dynamic = "force-dynamic";
export default async function Home() {
  let data: {
    featured_tours?: Tour[];
    tours?: Tour[];
    faqs?: Content[];
    banners?: {id:string;title:string;media_id:string;target_url:string|null}[];
  } | null = null;
  try {
    data = await read("/home");
  } catch {}
  let homeCopy:Content|null=null;try{homeCopy=await read<Content>("/pages/home")}catch{}
  const tours = data?.featured_tours || data?.tours || [];
  return (
    <main id="main">
      <Hero />
      <section className="section ground-surface" id="journeys">
        <div className="wrap">
          <div className="section-head">
            <div>
              <p className="eyebrow">ĐI GẦN, KHÁM PHÁ NHIỀU</p>
              <h2>Tour Ninh Chữ & Vĩnh Hy</h2>
            </div>
            <Link className="text-action" href="/tours">
              Xem tất cả tour ↗
            </Link>
          </div>
          {data?.banners?.map(b=><Link key={b.id} href={b.target_url||"/tours"} className="home-banner"><Image src={"/api/v1/media/"+b.media_id} alt={b.title} width={1240} height={320} sizes="100vw"/></Link>)}
          {tours.length ? (
            <div className="tour-grid">
              {tours.slice(0, 3).map((t) => (
                <TourCard key={t.id} tour={t} />
              ))}
            </div>
          ) : (
            <div className="state">
              <h3>Tour đang được cập nhật</h3>
              <p>
                {data
                  ? "Chúng tôi đang chuẩn bị những hành trình mới."
                  : "Chưa kết nối được danh mục. Vui lòng tải lại hoặc xem thông tin liên hệ."}
              </p>
              <Link href="/tours" className="btn">
                Xem danh mục tour
              </Link>
            </div>
          )}
        </div>
      </section>
      <PhotoMoment />
      <section className="section programmes">
        <div className="wrap">
          <div className="section-head">
            <h2>Tour trường học & doanh nghiệp</h2>
          </div>
          <div className="group-programmes grid split">
            <Image
              className="programme-poster"
              src={
                tours.find((t) => t.poster)?.poster?.url || "/brand/logo.jpg"
              }
              alt="Poster chương trình Ninh Chữ 2 ngày 1 đêm do Blue Sky Group cung cấp"
              width={380}
              height={542}
              sizes="(max-width:540px) 90vw, 380px"
            />
            <div className="service-rows">{homeCopy?.body&&<div className="prose" dangerouslySetInnerHTML={{__html:homeCopy.body}}/>}
              <Link href="/schools">
                <span>01 / Trường học</span>
                <h3>Ngoại khoá học đường</h3>
                <p>
                  Tìm hiểu văn hoá Chăm, làng nghề và thiên nhiên qua chuyến
                  tham quan.
                </p>
                <span>Xem chương trình ↗</span>
              </Link>
              <Link href="/companies">
                <span>02 / Doanh nghiệp</span>
                <h3>Du lịch gắn kết đồng đội</h3>
                <p>
                  Xem hành trình dành cho nhóm doanh nghiệp và liên hệ trao đổi
                  chương trình.
                </p>
                <span>Xem chương trình ↗</span>
              </Link>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
