---
layout: skill
title: QA Automation
name: qa-automation
category: product
description: Write effective automated tests including E2E, visual regression, accessibility, and performance testing.
---

# QA Automation

Guidelines for building comprehensive automated testing that catches bugs before production.

## When to Activate

- Writing E2E tests
- Setting up visual regression testing
- Adding accessibility tests
- Implementing performance testing

## E2E Testing with Playwright

### Page Object Model

```typescript
// pages/LoginPage.ts
import { Page } from '@playwright/test';

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
  }

  async login(email: string, password: string) {
    await this.page.fill('[data-testid="email"]', email);
    await this.page.fill('[data-testid="password"]', password);
    await this.page.click('[data-testid="submit"]');
  }

  async expectError(message: string) {
    await expect(this.page.locator('[data-testid="error"]'))
      .toContainText(message);
  }
}

// tests/login.spec.ts
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Login', () => {
  test('successful login redirects to dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('user@example.com', 'password123');
    
    await expect(page).toHaveURL('/dashboard');
  });

  test('invalid credentials show error', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('wrong@email.com', 'wrong');
    
    await loginPage.expectError('Invalid credentials');
  });
});
```

### Use data-testid for selectors

```tsx
// GOOD - stable selectors
<button data-testid="submit-order">Place Order</button>
<input data-testid="email-input" type="email" />
<div data-testid="order-confirmation">Order #123</div>

// Test
await page.click('[data-testid="submit-order"]');
await expect(page.locator('[data-testid="order-confirmation"]'))
  .toBeVisible();

// BAD - fragile selectors
await page.click('.btn.btn-primary.submit');
await page.click('button:has-text("Place Order")');
```

### Test realistic user flows

```typescript
test('complete checkout flow', async ({ page }) => {
  // Start from real entry point
  await page.goto('/');
  
  // Add product to cart
  await page.click('[data-testid="product-card"]:first-child');
  await page.click('[data-testid="add-to-cart"]');
  
  // Go to checkout
  await page.click('[data-testid="cart-icon"]');
  await page.click('[data-testid="checkout-button"]');
  
  // Fill shipping
  await page.fill('[data-testid="address"]', '123 Main St');
  await page.fill('[data-testid="city"]', 'São Paulo');
  await page.click('[data-testid="continue-to-payment"]');
  
  // Fill payment
  await page.fill('[data-testid="card-number"]', '4242424242424242');
  await page.fill('[data-testid="expiry"]', '12/28');
  await page.fill('[data-testid="cvv"]', '123');
  
  // Complete order
  await page.click('[data-testid="place-order"]');
  
  // Verify success
  await expect(page.locator('[data-testid="confirmation"]'))
    .toContainText('Order confirmed');
});
```

## Visual Regression Testing

### Playwright visual comparisons

```typescript
import { test, expect } from '@playwright/test';

test('homepage visual regression', async ({ page }) => {
  await page.goto('/');
  
  // Full page screenshot
  await expect(page).toHaveScreenshot('homepage.png', {
    fullPage: true,
    maxDiffPixels: 100, // Allow minor differences
  });
});

test('component visual states', async ({ page }) => {
  await page.goto('/storybook/button');
  
  // Default state
  await expect(page.locator('[data-testid="button"]'))
    .toHaveScreenshot('button-default.png');
  
  // Hover state
  await page.hover('[data-testid="button"]');
  await expect(page.locator('[data-testid="button"]'))
    .toHaveScreenshot('button-hover.png');
  
  // Disabled state
  await page.goto('/storybook/button?disabled=true');
  await expect(page.locator('[data-testid="button"]'))
    .toHaveScreenshot('button-disabled.png');
});
```

### Configure for CI

```typescript
// playwright.config.ts
export default defineConfig({
  expect: {
    toHaveScreenshot: {
      maxDiffPixelRatio: 0.01, // 1% difference allowed
    },
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile',
      use: { ...devices['iPhone 13'] },
    },
  ],
});
```

## Accessibility Testing

### Automated a11y checks with axe-core

```typescript
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Accessibility', () => {
  test('homepage has no a11y violations', async ({ page }) => {
    await page.goto('/');
    
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze();
    
    expect(results.violations).toEqual([]);
  });

  test('form has proper labels', async ({ page }) => {
    await page.goto('/signup');
    
    const results = await new AxeBuilder({ page })
      .include('form')
      .analyze();
    
    expect(results.violations).toEqual([]);
  });
});
```

### Test keyboard navigation

```typescript
test('can complete form with keyboard only', async ({ page }) => {
  await page.goto('/contact');
  
  // Tab through form
  await page.keyboard.press('Tab');
  await expect(page.locator('[data-testid="name"]')).toBeFocused();
  
  await page.keyboard.type('John Doe');
  await page.keyboard.press('Tab');
  await expect(page.locator('[data-testid="email"]')).toBeFocused();
  
  await page.keyboard.type('john@example.com');
  await page.keyboard.press('Tab');
  await page.keyboard.press('Tab'); // Skip to submit
  await page.keyboard.press('Enter');
  
  await expect(page.locator('[data-testid="success"]')).toBeVisible();
});
```

### Screen reader testing

```typescript
test('important content has ARIA labels', async ({ page }) => {
  await page.goto('/dashboard');
  
  // Check navigation has label
  const nav = page.locator('nav');
  await expect(nav).toHaveAttribute('aria-label', 'Main navigation');
  
  // Check images have alt text
  const images = page.locator('img');
  for (const img of await images.all()) {
    const alt = await img.getAttribute('alt');
    expect(alt).toBeTruthy();
  }
  
  // Check buttons have accessible names
  const buttons = page.locator('button');
  for (const button of await buttons.all()) {
    const name = await button.getAttribute('aria-label') 
      || await button.textContent();
    expect(name?.trim()).toBeTruthy();
  }
});
```

## Performance Testing

### Lighthouse in CI

```typescript
import { test } from '@playwright/test';
import { playAudit } from 'playwright-lighthouse';

test('homepage performance meets budget', async ({ page }) => {
  await page.goto('/');
  
  const results = await playAudit({
    page,
    thresholds: {
      performance: 80,
      accessibility: 90,
      'best-practices': 80,
      seo: 90,
    },
    port: 9222,
  });
  
  expect(results.lhr.categories.performance.score * 100)
    .toBeGreaterThan(80);
});
```

### Custom performance metrics

```typescript
test('page loads within budget', async ({ page }) => {
  await page.goto('/');
  
  // Measure Core Web Vitals
  const metrics = await page.evaluate(() => ({
    lcp: performance.getEntriesByType('largest-contentful-paint')[0]?.startTime,
    fid: performance.getEntriesByType('first-input')[0]?.processingStart,
    cls: performance.getEntriesByType('layout-shift')
      .reduce((sum, e) => sum + e.value, 0),
  }));
  
  expect(metrics.lcp).toBeLessThan(2500); // 2.5s
  expect(metrics.cls).toBeLessThan(0.1);
});
```

## Test Organization

### Follow AAA pattern

```typescript
test('user can update profile', async ({ page }) => {
  // Arrange - setup preconditions
  await loginAsUser(page, testUser);
  await page.goto('/settings/profile');
  
  // Act - perform the action
  await page.fill('[data-testid="display-name"]', 'New Name');
  await page.click('[data-testid="save-profile"]');
  
  // Assert - verify outcomes
  await expect(page.locator('[data-testid="success-toast"]'))
    .toContainText('Profile updated');
  await expect(page.locator('[data-testid="display-name"]'))
    .toHaveValue('New Name');
});
```

### Use fixtures for common setup

```typescript
// fixtures.ts
import { test as base } from '@playwright/test';

type Fixtures = {
  loggedInPage: Page;
  adminPage: Page;
};

export const test = base.extend<Fixtures>({
  loggedInPage: async ({ page }, use) => {
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'user@test.com');
    await page.fill('[data-testid="password"]', 'password');
    await page.click('[data-testid="submit"]');
    await page.waitForURL('/dashboard');
    await use(page);
  },
  
  adminPage: async ({ page }, use) => {
    await page.goto('/admin/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'adminpass');
    await page.click('[data-testid="submit"]');
    await use(page);
  },
});

// Usage
test('user sees their orders', async ({ loggedInPage }) => {
  await loggedInPage.goto('/orders');
  await expect(loggedInPage.locator('[data-testid="orders-list"]'))
    .toBeVisible();
});
```

## CI Integration

### Playwright in GitHub Actions

```yaml
# .github/workflows/e2e.yml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          
      - name: Install dependencies
        run: npm ci
        
      - name: Install Playwright browsers
        run: npx playwright install --with-deps
        
      - name: Run E2E tests
        run: npx playwright test
        
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: playwright-report
          path: playwright-report/
```
