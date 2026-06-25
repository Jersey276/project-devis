import { fetchWithRefresh, readUserModeCookie } from "./api";

export async function downloadBlob(
  path: string,
  fallbackFilename: string,
  accept = "application/pdf",
): Promise<void> {
  const clientMode = readUserModeCookie() === "customer";
  const res = await fetchWithRefresh(() =>
    fetch(path, {
      credentials: "include",
      headers: {
        Accept: accept,
        ...(clientMode ? { "X-Client-Mode": "customer" } : {}),
      },
    }),
  );
  if (!res) return;
  if (!res.ok) throw new Error(`HTTP ${res.status}`);

  const blob = await res.blob();
  const cd = res.headers.get("Content-Disposition") ?? "";
  const filename = parseContentDispositionFilename(cd) ?? fallbackFilename;

  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  // Defer the revoke: Safari/older Firefox start the download asynchronously
  // and a same-tick revoke can cancel it.
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function parseContentDispositionFilename(header: string): string | null {
  const ext = /filename\*\s*=\s*UTF-8''([^;]+)/i.exec(header);
  if (ext) {
    try {
      return decodeURIComponent(ext[1].trim());
    } catch {
      // fall through to ascii filename
    }
  }
  const ascii = /filename\s*=\s*"?([^";]+)"?/i.exec(header);
  return ascii ? ascii[1].trim() : null;
}
