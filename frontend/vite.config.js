import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// in dev, /api/* is proxied to the Go API on localhost:8080;
// in the docker image, nginx does the same proxying to the app container
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api/, ''),
      },
    },
  },
})
