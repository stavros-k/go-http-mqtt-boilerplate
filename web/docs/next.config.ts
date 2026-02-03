import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV !== "production";

const nextConfig: NextConfig = {
    basePath: isDev ? "" : "/docs",
    output: "export",
    distDir: "dist",
    typedRoutes: true,
    reactCompiler: true,
};

export default nextConfig;
