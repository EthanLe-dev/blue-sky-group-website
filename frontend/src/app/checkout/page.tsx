import { notFound } from "next/navigation";
import { read, type Tour, type Departure } from "@/lib/catalog";
import Checkout, { ResumeBooking } from "@/components/Checkout";
export const metadata = {
  title: "Đặt tour",
  robots: { index: false, follow: false },
};
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ tour?: string; departure?: string; booking?: string }>;
}) {
  const { tour, departure, booking } = await searchParams;
  const mockPayment = process.env.APP_ENV !== "production" && process.env.PAYMENT_PROVIDER === "mock";
  if (booking) {
    if (!/^[a-f0-9]{8}(?:-[a-f0-9]{4}){3}-[a-f0-9]{12}$/i.test(booking)) notFound();
    return (
      <main id="main" className="page wrap prose-page">
        <h1>Chi tiết đặt tour</h1>
        <ResumeBooking key={booking} id={booking} mockPayment={mockPayment} />
      </main>
    );
  }
  if (!tour || !departure) notFound();
  const t = await read<Tour>("/tours/" + encodeURIComponent(tour));
  const ds = await read<Departure[]>(
    "/tours/" + encodeURIComponent(tour) + "/departures?page_size=100",
  );
  const d = ds.find((v) => v.id === departure);
  if (!d) notFound();
  return (
    <main id="main" className="page wrap prose-page">
      <h1>Đặt tour</h1>
      <Checkout
        departure={departure}
        title={t.title}
        prices={d.prices}
        surcharges={d.surcharges || []}
        mock={
          ["local", "test"].includes(process.env.APP_ENV || "local") &&
          process.env.TURNSTILE_MODE !== "real"
        }
        mockPayment={mockPayment}
        siteKey={process.env.TURNSTILE_SITE_KEY || ""}
      />
    </main>
  );
}
