# Arjan Dashboard

The dashboard UI for Arjan. It is a **static** Next.js export
(`output: "export"`): `build.ps1` builds it and copies `out\` to
`dist\web\`, where `tracker.exe` serves it at
http://127.0.0.1:4000 together with the JSON API (`/api/*`,
implemented in Go — see `cmd/daemon/server.go`).

Stack: Next.js 16 · React 19 · Tailwind CSS 4 · Recharts 3.

## Develop with hot reload

Keep Arjan running in the tray, then:

```powershell
npm run dev    # http://localhost:4001
```

In dev mode, `/api/*` is proxied to the tray app on port 4000
(see `next.config.ts`), so you work against your real data.

## Layout

- `app/page.tsx` — the whole dashboard page
- `components/` — cards: Subjects ("What you learned"), TopList
  (apps/websites), Activities, Timeline, Review, SettingsPanel, charts
- `lib/types.ts` — API types (mirror the Go structs in `analyze.go`)
- `lib/api.ts` — fetch helper (adds the `X-Arjan` header the server requires)

The page `<title>` must stay **"Arjan Dashboard"** — the tray uses it
to find and focus an already-open dashboard window.
