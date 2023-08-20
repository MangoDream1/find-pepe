import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import environmentPlugin from "vite-plugin-environment";

const ENV_PREFIX = "REACT_APP_";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    environmentPlugin("all", {
      prefix: ENV_PREFIX,
    }),
  ],
  envPrefix: ENV_PREFIX,
  define: {
    "process.env": {},
  },
});
