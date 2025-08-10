// prettier.config.js
import sveltePlugin from 'prettier-plugin-svelte'

export default {
  plugins: [sveltePlugin],
  pluginSearchDirs: false,
  overrides: [
    {
      files: '*.svelte',
      options: {
        parser: 'svelte',
      },
    },
  ],
  // Optional: Tailwind CSS plugin and other rules
  tailwindConfig: './tailwind.config.js',
  printWidth: 100,
  singleQuote: true,
  semi: false,
}
