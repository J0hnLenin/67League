import type { NextConfig } from "next";

const apiTarget = process.env.API_PROXY_TARGET || "http://localhost:8080";
const adminTarget = process.env.ADMIN_PROXY_TARGET || "http://localhost:8081";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: "/api/v1/:path*",
        destination: `${apiTarget}/api/v1/:path*`,
      },
      {
        source: "/api/admin/:path*",
        destination: `${adminTarget}/admin/:path*`,
      },
    ];
  },
};

export default nextConfig;
