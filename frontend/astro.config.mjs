import astroSingleFile from 'astro-single-file'
import { defineConfig } from 'astro/config'
import tailwind from '@astrojs/tailwind'

// https://astro.build/config
export default defineConfig({
  integrations: [astroSingleFile(), tailwind()]
})
