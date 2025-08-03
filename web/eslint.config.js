import js from '@eslint/js'
import svelte from 'eslint-plugin-svelte'
import tseslint from 'typescript-eslint'
import globals from 'globals'
import { includeIgnoreFile } from '@eslint/compat'
import { fileURLToPath } from 'node:url'
import eslintConfigPrettier from 'eslint-config-prettier'

const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url))

export default tseslint.config(
  {
    ignores: ['build/', '.svelte-kit/', 'dist/'],
  },

  includeIgnoreFile(gitignorePath),

  js.configs.recommended,

  ...tseslint.configs.recommended,

  ...svelte.configs['flat/recommended'],

  eslintConfigPrettier,

  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    rules: {
      'no-undef': 'off',
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

  {
    files: ['**/*.ts'],
    rules: {},
  },
)
