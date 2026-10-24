import { contextBridge, ipcRenderer } from "electron";

export type JobResult = { cancelled?: boolean; outputPath?: string; format?: string; summary?: Record<string, unknown> };
export type SectionCatalog = { sections: string[]; presets: { key: string; label: string; description: string; sections: string[] }[] };

export type EtherScribeAPI = {
  version: () => Promise<{ version: string }>;
  sections: () => Promise<SectionCatalog>;
  cancel: () => Promise<unknown>;
  openFile: (filters?: { name: string; extensions: string[] }[]) => Promise<string | null>;
  convert: (args: { path: string; format: string; preset: string; sections: string[] }) => Promise<JobResult>;
  compare: (args: { before: string; after: string; format: string }) => Promise<JobResult>;
  previewHtml: (filePath: string) => Promise<string>;
  onEngineInfo: (cb: (info: unknown) => void) => void;
  onEngineError: (cb: (msg: string) => void) => void;
};

const api: EtherScribeAPI = {
  version: () => ipcRenderer.invoke("engine:version"),
  sections: () => ipcRenderer.invoke("engine:sections"),
  cancel: () => ipcRenderer.invoke("engine:cancel"),
  openFile: (filters) => ipcRenderer.invoke("dialog:openFile", filters),
  convert: (args) => ipcRenderer.invoke("convert:run", args),
  compare: (args) => ipcRenderer.invoke("compare:run", args),
  previewHtml: (filePath) => ipcRenderer.invoke("preview:html", filePath),
  onEngineInfo: (cb) => {
    ipcRenderer.on("engine:info", (_e, info) => cb(info));
  },
  onEngineError: (cb) => {
    ipcRenderer.on("engine:error", (_e, msg) => cb(String(msg)));
  },
};

contextBridge.exposeInMainWorld("etherscribe", api);
