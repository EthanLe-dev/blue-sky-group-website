import type { MetadataRoute } from "next";
export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: process.env.APP_ENV === "production" ? "/" : undefined,
      disallow:
        process.env.APP_ENV === "production"
          ? ["/checkout", "/lookup", "/saved", "/api/"]
          : "/",
    },
    sitemap: (process.env.SITE_URL || "http://localhost:3000") + "/sitemap.xml",
  };
}
