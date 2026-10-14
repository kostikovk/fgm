import type { Metadata, Viewport } from "next";
import type { ReactNode } from "react";
import "./globals.css";

export const viewport: Viewport = {
  themeColor: "#080c14",
  width: "device-width",
  initialScale: 1,
};

export const metadata: Metadata = {
  metadataBase: new URL("https://kostikovk.github.io/fgm/"),
  title:
    "FGM (Fast Go Manager) — Project Go & golangci-lint Toolchains From One CLI",
  description:
    "FGM manages your project's Go and compatible golangci-lint toolchains from one CLI. Read go.work and go.mod, select a compatible linter, and run checks with fgm exec.",
  alternates: {
    canonical: "https://kostikovk.github.io/fgm/",
  },
  icons: {
    icon: "/fgm/assets/favicon.svg",
  },
  openGraph: {
    type: "website",
    title: "FGM — Your project's Go and golangci-lint, managed together.",
    description:
      "FGM reads your Go project files, helps select a compatible golangci-lint version, and runs checks with the selected toolchains from one CLI.",
    url: "https://kostikovk.github.io/fgm/",
    images: [
      {
        url: "https://kostikovk.github.io/fgm/assets/social.png",
        width: 1200,
        height: 630,
        alt: "FGM — Project Go and golangci-lint managed from one CLI",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "FGM — Fast Go Manager",
    description:
      "Manage project Go versions and compatible golangci-lint toolchains from one CLI.",
    images: ["https://kostikovk.github.io/fgm/assets/social.png"],
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
