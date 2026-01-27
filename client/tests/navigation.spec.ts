import { test, expect } from '@playwright/test';

test.describe('Admin Navigation', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'admin123');
    await page.getByRole('button', { name: 'Sign in' }).click();
    await page.waitForURL('http://localhost:5173/');
  });

  test('should navigate to Domains page', async ({ page }) => {
    // Click on System group to expand
    const systemGroup = page.getByRole('button', { name: 'System' });
    // Check if expanded or not. If not, click.
    // Simplifying: Just click it to toggle open (assuming closed initially)
    await systemGroup.click();

    // Click on Domains link in sidebar
    const domainsLink = page.getByRole('link', { name: 'Domains' }).first();
    await domainsLink.click();
    
    await page.waitForURL('**/domains');
    
    // Verify page title
    await expect(page.getByRole('heading', { name: 'Domains' })).toBeVisible();
  });

  test('should navigate to Users page', async ({ page }) => {
    // Click on System group to expand
    // Note: If previous test ran in same context, it might be open. 
    // But Playwright tests are isolated by default (new context per test).
    await page.getByRole('button', { name: 'System' }).click();

    // Click on Users link
    const usersLink = page.getByRole('link', { name: 'Users' }).first();
    await usersLink.click();

    await page.waitForURL('**/users');
    
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
  });
});
