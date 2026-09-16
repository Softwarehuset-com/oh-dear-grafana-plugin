import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { OhDearQuery, OhDearDataSourceOptions, OhDearMonitor, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<OhDearQuery, OhDearDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<OhDearDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<OhDearQuery> {
    return DEFAULT_QUERY;
  }

  async getMonitors(): Promise<OhDearMonitor[]> {
    return this.getResource('monitors', undefined, { showErrorAlert: false });
  }

  filterQuery(query: OhDearQuery): boolean {
    return !!query.queryType;
  }
}
