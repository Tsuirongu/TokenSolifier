import { defineConfig } from 'vite'
import fs from 'fs'
import path from 'path'

export default defineConfig({
  root: 'src',
  publicDir: '../public',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    assetsDir: 'assets'
  },
  server: {
    port: 34115
  },
  plugins: [
    {
      name: 'html-raw-loader',
      load(id) {
        // 处理以 ?raw 结尾的HTML文件导入
        if (id.endsWith('.html?raw')) {
          const filePath = id.replace('?raw', '')
          try {
            const content = fs.readFileSync(filePath, 'utf-8')
            return `export default ${JSON.stringify(content)}`
          } catch (error) {
            throw new Error(`Failed to load HTML file: ${filePath}`)
          }
        }
      }
    }
  ],
  resolve: {
    alias: {
      '@html': path.resolve(__dirname, 'src/modules')
    }
  }
})
