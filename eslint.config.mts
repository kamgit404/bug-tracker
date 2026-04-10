// eslint.config.mjs
import { defineConfig } from 'eslint/config';
import tseslint from 'typescript-eslint';
import playwright from 'eslint-plugin-playwright';
import eslintConfigPrettier from 'eslint-config-prettier';

export default defineConfig([
  {
    ignores: [
      'node_modules/**',
      'dist/**',
      'build/**',
      'coverage/**',
      'playwright-report/**',
      'test-results/**',
      'bugtracker-frontend/**',
      'bugtracker-backend/**',
    ],
  },

  // Base TypeScript rules for all TS files
  ...tseslint.configs.recommended,

  // Typed linting for stronger bug detection
  ...tseslint.configs.recommendedTypeChecked,

  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      // Good for automation repos
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-misused-promises': 'error',
      '@typescript-eslint/await-thenable': 'error',
      '@typescript-eslint/require-await': 'warn',
      '@typescript-eslint/consistent-type-imports': 'warn',
      '@typescript-eslint/no-explicit-any': 'warn',

      // Useful balance for test code
      '@typescript-eslint/no-unused-vars': [
        'warn',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
    },
  },

  // Playwright-specific rules only for test files
  {
    files: ['tests/**/*.ts', '**/*.spec.ts', '**/*.test.ts'],
    plugins: {
      playwright,
    },
    extends: ['playwright/recommended'],
    rules: {
      // Commonly useful adjustments for test code
      'playwright/no-skipped-test': 'warn',
      'playwright/no-focused-test': 'error',
    },
  },

  // Turn off formatting-related lint conflicts
  eslintConfigPrettier,
]);
