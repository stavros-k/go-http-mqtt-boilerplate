import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV !== "production";

const nextConfig: NextConfig = {
    basePath: isDev ? "" : "/ui/docs",
    output: "export",
    distDir: "dist",
    typedRoutes: true,
    reactCompiler: true,
};

export default nextConfig;
