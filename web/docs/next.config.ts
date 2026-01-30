import type { NextConfig } from "next";

const nextConfig: NextConfig = {
    basePath: process.env.DOCS_BASE_PATH || undefined,
    output: "export",
    distDir: "dist",
    typedRoutes: true,
    reactCompiler: true,
};

export default nextConfig;
