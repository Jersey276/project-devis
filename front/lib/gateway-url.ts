// Server-side gateway target. NODE_ENV drives the default (dev → localhost,
// anything else → the Docker-internal service name); API_PROXY_TARGET lets a
// specific environment (e.g. CI's e2e-integration job, where Next.js runs
// natively on the runner but the gateway is only reachable via its host port
// mapping) override this without touching the dev/prod semantic split.
export const gatewayUrl =
  process.env.API_PROXY_TARGET ??
  (process.env.NODE_ENV === "development"
    ? "http://localhost:8080"
    : "http://devis-gateway:8080");
