import { test, expect, type Page } from '@playwright/test';

async function fetchVolumesWithRetry(page: Page, maxRetries = 3): Promise<any[]> {
  let retries = 0;
  while (retries < maxRetries) {
    try {
      const response = await page.request.get('/api/volumes');
      const volumes = await response.json();
      return volumes.data;
    } catch (error) {
      retries++;
      console.log(`Attempt ${retries} failed, ${maxRetries - retries} retries left`);
      if (retries >= maxRetries) throw error;
      await new Promise((resolve) => setTimeout(resolve, 1000));
    }
  }
  return [];
}

let realVolumes: any[] = [];

test.beforeEach(async ({ page }) => {
  await page.goto('/volumes');
  await page.waitForLoadState('networkidle');
  realVolumes = await fetchVolumesWithRetry(page);
});

async function setDropdownMenuItemCheckbox(page: Page, label: string, desired: boolean) {
  const menu = page.locator('[role="menu"]').last();
  await expect(menu).toBeVisible();

  const desiredValue = desired ? 'true' : 'false';

  for (let attempt = 0; attempt < 5; attempt++) {
    const item = menu.getByRole('menuitemcheckbox', { name: label });

    await item.waitFor({ state: 'visible' });

    const current = (await item.getAttribute('aria-checked')) === 'true';
    if (current === desired) return;

    await item.focus();
    await expect(item).toBeFocused();

    await page.waitForTimeout(50);

    await page.keyboard.press(' ');

    try {
      await expect(item).toHaveAttribute('aria-checked', desiredValue, { timeout: 1000 });
      return;
    } catch {
      continue;
    }
  }

  await expect(menu.getByRole('menuitemcheckbox', { name: label })).toHaveAttribute('aria-checked', desiredValue);
}

test.describe('Volumes Page', () => {
  test('Volume Page Display', async ({ page }) => {
    await page.goto('/volumes');

    await expect(page.getByRole('heading', { name: 'Volumes', level: 1 })).toBeVisible();
    await expect(page.getByText('Manage your Docker volumes').first()).toBeVisible();
  });

  test('Correct Volume Stat Card Counts', async ({ page }) => {
    await page.goto('/volumes');

    await expect(page.locator('p:has-text("Total Volumes") + p')).toHaveText(realVolumes.length.toString());
  });

  test('Tables Displays When Volumes Exist', async ({ page }) => {
    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    if (realVolumes.length > 0) {
      await expect(page.getByText('Volumes List')).toBeVisible();
      await expect(page.locator('table')).toBeVisible();
    } else {
      await expect(page.getByText('No volumes found')).toBeVisible();
    }
  });

  test('Create Volume Sheet Opens', async ({ page }) => {
    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    await page.locator('button:has-text("Create Volume")').first().click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Create New Volume')).toBeVisible();
  });

  test('Display Volume Filters', async ({ page }) => {
    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    await page.locator('button:has-text("Filter")').click();
    await expect(page.getByText('Volume Usage')).toBeVisible();
    await expect(page.getByText('Show Used Volumes')).toBeVisible();
    await expect(page.getByText('Show Unused Volumes')).toBeVisible();
  });

  test('Inspect Volume', async ({ page }) => {
    test.skip(!realVolumes.length, 'No volumes available for inspect test');

    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    const firstRow = page.locator('tbody tr').first();
    await firstRow.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Inspect' }).click();

    await expect(page).toHaveURL(new RegExp(`/volumes/.+`));
  });

  test('Remove Volume', async ({ page }) => {
    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    const firstRow = await page.getByRole('row', { name: 'my-app-data' });
    await firstRow.getByRole('button', { name: 'Open menu' }).click();

    await page.getByRole('menuitem', { name: 'Remove' }).click();

    await expect(page.getByRole('heading', { name: 'Remove Volume' })).toBeVisible();

    await page.locator('button:has-text("Remove")').click();

    await expect(page.locator('li[data-sonner-toast][data-type="success"] div[data-title]')).toBeVisible();
  });

  test('Create Volume', async ({ page }) => {
    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    await page.locator('button:has-text("Create Volume")').first().click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const volumeName = `test-volume-${Date.now()}`;
    await page.locator('input[id="volume-name-*"]').fill(volumeName);

    await page.getByRole('dialog').locator('button:has-text("Create Volume")').click();

    await expect(
      page.locator(`li[data-sonner-toast][data-type="success"] div[data-title]:has-text("created successfully")`)
    ).toBeVisible();
  });

  test('Disable remove action when volume is in use', async ({ page }) => {
    const volumeInUse = realVolumes.find((vol) => vol.inUse);

    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');
    const volumeRow = page.locator(`tr:has-text("${volumeInUse.name}")`);
    await volumeRow.getByRole('button', { name: 'Open menu' }).click();

    await expect(page.getByRole('menuitem', { name: 'Remove' })).toBeDisabled();
  });

  test('Display correct volume usage badge', async ({ page }) => {
    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    const usedVolumes = realVolumes.filter((vol) => vol.inUse);
    const unusedVolumes = realVolumes.filter((vol) => !vol.inUse);

    if (usedVolumes.length > 0) {
      await expect(page.locator('text="In Use"').first()).toBeVisible();
    }

    if (unusedVolumes.length > 0) {
      await expect(page.locator('text="Unused"').first()).toBeVisible();
    }
  });

  test('Filter Volumes by Usage Status', async ({ page }) => {
    const usedVolumes = realVolumes.filter((v) => v.inUse);
    const unusedVolumes = realVolumes.filter((v) => !v.inUse);

    await page.goto('/volumes');
    await page.waitForLoadState('networkidle');

    // Show only Used
    await page.locator('button:has-text("Filter")').click();
    await setDropdownMenuItemCheckbox(page, 'Show Used Volumes', true);
    await setDropdownMenuItemCheckbox(page, 'Show Unused Volumes', false);
    await page.keyboard.press('Escape');
    await expect(page.locator('tbody tr')).toHaveCount(usedVolumes.length);

    // Show only Unused
    await page.locator('button:has-text("Filter")').click();
    await setDropdownMenuItemCheckbox(page, 'Show Used Volumes', false);
    await setDropdownMenuItemCheckbox(page, 'Show Unused Volumes', true);
    await page.keyboard.press('Escape');
    await expect(page.locator('tbody tr')).toHaveCount(unusedVolumes.length);
  });
});
