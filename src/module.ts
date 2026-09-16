import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { OhDearQuery, OhDearDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, OhDearQuery, OhDearDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
