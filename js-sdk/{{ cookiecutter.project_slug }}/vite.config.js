import { defineConfig } from 'vite'
import checker from 'vite-plugin-checker'
import dts from 'vite-plugin-dts'

export default defineConfig({
  plugins: [checker({ typescript: true }), dts()],
  build: {
    minify: true, // temporarily, disable minify
    lib: {
      entry: 'src/index.ts',
      name: 'example-sdk',
      fileName: 'example-sdk',
    },
    rollupOptions: {
      output: {
        exports: 'named',
        format: 'umd',
        name: 'rabpid',
      },
    },
  },
})
