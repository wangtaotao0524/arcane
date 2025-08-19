import { defineConfig, devices } from '@playwright/test';

const authFile = '.auth/login.json';
const baseURL = process.env.BASE_URL || 'http://localhost:3000';
const composeFile = process.env.COMPOSE_FILE || 'setup/compose.yaml';

export default defineConfig({
  testDir: '.',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI
    ? [['html', { outputFolder: '.report' }], ['github']]
    : [['line'], ['html', { open: 'never', outputFolder: '.report' }]],
  use: {
    baseURL,
    trace: 'on-first-retry',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'setup',
      testMatch: /setup\/.*\.setup\.ts/,
    },
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], storageState: authFile },
      dependencies: ['setup'],
      testMatch: /spec\/.*\.spec\.ts/,
    },
  ],
  webServer: {
    command: `docker compose -f ${composeFile} up --abort-on-container-exit`,
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
});
