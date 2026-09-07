#!/usr/bin/env node

import { appendFile, readFile } from "node:fs/promises";
import path from "node:path";

const logDirectory = process.env.CADDYMGM_DEMO_LOG_DIR;
const configPath = process.env.CADDYMGM_DEMO_CONFIG ?? path.resolve("caddy-config/Caddyfile");

if (!logDirectory) {
  console.error("Set CADDYMGM_DEMO_LOG_DIR to the local development access-log directory.");
  process.exit(1);
}

const config = await readFile(configPath, "utf8");
const sites = [...config.matchAll(/^# caddymgm:site\s+(.+)$/gm)]
  .map((match) => match[1].trim())
  .filter(Boolean);

if (!sites.length) {
  console.error(`No managed sites found in ${configPath}.`);
  process.exit(1);
}

const sourceIPs = [
  "8.8.8.8",          // United States
  "1.1.1.1",          // Australia
  "91.198.174.192",   // Netherlands
  "104.28.222.47",    // Singapore
  "1.10.16.1",        // China, included in the configured FireHOL demo feed
  "185.220.101.1",    // Germany
  "147.161.246.112",  // Switzerland
  "178.197.218.96",   // Switzerland
  "45.83.76.194",     // Switzerland
];
const paths = ["/", "/login", "/api/health", "/assets/app.js", "/wp-login.php", "/.env", "/api/v1/status"];
const now = Date.now();
const entriesPerSite = 450;

function record(site, index, siteOffset) {
  const ageMinutes = (entriesPerSite - index) * 96 + (siteOffset * 7);
  const timestamp = (now - ageMinutes * 60_000) / 1000;
  const address = sourceIPs[(index * 3 + siteOffset) % sourceIPs.length];
  const status = index % 19 === 0 ? 503 : index % 5 === 0 ? 403 : index % 11 === 0 ? 404 : 200;
  return JSON.stringify({
    level: "info",
    ts: timestamp,
    logger: "http.log.access.demo",
    msg: "handled request",
    request: {
      remote_ip: address,
      remote_port: String(40000 + ((index * 37) % 20000)),
      client_ip: address,
      proto: "HTTP/2.0",
      method: index % 7 === 0 ? "POST" : "GET",
      host: site,
      uri: paths[(index + siteOffset) % paths.length],
      headers: { "User-Agent": ["CaddyMGM development demo traffic"] },
    },
    duration: 0.003 + ((index % 17) / 1000),
    size: status === 200 ? 1824 : 0,
    status,
  });
}

for (const [siteOffset, site] of sites.entries()) {
  const lines = Array.from({ length: entriesPerSite }, (_, index) => record(site, index, siteOffset));
  await appendFile(path.join(logDirectory, `${site}.access.log`), `${lines.join("\n")}\n`, "utf8");
}

console.log(`Appended ${entriesPerSite * sites.length} synthetic access-log entries for ${sites.length} development sites.`);
