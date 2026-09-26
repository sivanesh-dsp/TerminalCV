import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// `base` is configurable so the same build works on Vercel (root "/")
// and GitHub Pages project sites ("/<repo-name>/").
// Set VITE_BASE when deploying to a subpath, e.g. VITE_BASE=/terminal-resume/
export default defineConfig({
  base: process.env.VITE_BASE || '/',
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@shared': fileURLToPath(new URL('./shared', import.meta.url)),
    },
  },
  server: {
    // In development the site hydrates from the résumé API running separately
    // (see ../resume-api: `make dev`). In production Caddy proxies /api/* to the
    // same container network, so the site always calls a same-origin URL.
    proxy: {
      '/api': {
        target: process.env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8000',
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'es2020',
    cssMinify: true,
    rollupOptions: {
      output: {
        // Split heavier animation libs so the first paint stays light.
        manualChunks: {
          motion: ['framer-motion'],
        },
      },
    },
  },
});
