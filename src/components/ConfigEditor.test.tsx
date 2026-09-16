import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DataSourcePluginOptionsEditorProps, DataSourceSettings, PluginType } from '@grafana/data';
import { ConfigEditor } from './ConfigEditor';
import { OhDearDataSourceOptions, OhDearSecureJsonData } from '../types';

const setup = (secureJsonFields: Record<string, boolean> = {}) => {
  const onOptionsChange = jest.fn();
  const options = {
    id: 1,
    uid: 'ohdear',
    name: 'Oh Dear',
    type: 'softwarehuset-ohdear-datasource',
    typeName: 'Oh Dear',
    typeLogoUrl: '',
    access: 'proxy',
    url: '',
    user: '',
    database: '',
    basicAuth: false,
    isDefault: false,
    readOnly: false,
    withCredentials: false,
    jsonData: {},
    secureJsonFields,
    meta: { type: PluginType.datasource },
  } as unknown as DataSourceSettings<OhDearDataSourceOptions, OhDearSecureJsonData>;

  render(
    <ConfigEditor
      {...({ options, onOptionsChange } as DataSourcePluginOptionsEditorProps<
        OhDearDataSourceOptions,
        OhDearSecureJsonData
      >)}
    />
  );

  return { onOptionsChange };
};

describe('ConfigEditor', () => {
  it('stores the API token in secure json data', async () => {
    const { onOptionsChange } = setup();

    await userEvent.type(screen.getByPlaceholderText('Enter your Oh Dear API token'), 'token');

    expect(onOptionsChange).toHaveBeenLastCalledWith(
      expect.objectContaining({ secureJsonData: { ohDearToken: 'token' } })
    );
  });

  it('stores the API url in json data', async () => {
    const { onOptionsChange } = setup();

    await userEvent.type(screen.getByPlaceholderText('https://ohdear.app/api'), 'h');

    expect(onOptionsChange).toHaveBeenLastCalledWith(expect.objectContaining({ jsonData: { baseUrl: 'h' } }));
  });

  it('marks a saved token as configured', () => {
    setup({ ohDearToken: true });

    expect(screen.getByDisplayValue('configured')).toBeInTheDocument();
  });
});
