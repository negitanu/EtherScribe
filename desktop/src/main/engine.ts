import { ChildProcessWithoutNullStreams, spawn } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { app } from "electron";
import readline from "node:readline";

type Pending = {
  resolve: (v: unknown) => void;
  reject: (e: Error) => void;
};

export class EngineClient {
  private child: ChildProcessWithoutNullStreams | null = null;
  private nextId = 1;
  private pending = new Map<number, Pending>();
  private onProgress: ((p: unknown) => void) | null = null;

  start(): void {
    if (this.child) return;
    const bin = resolveEnginePath();
    this.child = spawn(bin, ["rpc"], { stdio: ["pipe", "pipe", "pipe"] });
    const rl = readline.createInterface({ input: this.child.stdout });
    rl.on("line", (line) => this.onLine(line));
    this.child.stderr.on("data", (chunk) => {
      // Never log config bodies; engine should not emit secrets.
      console.error("[engine]", String(chunk).slice(0, 200));
    });
    this.child.on("error", () => {
      this.child = null;
      for (const p of this.pending.values()) p.reject(new Error("engine could not start"));
      this.pending.clear();
    });
    this.child.on("exit", (code) => {
      this.child = null;
      for (const [, p] of this.pending) {
        p.reject(new Error(`engine exited (${code})`));
      }
      this.pending.clear();
    });
  }

  stop(): void {
    this.child?.kill();
    this.child = null;
  }

  setProgressHandler(handler: ((p: unknown) => void) | null): void {
    this.onProgress = handler;
  }

  async call(method: string, params?: unknown): Promise<unknown> {
    this.start();
    if (!this.child) throw new Error("engine not started");
    const id = this.nextId++;
    const payload = JSON.stringify({ jsonrpc: "2.0", id, method, params: params ?? {} });
    return new Promise((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
      this.child!.stdin.write(payload + "\n");
    });
  }

  private onLine(line: string): void {
    let msg: any;
    try {
      msg = JSON.parse(line);
    } catch {
      return;
    }
    if (msg.method === "progress" && this.onProgress) {
      this.onProgress(msg.params);
      return;
    }
    if (msg.id == null) return;
    const pending = this.pending.get(Number(msg.id));
    if (!pending) return;
    this.pending.delete(Number(msg.id));
    if (msg.error) {
      pending.reject(new Error(msg.error.message || "rpc error"));
      return;
    }
    pending.resolve(msg.result);
  }
}

function resolveEnginePath(): string {
  const name = process.platform === "win32" ? "etherscribe.exe" : "etherscribe";
  const candidates = [
    path.join(process.resourcesPath, "engine", name),
    path.join(app.getAppPath(), "resources", "engine", name),
    path.join(__dirname, "..", "resources", "engine", name),
    path.join(__dirname, "..", "..", "resources", "engine", name),
  ];
  for (const c of candidates) {
    if (fs.existsSync(c)) return c;
  }
  throw new Error("etherscribe engine binary not found; run scripts/build-engine.sh");
}
