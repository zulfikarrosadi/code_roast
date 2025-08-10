import js from '@eslint/js'
import svelte from 'eslint-plugin-svelte'
import tseslint from 'typescript-eslint'
import { includeIgnoreFile } from '@eslint/compat'
import { fileURLToPath } from 'node:url'
import eslintConfigPrettier from 'eslint-config-prettier'

const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url))

export default tseslint.config(
  {
    ignores: ['build/', '.svelte-kit/', 'dist/', 'src/lib/components/ui/**/*'],
  },

  includeIgnoreFile(gitignorePath),

  js.configs.recommended,
  ...tseslint.configs.recommended,

  ...svelte.configs['flat/recommended'],

  eslintConfigPrettier,

  {
    files: ['**/*.svelte'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
        project: true,
        tsconfigRootDir: import.meta.dirname,
        extraFileExtensions: ['.svelte'],
      },
    },
    rules: {
      // You can place Svelte-specific rule overrides here
    },
  },

  // Global rules for all files
  {
    rules: {
      'no-undef': 'off', // TypeScript handles undefined variables
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
    },
  },
)
