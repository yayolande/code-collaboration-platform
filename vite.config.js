import { defineConfig } from 'vite'

export default defineConfig({
  root: "./src",
  publicDir: "public",
  build: {
    copyPublicDir: true,
    outDir: "../dist",
    watch: true,
    rollupOptions: {
      input: {
        index: "src/index.html",
        new: "./src/new/index.html",
        new_snippet: "./src/new_snippet.html",
        app: "src/app.js"
      },
      watch: {
        include: ["./src/**/*"],
      }
    }
  }
})
