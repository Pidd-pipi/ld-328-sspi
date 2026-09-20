import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 18628,
    proxy: { '/api': { target: 'http://localhost:19628', changeOrigin: true } }
  }
});
