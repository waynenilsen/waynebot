import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 53461,
    proxy: {
      "/api": {
        target: "http://localhost:59731",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://localhost:59731",
        ws: true,
      },
    },
  },
});
