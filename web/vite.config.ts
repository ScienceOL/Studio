import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import { resolve } from "path";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": resolve(__dirname, "./src"),
    },
  },
  server: {
    host: true, // 监听所有地址
    port: 32234,
    strictPort: true,
    watch: {
      usePolling: true, // Docker 环境下必须启用轮询
      interval: 100, // 轮询间隔（毫秒）
    },
    hmr: {
      // 热模块替换配置
      host: "localhost",
      port: 32234,
    },
  },
});
