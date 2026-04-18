# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tickets.test.js >> Ticket-Based E2E Verification >> A10: Single SPA Shell Entry - Direct access to sub-routes
- Location: SPA/tests/e2e/tickets.test.js:5:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('[data-screen="login"]')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('[data-screen="login"]')

```

# Page snapshot

```yaml
- generic [ref=e2]: <!DOCTYPE html> <html lang="en"> <head> <meta charset="UTF-8"> <meta name="viewport" content="width=device-width, initial-scale=1.0"> <title>Real-Time Forum</title> <!-- SEO Meta Tags --> <meta name="description" content="A modern, real-time forum experience with instant messaging and community interaction."> <meta name="theme-color" content="#0f172a"> <!-- Fonts --> <link rel="preconnect" href="https://fonts.googleapis.com"> <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin> <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Outfit:wght@400;500;600;700&display=swap" rel="stylesheet"> <!-- Styles --> <link rel="stylesheet" href="/assets/css/main.css"> <!-- Favicon --> <link rel="icon" type="image/x-icon" href="/favicon.ico"> </head> <body> <div id="loading-overlay" class="loading-overlay" aria-hidden="true"> <div class="spinner"></div> </div> <div id="app"> <!-- The persistent shell and features will be rendered here by main.js --> <main id="main-content" aria-live="polite"> <!-- Dynamic content injected here --> </main> </div> <!-- Main JavaScript Entry Point --> <script type="module" src="/main.js"></script> </body> </html>
```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  | 
  3  | test.describe('Ticket-Based E2E Verification', () => {
  4  |   
  5  |   test('A10: Single SPA Shell Entry - Direct access to sub-routes', async ({ page }) => {
  6  |     // Accessing a sub-route directly should load the SPA and render the correct screen
  7  |     page.on('console', msg => console.log('BROWSER CONSOLE:', msg.text()));
  8  |     page.on('pageerror', err => console.log('BROWSER ERROR:', err.message));
  9  |     await page.goto('/login');
  10 |     try {
> 11 |         await expect(page.locator('[data-screen="login"]')).toBeVisible({ timeout: 5000 });
     |                                                             ^ Error: expect(locator).toBeVisible() failed
  12 |     } catch (e) {
  13 |         console.log('Page content:', await page.content());
  14 |         throw e;
  15 |     }
  16 |     
  17 |     await page.goto('/register');
  18 |     await expect(page.locator('[data-screen="register"]')).toBeVisible({ timeout: 10000 });
  19 |   });
  20 | 
  21 |   test('A05: Auth Gating - Unauthenticated users redirected to login', async ({ page }) => {
  22 |     // Accessing protected routes should redirect to login
  23 |     const protectedRoutes = ['/activity', '/create-post'];
  24 |     for (const route of protectedRoutes) {
  25 |       await page.goto(route);
  26 |       // Wait for the redirect to happen
  27 |       await page.waitForURL(/\/login/, { timeout: 10000 });
  28 |       await expect(page).toHaveURL(/\/login/);
  29 |     }
  30 |   });
  31 | 
  32 |   test('A03: Client Routing - Navigation without full reload', async ({ page }) => {
  33 |     await page.goto('/login');
  34 |     
  35 |     // Clicking a link should change the URL and content without page reload
  36 |     // We can check if the window object persists, but usually checking the speed and data-screen is enough
  37 |     await page.click('a[href="/register"]');
  38 |     await expect(page).toHaveURL(/\/register/);
  39 |     await expect(page.locator('[data-screen="register"]')).toBeVisible();
  40 |     
  41 |     await page.goBack();
  42 |     await expect(page).toHaveURL(/\/login/);
  43 |     await expect(page.locator('[data-screen="login"]')).toBeVisible();
  44 |   });
  45 | 
  46 |   test('A04 & A06: Full Auth Journey - Registration, Login, Shell Persistence, and Logout', async ({ page }) => {
  47 |     const timestamp = Date.now();
  48 |     const username = `user_${timestamp}`;
  49 |     const email = `${username}@example.com`;
  50 | 
  51 |     // 1. Registration
  52 |     await page.goto('/register');
  53 |     await page.fill('input[name="username"]', username);
  54 |     await page.fill('input[name="email"]', email);
  55 |     await page.fill('input[name="password"]', 'password123');
  56 |     await page.fill('input[name="age"]', '25');
  57 |     await page.selectOption('select[name="gender"]', 'male');
  58 |     await page.fill('input[name="first_name"]', 'Test');
  59 |     await page.fill('input[name="last_name"]', 'User');
  60 |     await page.click('button[type="submit"]');
  61 | 
  62 |     // Should redirect to login after registration or home if auto-logged in
  63 |     // Real forum usually requires login after registration per SDS
  64 |     await expect(page).toHaveURL(/\/(login)?$/);
  65 | 
  66 |     // 2. Login
  67 |     if (page.url().includes('login')) {
  68 |         await page.fill('input[name="username"]', username);
  69 |         await page.fill('input[name="password"]', 'password123');
  70 |         await page.click('button[type="submit"]');
  71 |     }
  72 | 
  73 |     // 3. Verify Authenticated Shell (A04)
  74 |     await expect(page).toHaveURL(/\/$/);
  75 |     await expect(page.locator('[data-auth-shell]')).toBeVisible();
  76 |     await expect(page.locator('[data-action="logout"]')).toBeVisible();
  77 |     await expect(page.locator('[aria-label="Forum navigation"]')).toBeVisible();
  78 |     await expect(page.locator('[data-chat-roster]')).toBeVisible();
  79 | 
  80 |     // 4. Persistence across navigation (A04)
  81 |     await page.click('a[href="/activity"]');
  82 |     await expect(page.locator('[data-screen="activity"]')).toBeVisible();
  83 |     await expect(page.locator('[data-auth-shell]')).toBeVisible(); // Shell should still be there
  84 | 
  85 |     // 5. Logout (A06)
  86 |     await page.click('[data-action="logout"]');
  87 |     await expect(page).toHaveURL(/\/login/);
  88 |     await expect(page.locator('[data-auth-shell]')).not.toBeVisible();
  89 |   });
  90 | });
  91 | 
```