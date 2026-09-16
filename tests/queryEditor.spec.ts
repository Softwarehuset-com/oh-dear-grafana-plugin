import { test, expect } from '@grafana/plugin-e2e';
import type { Locator } from '@playwright/test';

const monitorsLoaded = (row: Locator) => expect(row.locator('[aria-busy]')).toHaveAttribute('aria-busy', 'false');

test('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Query type' })).toBeVisible();
});

test('should list the supported query types', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  const row = panelEditPage.getQueryEditorRow('A');
  await monitorsLoaded(row);
  await row.getByRole('combobox', { name: 'Query type' }).click();
  await expect(page.getByRole('option')).toHaveText([
    /^Monitors/,
    /^Uptime/,
    /^Downtime/,
    /^HTTP metrics/,
    /^Ping metrics/,
    /^TCP metrics/,
    /^Lighthouse/,
  ]);
});

test('should hide the monitor and group by pickers for the monitors query type', async ({
  panelEditPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  const row = panelEditPage.getQueryEditorRow('A');
  await monitorsLoaded(row);

  await expect(row.getByRole('combobox', { name: 'Monitor' })).toBeVisible();
  await expect(row.getByRole('combobox', { name: 'Group by' })).toBeVisible();

  await row.getByRole('combobox', { name: 'Query type' }).click();
  await page.getByRole('option', { name: 'Monitors' }).click();

  await expect(row.getByRole('combobox', { name: 'Monitor' })).toBeHidden();
  await expect(row.getByRole('combobox', { name: 'Group by' })).toBeHidden();
});

test('should keep the group by picker for time grouped query types only', async ({
  panelEditPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  const row = panelEditPage.getQueryEditorRow('A');
  await monitorsLoaded(row);

  await row.getByRole('combobox', { name: 'Query type' }).click();
  await page.getByRole('option', { name: 'Downtime' }).click();

  await expect(row.getByRole('combobox', { name: 'Monitor' })).toBeVisible();
  await expect(row.getByRole('combobox', { name: 'Group by' })).toBeHidden();
});
