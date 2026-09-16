import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { OhDearDataSourceOptions, OhDearSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<OhDearDataSourceOptions, OhDearSecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onBaseUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        baseUrl: event.target.value,
      },
    });
  };

  const onTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ohDearToken: event.target.value,
      },
    });
  };

  const onResetToken = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        ohDearToken: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        ohDearToken: '',
      },
    });
  };

  return (
    <>
      <InlineField
        label="API token"
        labelWidth={18}
        interactive
        tooltip="Personal API token from ohdear.app under Settings > API tokens"
      >
        <SecretInput
          required
          id="config-editor-token"
          isConfigured={secureJsonFields.ohDearToken}
          value={secureJsonData?.ohDearToken}
          placeholder="Enter your Oh Dear API token"
          width={40}
          onReset={onResetToken}
          onChange={onTokenChange}
        />
      </InlineField>
      <InlineField
        label="API URL"
        labelWidth={18}
        interactive
        tooltip="Only change this if you are pointing at a different Oh Dear API endpoint"
      >
        <Input
          id="config-editor-base-url"
          onChange={onBaseUrlChange}
          value={jsonData.baseUrl ?? ''}
          placeholder="https://ohdear.app/api"
          width={40}
        />
      </InlineField>
    </>
  );
}
