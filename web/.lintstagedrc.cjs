module.exports = {
  // Match all src files we care about, but filter out the shadcn ui folder
  'src/**/*.{js,ts,svelte}': (filenames) => {
    const files = filenames
      .map(f => f.replace(/\\/g, '/')) // normalize windows paths
      .filter(f => !f.startsWith('src/lib/components/ui/'));

    if (files.length === 0) return [];
    return [
      `prettier --plugin=prettier-plugin-svelte --write ${files.join(' ')}`,
      `eslint --fix ${files.join(' ')}`
    ];
  },

  // Other non-src files to format with prettier only
  '*.{js,ts,css,scss,postcss,md,json}': (filenames) => {
    const files = filenames
      .map(f => f.replace(/\\/g, '/'))
      .filter(f => !f.startsWith('src/lib/components/ui/'));

    if (files.length === 0) return [];
    return `prettier --plugin=prettier-plugin-svelte --write ${files.join(' ')}`;
  }
};
