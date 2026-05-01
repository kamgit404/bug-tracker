import { defineConfig } from '@playwright/test';
import { config } from 'dotenv';
import path from 'path';

config({
  path: path.resolve(process.cwd(), '../.env'),
});

export default defineConfig({
  testDir: '.',
  testMatch: 'integration.spec.ts',
  use: {
    baseURL: process.env.STAGING_APP_BASE_URL ?? 'http://localhost:3000',
    trace: 'on-first-retry',
    headless: process.env.CI ? true : false,
    launchOptions: {
      slowMo: process.env.CI ? 0 : 1000,
    },
  },
  timeout: 30000,
  reporter: [
    ['list'],
    ['junit', { outputFile: 'test-results/results.xml' }],
    ['html', { outputFolder: 'playwright-report' }],
  ],
});
