export function sanitizeSnippet(raw: string): string {
  return raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.length > 0 && !line.startsWith("#"))
    .map((line) => line.replace(/^\$\s+/, ""))
    .join("\n");
}
