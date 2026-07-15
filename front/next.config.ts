import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";
import { gatewayUrl as apiProxyTarget } from "./lib/gateway-url";

const withNextIntl = createNextIntlPlugin("./i18n/request.ts");

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${apiProxyTarget}/api/:path*`,
      },
    ];
  },
};

export default withNextIntl(nextConfig);
