import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import fs from 'fs';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    https: {
      key: fs.readFileSync(path.resolve(__dirname, '../certs/privkey.pem')),
      cert: fs.readFileSync(path.resolve(__dirname, '../certs/fullchain.pem')),
    },
    hmr: {
      host: 'edge.2gc.ru',
      port: 5173,
      protocol: 'wss'
    }
  },
}); 