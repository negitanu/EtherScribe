import { app, BrowserWindow, dialog, ipcMain } from "electron";
import path from "node:path";
import fs from "node:fs/promises";
import { EngineClient } from "./engine";

const engine = new EngineClient();
let mainWindow: BrowserWindow | null = null;
const previewPaths = new Set<string>();

function createWindow(): void {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 900,
    minWidth: 760,
    minHeight: 620,
    backgroundColor: "#f4f6fa",
    webPreferences: {
      preload: path.join(__dirname, "..", "preload", "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });

  const devUrl = process.env.VITE_DEV_SERVER_URL;
  if (devUrl) {
    mainWindow.loadURL(devUrl);
  } else {
    mainWindow.loadFile(path.join(__dirname, "..", "..", "dist", "index.html"));
  }
}

app.whenReady().then(async () => {
  createWindow();
  try {
    engine.start();
    const info = await engine.call("engine.version");
    mainWindow?.webContents.send("engine:info", info);
  } catch (err) {
    console.error("engine handshake failed", err);
    mainWindow?.webContents.send("engine:error", String(err));
  }
});

app.on("activate", () => {
  if (BrowserWindow.getAllWindows().length === 0) createWindow();
});

app.on("window-all-closed", () => {
  engine.stop();
  if (process.platform !== "darwin") app.quit();
});

ipcMain.handle("engine:version", async () => engine.call("engine.version"));
ipcMain.handle("engine:sections", async () => engine.call("sections.list"));
ipcMain.handle("engine:cancel", async () => engine.call("cancel"));

ipcMain.handle("dialog:openFile", async (_e, filters?: Electron.FileFilter[]) => {
  const result = await dialog.showOpenDialog(mainWindow!, {
    properties: ["openFile"],
    filters: filters ?? [{ name: "Config", extensions: ["conf", "xml", "set", "txt"] }],
  });
  if (result.canceled || result.filePaths.length === 0) return null;
  return result.filePaths[0];
});

ipcMain.handle("dialog:saveFile", async (_e, defaultPath: string) => {
  const result = await dialog.showSaveDialog(mainWindow!, { defaultPath });
  if (result.canceled || !result.filePath) return null;
  return result.filePath;
});

ipcMain.handle(
  "convert:run",
  async (_e, args: { path: string; format: string; preset: string; sections: string[] }) => {
    const suggested =
      args.format === "excel" ? "sheet.xlsx" : args.format === "html" ? "sheet.html" : "sheet.json";
    const output = await dialog.showSaveDialog(mainWindow!, { defaultPath: suggested });
    if (output.canceled || !output.filePath) return { cancelled: true };
    const result = (await engine.call("convert.run", {
      path: args.path,
      format: args.format,
      preset: args.preset,
      sections: args.sections,
      output: output.filePath,
    })) as { outputPath: string; summary: unknown; format: string };
    if (result.format === "html") previewPaths.add(result.outputPath);
    return result;
  },
);

ipcMain.handle(
  "compare:run",
  async (_e, args: { before: string; after: string; format: string }) => {
    const output = await dialog.showSaveDialog(mainWindow!, {
      defaultPath: args.format === "html" ? "diff.html" : "diff.json",
    });
    if (output.canceled || !output.filePath) return { cancelled: true };
    const result = await engine.call("compare.run", {
      before: args.before,
      after: args.after,
      format: args.format,
      output: output.filePath,
    }) as { outputPath: string; format: string };
    if (result.format === "html") previewPaths.add(result.outputPath);
    return result;
  },
);

ipcMain.handle("preview:html", async (_e, filePath: string) => {
  if (!previewPaths.has(filePath)) throw new Error("Preview unavailable");
  const data = await fs.readFile(filePath, "utf8");
  return data;
});
