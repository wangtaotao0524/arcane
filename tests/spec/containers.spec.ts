import { test, expect, type Page } from '@playwright/test';

type ContainerSummary = {
  id: string;
  names?: string[];
  image?: string;
  state: string;
  status?: string;
  created?: number; // seconds
};

type Paginated<T> = { data: T[]; pagination?: { totalItems?: number } };

async function fetchContainersWithRetry(page: Page, maxRetries = 3): Promise<Paginated<ContainerSummary>> {
  let retries = 0;
  while (retries < maxRetries) {
    try {
      const response = await page.request.get('/api/environments/0/containers');
      if (!response.ok()) throw new Error(`HTTP ${response.status()}`);
      const body = await response.json().catch(() => null as any);
      const data = Array.isArray(body?.data) ? (body.data as ContainerSummary[]) : [];
      const pagination = body?.pagination || { totalItems: data.length };
      return { data, pagination };
    } catch {
      retries++;
      if (retries >= maxRetries) break;
      await page.waitForTimeout(1000);
    }
  }
  return { data: [], pagination: { totalItems: 0 } };
}

let containersData: Paginated<ContainerSummary> = { data: [], pagination: { totalItems: 0 } };

test.beforeEach(async ({ page }) => {
  await page.goto('/containers');
  await page.waitForLoadState('networkidle');
  containersData = await fetchContainersWithRetry(page);
  console.log(`Found ${containersData.data.length} containers (totalItems=${containersData.pagination?.totalItems ?? 0})`);
});

test.describe('Containers Page', () => {
  test('should display the containers page title and description', async ({ page }) => {
    await page.goto('/containers');
    await expect(page.getByRole('heading', { name: 'Containers', level: 1 })).toBeVisible();
    await expect(page.getByText('View and Manage your Containers').first()).toBeVisible();
  });

  test('should display stat cards with correct counts', async ({ page }) => {
    await page.goto('/containers');

    const total = containersData.pagination?.totalItems ?? containersData.data.length;
    const running = containersData.data.filter((c) => c.state === 'running').length;
    const stopped = containersData.data.filter((c) => c.state !== 'running').length;

    await expect(page.locator('p:has-text("Total") + p')).toHaveText(String(total));
    await expect(page.locator('p:has-text("Running") + p')).toHaveText(String(running));
    await expect(page.locator('p:has-text("Stopped") + p')).toHaveText(String(stopped));
  });

  test('should display the container table with columns', async ({ page }) => {
    await page.goto('/containers');
    await page.waitForLoadState('networkidle');

    await expect(page.locator('table')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Name' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'ID' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Image' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'State' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Created' })).toBeVisible();
  });

  test('should navigate to container details on Inspect', async ({ page }) => {
    test.skip(containersData.data.length === 0, 'No containers available');
    await page.goto('/containers');
    await page.waitForLoadState('networkidle');

    const firstRow = page.locator('tbody tr').first();
    await firstRow.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Inspect' }).click();

    await expect(page).toHaveURL(/\/containers\/.+/);
    await expect(page.getByRole('heading', { name: 'Container Details', level: 2 }).first()).toBeVisible();
  });

  test('should show correct actions based on container state (without changing state)', async ({ page }) => {
    const running = containersData.data.find((c) => c.state === 'running');
    const stopped = containersData.data.find((c) => c.state !== 'running');

    await page.goto('/containers');
    await page.waitForLoadState('networkidle');

    if (running) {
      const row = page.locator(`tr:has(a[href="/containers/${running.id}/"])`);
      await row.getByRole('button', { name: 'Open menu' }).click();
      await expect(page.getByRole('menuitem', { name: 'Restart' })).toBeVisible();
      await expect(page.getByRole('menuitem', { name: 'Stop' })).toBeVisible();
      await page.keyboard.press('Escape'); // close menu
    } else {
      test.info().annotations.push({ type: 'note', description: 'No running container to validate actions' });
    }

    if (stopped) {
      const row = page.locator(`tr:has(a[href="/containers/${stopped.id}/"])`);
      await row.getByRole('button', { name: 'Open menu' }).click();
      await expect(page.getByRole('menuitem', { name: 'Start' })).toBeVisible();
      await page.keyboard.press('Escape');
    } else {
      test.info().annotations.push({ type: 'note', description: 'No stopped container to validate actions' });
    }
  });

  test('should open the Remove dialog from row actions and allow cancel', async ({ page }) => {
    test.skip(containersData.data.length === 0, 'No containers available');
    const any = containersData.data[0];

    await page.goto('/containers');
    await page.waitForLoadState('networkidle');

    const row = page.locator(`tr:has(a[href="/containers/${any.id}/"])`);
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Remove' }).click();

    const dialog = page.locator('div[role="heading"][aria-level="2"][data-dialog-title]:has-text("Confirm Container Removal")');
    await expect(dialog).toBeVisible();

    // Cancel removal (do not mutate)
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(dialog).toBeHidden();
  });
});
