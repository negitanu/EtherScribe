import { defineConfig } from "vite";
import path from "node:path";

export default defineConfig({
  root: path.join(__dirname, "src/renderer"),
  base: "./",
  build: {
    outDir: path.join(__dirname, "dist"),
    emptyOutDir: true,
  },
  server: {
    port: 5173,
  },
});
