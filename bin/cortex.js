#!/usr/bin/env node

import { spawn, execSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, "..");

const isWindows = process.platform === "win32";
const binaryName = isWindows ? "cortex.exe" : "cortex";
const binaryPath = path.join(__dirname, binaryName);

function ensureBinary() {
  if (fs.existsSync(binaryPath)) {
    return binaryPath;
  }

  // Attempt local build if Go is available
  try {
    const goCheck = execSync("go version", { stdio: "ignore" });
    process.stderr.write("Building cortex binary from source...\n");
    execSync(`go build -o "${binaryPath}" ./cmd/cortex`, {
      cwd: rootDir,
      stdio: "inherit",
    });
    if (fs.existsSync(binaryPath)) {
      return binaryPath;
    }
  } catch {
    // Go compiler not available
  }

  throw new Error(
    `Cortex binary not found at ${binaryPath}. Please run 'npm run build' or install Go 1.23+ to compile.`
  );
}

function main() {
  const binary = ensureBinary();
  const child = spawn(binary, process.argv.slice(2), {
    stdio: "inherit",
  });

  child.on("error", (err) => {
    console.error("Failed to execute cortex:", err);
    process.exit(1);
  });

  child.on("exit", (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal);
    } else {
      process.exit(code ?? 0);
    }
  });
}

main();
