import type { Metadata } from "next";
import "./globals.css";

// The tray finds an open dashboard window by this exact title
// (dashboardTitle in cmd/daemon/tray.go) — keep them in sync.
export const metadata: Metadata = {
  title: "Arjan Dashboard",
  description: "Where your time went — local study & laptop usage tracker",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en" className="h-full">
      <body className="h-full bg-zinc-950 text-zinc-100 antialiased">
        {children}
      </body>
    </html>
  );
}
