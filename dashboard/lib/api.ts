// Talks to tracker.exe. State-changing requests carry X-Arjan,
// which the server requires (cross-site pages can't send it).

export async function api<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method: body === undefined ? "GET" : "POST",
    cache: "no-store",
    headers: {
      "Content-Type": "application/json",
      "X-Arjan": "1",
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await res.text();
  let data: unknown = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    /* not JSON */
  }
  if (!res.ok) {
    const msg =
      (data as { error?: string } | null)?.error ?? (text || res.statusText);
    throw new Error(msg);
  }
  return data as T;
}
