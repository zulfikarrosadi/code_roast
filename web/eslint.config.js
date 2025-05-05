import js from "@eslint/js";
import pkg from "eslint-plugin-svelte";
const { eslintPluginSvelteParser } = pkg;
import { includeIgnoreFile } from "@eslint/compat";
import sveltePlugin from "eslint-plugin-svelte";
import globals from "globals";
import { fileURLToPath } from "node:url";
import * as tseslint from "typescript-eslint";
import svelteConfig from "./svelte.config.js";

const gitignorePath = fileURLToPath(new URL("./.gitignore", import.meta.url));

// Prettier rules for ESLint
const prettierRules = {
  "arrow-body-style": "off",
  "prefer-arrow-callback": "off",
};

// Create typescript-eslint config with project specific settings
const typescript = tseslint.configs.recommended;

export default [
  // Include .gitignore
  includeIgnoreFile(gitignorePath),

  // Basic JavaScript rules
  js.configs.recommended,

  // TypeScript rules
  ...typescript,

  // Standard browser and node globals
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },

  // Svelte rules for .svelte files
  {
    files: ["**/*.svelte"],
    plugins: {
      svelte: sveltePlugin,
    },
    languageOptions: {
      parser: eslintPluginSvelteParser,
      parserOptions: {
        parser: tseslint.parser,
        extraFileExtensions: [".svelte"],
        svelteConfig,
      },
    },
    rules: {
      ...sveltePlugin.configs.recommended.rules,
      // Turn off no-undef because TypeScript handles this
      "no-undef": "off",
    },
  },

  // TypeScript files configuration
  {
    files: ["**/*.ts", "**/*.js"],
    ignores: ["eslint.config.js", "svelte.config.js", "node_modules/**"],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: "./tsconfig.json",
      },
    },
    rules: {
      ...prettierRules,
      "no-undef": "off", // TypeScript handles this
    },
  },

  // Prettier compatibility
  {
    rules: {
      ...prettierRules,
    },
  },
];
