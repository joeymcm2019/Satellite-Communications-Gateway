import type { NextConfig } from "next";

const gatewayUrl =
  process.env.GATEWAY_HTTP_URL ?? "http://localhost:18080";

const config: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/satellites",
        destination: `${gatewayUrl}/api/satellites`,
      },
    ];
  },
};

export default config;