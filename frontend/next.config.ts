import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  htmlLimitedBots: /.*/,
  images: {
    localPatterns: [
      { pathname: "/brand/**", search: "" },
      { pathname: "/images/**", search: "" },
      { pathname: "/api/v1/media/*", search: "" },
    ],
  },
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
