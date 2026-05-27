#!/usr/bin/env node

const https = require("https");
const fs = require("fs");
const path = require("path");

const VERSION = require("./package.json").version;

const PLATFORM_MAP = {
  "darwin-arm64": "val-darwin-arm64",
  "darwin-x64": "val-darwin-amd64",
  "linux-x64": "val-linux-amd64",
  "win32-x64": "val-windows-amd64.exe",
};

const key = `${process.platform}-${process.arch}`;
const binaryName = PLATFORM_MAP[key];

if (!binaryName) {
  console.error(`val: unsupported platform ${key}`);
  process.exit(1);
}

const url = `https://github.com/tq303/valet/releases/download/v${VERSION}/${binaryName}`;
const dest = path.join(__dirname, "bin", "val" + (process.platform === "win32" ? ".exe" : ""));

fs.mkdirSync(path.join(__dirname, "bin"), { recursive: true });

console.log(`val: downloading binary for ${key}...`);

function download(url, dest, cb) {
  const file = fs.createWriteStream(dest);
  https.get(url, (res) => {
    if (res.statusCode === 301 || res.statusCode === 302) {
      return download(res.headers.location, dest, cb);
    }
    if (res.statusCode !== 200) {
      cb(new Error(`download failed: ${res.statusCode}`));
      return;
    }
    res.pipe(file);
    file.on("finish", () => file.close(cb));
  }).on("error", cb);
}

download(url, dest, (err) => {
  if (err) {
    console.error(`val: failed to download binary: ${err.message}`);
    process.exit(1);
  }
  fs.chmodSync(dest, 0o755);
  console.log("val: ready");
});
