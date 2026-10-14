import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  basePath: "/fgm",
  typedRoutes: true,
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
