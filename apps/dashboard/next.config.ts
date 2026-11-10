import type { NextConfig } from "next";
const config: NextConfig = {
  output: "standalone",
    async rewrites() {
    return [
      {
        source: "/api/satellites",
        destination: "http://localhost:18080/api/satellites",
      },
    ];
  },
};
export default config;
