import { test, expect } from '@grafana/plugin-e2e';
import { OhDearDataSourceOptions, OhDearSecureJsonData } from '../src/types';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();
  await expect(page.getByRole('textbox', { name: 'API URL' })).toBeVisible();
});

test('"Save & test" should fail when the API token is missing', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<OhDearDataSourceOptions, OhDearSecureJsonData>({
    fileName: 'datasources.yml',
  });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'API URL' }).fill(ds.jsonData.baseUrl ?? '');
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: 'API token is missing' });
});

test('"Save & test" should fail when the API token is invalid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<OhDearDataSourceOptions, OhDearSecureJsonData>({
    fileName: 'datasources.yml',
  });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'API URL' }).fill(ds.jsonData.baseUrl ?? '');
  await page.getByRole('textbox', { name: 'API token' }).fill('not-a-valid-token');
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: 'Invalid API token' });
});
