export const dynamic = "force-dynamic";

export function GET() {
  const id = process.env.GA4_MEASUREMENT_ID || "";
  return Response.json({
    id: /^G-[A-Z0-9]+$/.test(id) ? id : "",
    debug: process.env.APP_ENV !== "production" && process.env.GA4_DEBUG === "true",
  }, { headers: { "Cache-Control": "no-store" } });
}
