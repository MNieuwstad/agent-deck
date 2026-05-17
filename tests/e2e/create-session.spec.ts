import { test, expect } from '@playwright/test'

test('create session dialog opens via new session button', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Sessions')).toBeVisible({ timeout: 10000 })

  // Find the "New session" button in the sidebar header (aria-label="New session")
  const newBtn = page.locator('button[aria-label="New session"]')
  if (await newBtn.isVisible()) {
    await newBtn.click()
    // Dialog renders a form for session creation
    const dialog = page.locator('form')
    await expect(dialog.first()).toBeVisible({ timeout: 5000 })
  }
})

test('create session dialog opens via keyboard shortcut n', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Sessions')).toBeVisible({ timeout: 10000 })

  // Press 'n' to open create session dialog (keyboard shortcut in AppShell)
  await page.keyboard.press('n')

  // Dialog should be visible (contains "New Session" heading)
  await expect(page.getByText('New Session')).toBeVisible({ timeout: 5000 })
})

test('create session dialog shows all seven tools', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Sessions')).toBeVisible({ timeout: 10000 })

  // Open the dialog
  const newBtn = page.locator('button[aria-label="New session"]')
  if (await newBtn.isVisible()) {
    await newBtn.click()
  } else {
    await page.keyboard.press('n')
  }
  await expect(page.locator('form')).toBeVisible({ timeout: 5000 })

  // All seven built-in tools from PR-1 must appear as seg-btn buttons.
  const expectedTools = ['claude', 'codex', 'gemini', 'opencode', 'copilot', 'pi', 'shell']
  for (const tool of expectedTools) {
    await expect(
      page.locator('.seg-btn', { hasText: tool }),
      `tool button "${tool}" must be visible in the dialog`
    ).toBeVisible({ timeout: 3000 })
  }
})

test('create session dialog shows group selector', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Sessions')).toBeVisible({ timeout: 10000 })

  const newBtn = page.locator('button[aria-label="New session"]')
  if (await newBtn.isVisible()) {
    await newBtn.click()
  } else {
    await page.keyboard.press('n')
  }
  await expect(page.locator('form')).toBeVisible({ timeout: 5000 })

  // The GROUP label must be present (field label, text-transform:uppercase in CSS).
  await expect(
    page.locator('.field label', { hasText: /^GROUP$/i }),
    'GROUP label must be visible'
  ).toBeVisible({ timeout: 3000 })

  // A <select> element for group selection must be present inside the dialog form.
  await expect(
    page.locator('form select'),
    'group <select> must be present inside the dialog'
  ).toBeVisible({ timeout: 3000 })

  // The "No group" sentinel option must exist.
  const noGroup = page.locator('form select option', { hasText: /no group/i })
  await expect(noGroup).toHaveCount(1)
})

test('create session dialog closes on Escape', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Sessions')).toBeVisible({ timeout: 10000 })

  // Open dialog via keyboard shortcut
  await page.keyboard.press('n')
  await expect(page.getByText('New Session')).toBeVisible({ timeout: 5000 })

  // Close with Escape
  await page.keyboard.press('Escape')

  // Dialog should be hidden
  await expect(page.getByText('New Session')).not.toBeVisible({ timeout: 3000 })
})
