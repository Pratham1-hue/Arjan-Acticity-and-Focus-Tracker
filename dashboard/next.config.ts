import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === "development";

// Production: a static export that tracker.exe serves itself
// (build.ps1 copies dashboard/out -> dist/web).
// Development: `npm run dev` on :4001 proxies /api to the running
// tray app on :4000, so you get hot reload against real data.
const nextConfig: NextConfig = isDev
  ? {
      async rewrites() {
        return [
          { source: "/api/:path*", destination: "http://127.0.0.1:4000/api/:path*" },
        ];
      },
    }
  : { output: "export", images: { unoptimized: true } };

export default nextConfig;
