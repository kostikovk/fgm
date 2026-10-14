import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const outDir = path.resolve(__dirname, "..", "out");
const port = Number(process.env.PORT || 4173);

const MIME_TYPES = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
  ".ico": "image/x-icon",
  ".txt": "text/plain; charset=utf-8",
  ".woff2": "font/woff2",
};

const server = http.createServer((req, res) => {
  const rawUrl = (req.url || "/").split("?")[0].split("#")[0];
  if (!rawUrl.startsWith("/fgm")) {
    res.writeHead(404, { "Content-Type": "text/plain" });
    res.end("Not Found (expected /fgm prefix)");
    return;
  }

  let subPath = rawUrl.slice("/fgm".length);
  if (subPath === "" || subPath === "/") {
    subPath = "/index.html";
  }

  const safePath = path.normalize(subPath).replace(/^(\.\.(\/|\\|$))+/, "");
  let filePath = path.join(outDir, safePath);

  if (fs.existsSync(filePath) && fs.statSync(filePath).isDirectory()) {
    filePath = path.join(filePath, "index.html");
  }

  if (!fs.existsSync(filePath) || !fs.statSync(filePath).isFile()) {
    res.writeHead(404, { "Content-Type": "text/plain" });
    res.end(`Not Found: ${rawUrl}`);
    return;
  }

  const ext = path.extname(filePath).toLowerCase();
  const contentType = MIME_TYPES[ext] || "application/octet-stream";
  res.writeHead(200, { "Content-Type": contentType });
  fs.createReadStream(filePath).pipe(res);
});

server.listen(port, "127.0.0.1", () => {
  console.log(`Static export server listening on http://127.0.0.1:${port}/fgm/`);
});
