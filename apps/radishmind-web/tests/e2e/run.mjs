import { spawn } from "node:child_process";
import { createWriteStream } from "node:fs";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { setTimeout as delay } from "node:timers/promises";

if (process.platform === "win32") {
  throw new Error("Browser regression service cleanup requires POSIX process groups; use Linux, macOS, or WSL.");
}

const webRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const repoRoot = resolve(webRoot, "../..");
const artifactRoot = join(repoRoot, "output/playwright/workflow-e2e");
await mkdir(artifactRoot, { recursive: true });
const output = await mkdtemp(join(artifactRoot, "run-"));
const runtime = await mkdtemp(join(tmpdir(), "radishmind-workflow-e2e-"));
const configPath = join(runtime, "platform.json");

// Pass toolchain settings, not inherited application settings, tokens, or database URLs.
const environmentKeys = new Set([
  "PATH", "HOME", "USER", "SHELL", "TMPDIR", "LANG", "LC_ALL", "TERM", "CI",
  "GOCACHE", "GOMODCACHE", "GOPATH", "GOROOT", "GOENV", "GOTOOLCHAIN",
  "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY",
  "http_proxy", "https_proxy", "all_proxy", "no_proxy",
  "NODE_EXTRA_CA_CERTS", "SSL_CERT_FILE", "PLAYWRIGHT_BROWSERS_PATH",
]);
const environment = Object.fromEntries(Object.entries(process.env).filter(([key]) => environmentKeys.has(key)));
environment.NO_PROXY = environment.no_proxy = "127.0.0.1,localhost";
const groups = [];
const logs = [];
let interrupted = false;
let shutdown;

function groupExists(pid) {
  try {
    process.kill(-pid, 0);
    return true;
  } catch (error) {
    if (error.code === "ESRCH") return false;
    throw error;
  }
}

function signalGroup(pid, signal) {
  try {
    process.kill(-pid, signal);
  } catch (error) {
    if (error.code !== "ESRCH") throw error;
  }
}

function stop() {
  shutdown ??= (async () => {
    for (const child of groups) signalGroup(child.pid, "SIGTERM");
    const deadline = Date.now() + 8_000;
    while (groups.some((child) => groupExists(child.pid)) && Date.now() < deadline) await delay(100);
    for (const child of groups) {
      if (groupExists(child.pid)) signalGroup(child.pid, "SIGKILL");
      await child.finished;
    }
    try {
      for (const log of logs) {
        if (!log.destroyed) await new Promise((accept, reject) => {
          log.once("error", reject);
          log.end(accept);
        });
      }
    } finally {
      await rm(runtime, { recursive: true, force: true });
    }
    console.log(`[workflow-e2e] Stopped owned process groups; removed temporary runtime ${runtime}`);
  })();
  return shutdown;
}

function start(command, args, options = {}) {
  const child = spawn(command, args, { cwd: webRoot, env: environment, detached: true, ...options });
  child.finished = new Promise((accept, reject) => {
    child.once("error", reject);
    child.once("close", (code, signal) => accept({ code, signal }));
  });
  // The startup/readiness path awaits this rejection; avoid an unhandled error while wiring output.
  child.finished.catch(() => {});
  if (child.pid) groups.push(child);
  return child;
}

for (const signal of ["SIGINT", "SIGTERM"]) {
  process.once(signal, () => {
    interrupted = true;
    void stop().then(() => process.exit(signal === "SIGINT" ? 130 : 143), (error) => {
      console.error(error);
      process.exit(1);
    });
  });
}

console.log(`[workflow-e2e] Artifacts: ${output}`);
let exitCode = 1;
try {
  await writeFile(configPath, "{}\n", { mode: 0o600 });
  if (interrupted) throw new Error("Interrupted during configuration setup.");
  const launcher = start("bash", [
    join(repoRoot, "scripts/run-radishmind-web-dev.sh"),
    "--mode", "dev-live", "--workflow-definition-local-product", "--no-reuse-existing",
    "--frontend-url", "http://127.0.0.1:4100", "--backend-url", "http://127.0.0.1:17000",
    "--timeout-seconds", "120", "--log-dir", join(output, "services"),
    "--frontend-config", join(webRoot, "tests/e2e/vite.config.ts"),
  ], {
    cwd: repoRoot,
    env: {
      ...environment,
      RADISHMIND_PLATFORM_CONFIG: configPath,
      RADISHMIND_PLATFORM_PROVIDER: "mock",
      RADISHMIND_SQLITE_DEV_DATABASE_PATH: join(runtime, "workflow.db"),
    },
    stdio: ["ignore", "pipe", "pipe"],
  });
  const launcherLog = createWriteStream(join(output, "launcher.log"));
  logs.push(launcherLog);
  const loggingFailure = new Promise((_, reject) => launcherLog.once("error", reject));
  loggingFailure.catch(() => {});
  launcher.stdout.pipe(launcherLog, { end: false });
  launcher.stderr.pipe(launcherLog, { end: false });
  let ready;
  const readiness = new Promise((accept) => { ready = accept; });
  let partialLine = "";
  launcher.stdout.on("data", (chunk) => {
    partialLine += chunk.toString();
    if (partialLine.includes("RadishMind web is ready: http://127.0.0.1:4100")) ready();
    partialLine = partialLine.slice(-2048);
  });
  const startupTimer = new AbortController();
  try {
    await Promise.race([
      readiness,
      loggingFailure,
      launcher.finished.then(({ code, signal }) => { throw new Error(`Service launcher exited before readiness (${code ?? signal}). See ${output}/services.`); }),
      delay(150_000, null, { signal: startupTimer.signal }).then(() => { throw new Error("Isolated service startup timed out."); }),
    ]);
  } finally {
    startupTimer.abort();
  }
  if (interrupted) throw new Error("Interrupted during startup.");
  console.log("[workflow-e2e] SQLite/mock services ready on 4100 and 17000.");
  const tests = start(process.execPath, [
    join(webRoot, "node_modules/@playwright/test/cli.js"), "test",
    "--config", "tests/e2e/playwright.config.ts", ...process.argv.slice(2),
  ], {
    env: { ...environment, RADISHMIND_E2E_WEB_URL: "http://127.0.0.1:4100", RADISHMIND_E2E_OUTPUT_DIR: output },
    stdio: "inherit",
  });
  const result = await Promise.race([
    tests.finished,
    loggingFailure,
    launcher.finished.then(() => { throw new Error("Services exited while browser tests were running."); }),
  ]);
  exitCode = result.code ?? 1;
} catch (error) {
  console.error(`[workflow-e2e] ${error.message}`);
} finally {
  await stop();
}
process.exitCode = exitCode;
