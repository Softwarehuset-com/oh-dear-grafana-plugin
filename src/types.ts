import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export const QueryTypes = {
  Monitors: 'monitors',
  Uptime: 'uptime',
  Downtime: 'downtime',
  HTTPMetrics: 'http-uptime-metrics',
  PingMetrics: 'ping-uptime-metrics',
  TCPMetrics: 'tcp-uptime-metrics',
  Lighthouse: 'lighthouse',
} as const;

export type OhDearQueryType = (typeof QueryTypes)[keyof typeof QueryTypes];

export type GroupBy = 'hour' | 'day' | 'month';

export interface OhDearQuery extends DataQuery {
  queryType?: OhDearQueryType;
  monitorId?: number;
  groupBy?: GroupBy;
}

export const DEFAULT_QUERY: Partial<OhDearQuery> = {
  queryType: QueryTypes.Uptime,
  groupBy: 'hour',
};

export interface OhDearMonitor {
  id: number;
  label: string;
  url: string;
  type: string;
  status: string;
}

export interface OhDearDataSourceOptions extends DataSourceJsonData {
  baseUrl?: string;
}

export interface OhDearSecureJsonData {
  ohDearToken?: string;
}
