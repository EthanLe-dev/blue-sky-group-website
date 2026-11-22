import Lookup from "@/components/Lookup";
export const dynamic = "force-dynamic";
export const metadata = {
  title: "Tra cứu đơn",
  robots: { index: false, follow: false },
};
export default function Page() {
  return (
    <main id="main" className="page wrap prose-page">
      <h1>Tra cứu đơn tour</h1>
      <Lookup
        mock={
          ["local", "test"].includes(process.env.APP_ENV || "local") &&
          process.env.TURNSTILE_MODE !== "real"
        }
        siteKey={process.env.TURNSTILE_SITE_KEY || ""}
      />
    </main>
  );
}
