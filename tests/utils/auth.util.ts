import type { Page } from '@playwright/test';

async function login(page: Page) {
  await page.goto('/auth/login');
  await page.getByLabel('Username').fill('arcane');
  await page.getByLabel('Password').fill('arcane-admin');
  await page.getByRole('button', { name: 'Sign in to Arcane', exact: true }).click();
}

export default { login };
