import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react-swc';
import { resolve } from 'path';
import { defineConfig } from 'vite';
import { VitePWA } from 'vite-plugin-pwa';

// https://vite.dev/config/
export default defineConfig(({ command }) => ({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'mask-icon.svg'],
      manifest: {
        name: 'ScienceOL',
        short_name: 'ScienceOL',
        description: 'ScienceOL Application',
        display: 'standalone',
        display_override: ['window-controls-overlay'],
        icons: [
          {
            src: 'pwa-icon.png',
            sizes: '192x192 512x512',
            type: 'image/svg+xml',
          },
        ],
      },
      devOptions: {
        enabled: true,
      },
      workbox: {
        maximumFileSizeToCacheInBytes: 10 * 1024 * 1024, // 10 MiB
        // 排除 /api 和 /xyzen 开头的请求，避免 Service Worker 拦截后端接口
        navigateFallbackDenylist: [/^\/api/, /^\/xyzen/],
      },
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
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
      host: 'localhost',
      port: 32234,
    },
  },
  // 仅在生产构建时移除 console.log / console.debug（保留 warn / error）
  esbuild:
    command === 'build'
      ? { pure: ['console.log', 'console.debug'] }
      : undefined,
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'three-vendor': ['three', '@react-three/fiber', '@react-three/drei'],
          'monaco-vendor': ['monaco-editor', '@monaco-editor/react'],
          'ui-vendor': [
            '@headlessui/react',
            '@radix-ui/react-slot',
            'framer-motion',
            'motion',
          ],
        },
      },
    },
  },
}));
