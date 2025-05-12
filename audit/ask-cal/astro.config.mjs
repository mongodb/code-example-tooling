// @ts-check
import { defineConfig } from "astro/config";

import react from "@astrojs/react";
import netlify from "@astrojs/netlify";
// Vite plugins
import { nodePolyfills } from "vite-plugin-node-polyfills";

// https://astro.build/config
export default defineConfig({
  integrations: [react()],
  adapter: netlify(),
  vite: {
    plugins: [
      nodePolyfills({
        exclude: [
          "fs", // Excludes the polyfill for `fs` and `node:fs`.
        ],
        globals: {
          Buffer: true,
          global: true,
          process: true,
        },
        protocolImports: true,
      }),
    ],
  },
});
