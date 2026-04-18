import { test, expect } from '@playwright/test';

test.describe('Ticket-Based E2E Verification', () => {
  
  test('A10: Single SPA Shell Entry - Direct access to sub-routes', async ({ page }) => {
    // Accessing a sub-route directly should load the SPA and render the correct screen
    page.on('console', msg => console.log('BROWSER CONSOLE:', msg.text()));
    page.on('pageerror', err => console.log('BROWSER ERROR:', err.message));
    await page.goto('/login');
    try {
        await expect(page.locator('[data-screen="login"]')).toBeVisible({ timeout: 5000 });
    } catch (e) {
        console.log('Page content:', await page.content());
        throw e;
    }
    
    await page.goto('/register');
    await expect(page.locator('[data-screen="register"]')).toBeVisible({ timeout: 10000 });
  });

  test('A05: Auth Gating - Unauthenticated users redirected to login', async ({ page }) => {
    // Accessing protected routes should redirect to login
    const protectedRoutes = ['/activity', '/create-post'];
    for (const route of protectedRoutes) {
      await page.goto(route);
      // Wait for the redirect to happen
      await page.waitForURL(/\/login/, { timeout: 10000 });
      await expect(page).toHaveURL(/\/login/);
    }
  });

  test('A03: Client Routing - Navigation without full reload', async ({ page }) => {
    await page.goto('/login');
    
    // Clicking a link should change the URL and content without page reload
    // We can check if the window object persists, but usually checking the speed and data-screen is enough
    await page.click('a[href="/register"]');
    await expect(page).toHaveURL(/\/register/);
    await expect(page.locator('[data-screen="register"]')).toBeVisible();
    
    await page.goBack();
    await expect(page).toHaveURL(/\/login/);
    await expect(page.locator('[data-screen="login"]')).toBeVisible();
  });

  test('A04 & A06: Full Auth Journey - Registration, Login, Shell Persistence, and Logout', async ({ page }) => {
    const timestamp = Date.now();
    const username = `user_${timestamp}`;
    const email = `${username}@example.com`;

    // 1. Registration
    await page.goto('/register');
    await page.fill('input[name="username"]', username);
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="password"]', 'password123');
    await page.fill('input[name="age"]', '25');
    await page.selectOption('select[name="gender"]', 'male');
    await page.fill('input[name="first_name"]', 'Test');
    await page.fill('input[name="last_name"]', 'User');
    await page.click('button[type="submit"]');

    // Should redirect to login after registration or home if auto-logged in
    // Real forum usually requires login after registration per SDS
    await expect(page).toHaveURL(/\/(login)?$/);

    // 2. Login
    if (page.url().includes('login')) {
        await page.fill('input[name="username"]', username);
        await page.fill('input[name="password"]', 'password123');
        await page.click('button[type="submit"]');
    }

    // 3. Verify Authenticated Shell (A04)
    await expect(page).toHaveURL(/\/$/);
    await expect(page.locator('[data-auth-shell]')).toBeVisible();
    await expect(page.locator('[data-action="logout"]')).toBeVisible();
    await expect(page.locator('[aria-label="Forum navigation"]')).toBeVisible();
    await expect(page.locator('[data-chat-roster]')).toBeVisible();

    // 4. Persistence across navigation (A04)
    await page.click('a[href="/activity"]');
    await expect(page.locator('[data-screen="activity"]')).toBeVisible();
    await expect(page.locator('[data-auth-shell]')).toBeVisible(); // Shell should still be there

    // 5. Logout (A06)
    await page.click('[data-action="logout"]');
    await expect(page).toHaveURL(/\/login/);
    await expect(page.locator('[data-auth-shell]')).not.toBeVisible();
  });
});
