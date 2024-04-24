import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "http://go_container:8080", // goのコンテナ名を指定する
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
