import { defineConfig } from 'vite'

export default defineConfig({
  root: "./src",
  publicDir: "public",
  assetsInclude: ["**/*.css"],
  build: {
    copyPublicDir: true,
    outDir: "../dist",
    watch: true,
    rollupOptions: {
      input: {
        index: "src/index.html",
        // partial: "./src/partial.tmpl",
        login: "./src/login/index.html",
        register: "./src/register/index.html",
        post: "./src/post/index.html",
        new_post: "./src/new/index.html",
        partial: "./src/partial.html",
        // style: "./src/assets/style.css",
        // assets: "./src/assets/",
        app: "src/app.js"
      },
      watch: {
        include: ["./src/**/*"],
      }
    }
  }
})
