import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('should allow admin to login with valid credentials', async ({ page }) => {
    // 1. Navigate to Login Page
    await page.goto('/login');
    
    // 2. Verify Login Form is present
    await expect(page.getByText('Admin Login')).toBeVisible();
    await expect(page.getByText('Enter your credentials')).toBeVisible();
    
    // 3. Enter Credentials
    await page.fill('input[name="username"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'admin123');
    
    // 4. Click Submit
    await page.getByRole('button', { name: 'Sign in' }).click();

    // 5. Wait for navigation to Dashboard
    await page.waitForURL('http://localhost:5173/');
    
    // 6. Verify generic dashboard element or URL
    expect(page.url()).toBe('http://localhost:5173/');
    
    // 7. Verify Toast Success
    await expect(page.getByText('Logged in successfully')).toBeVisible();
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[name="username"]', 'wrong@example.com');
    await page.fill('input[name="password"]', 'wrongpass');
    
    await page.locator('button[type="submit"]').click();
    
    await expect(page.getByText('Invalid credentials')).toBeVisible();
  });
});
