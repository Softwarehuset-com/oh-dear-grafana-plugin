import React, { useEffect, useState } from 'react';
import { Alert, Combobox, ComboboxOption, InlineField, Stack } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import type { DataSource } from '../datasource';
import { DEFAULT_QUERY, GroupBy, OhDearDataSourceOptions, OhDearQuery, OhDearQueryType, QueryTypes } from '../types';

type Props = QueryEditorProps<DataSource, OhDearQuery, OhDearDataSourceOptions>;

const queryTypeOptions: Array<ComboboxOption<OhDearQueryType>> = [
  { label: 'Monitors', value: QueryTypes.Monitors, description: 'All monitors and their current status' },
  { label: 'Uptime', value: QueryTypes.Uptime, description: 'Uptime percentage over time' },
  { label: 'Downtime', value: QueryTypes.Downtime, description: 'Downtime periods in the selected range' },
  { label: 'HTTP metrics', value: QueryTypes.HTTPMetrics, description: 'DNS, TCP, SSL and response timings' },
  { label: 'Ping metrics', value: QueryTypes.PingMetrics, description: 'Round trip times and packet loss' },
  { label: 'TCP metrics', value: QueryTypes.TCPMetrics, description: 'Time to connect and uptime' },
  { label: 'Lighthouse', value: QueryTypes.Lighthouse, description: 'Lighthouse performance scores' },
];

const groupByOptions: Array<ComboboxOption<GroupBy>> = [
  { label: 'Hour', value: 'hour' },
  { label: 'Day', value: 'day' },
  { label: 'Month', value: 'month' },
];

const ALL_MONITORS = 0;
const allMonitorsOption: ComboboxOption<number> = { label: 'All monitors', value: ALL_MONITORS };

const usesMonitor = (queryType?: OhDearQueryType) => queryType !== QueryTypes.Monitors;

const usesGroupBy = (queryType?: OhDearQueryType) =>
  queryType === QueryTypes.Uptime ||
  queryType === QueryTypes.HTTPMetrics ||
  queryType === QueryTypes.PingMetrics ||
  queryType === QueryTypes.TCPMetrics;

export function QueryEditor({ query, datasource, onChange, onRunQuery }: Props) {
  const queryType = query.queryType ?? DEFAULT_QUERY.queryType;
  const [monitors, setMonitors] = useState<Array<ComboboxOption<number>>>([allMonitorsOption]);
  const [loadingMonitors, setLoadingMonitors] = useState(true);
  const [monitorError, setMonitorError] = useState<string>();

  useEffect(() => {
    let cancelled = false;

    datasource
      .getMonitors()
      .then((result) => {
        if (cancelled) {
          return;
        }
        setMonitorError(undefined);
        setMonitors([
          allMonitorsOption,
          ...result.map((monitor) => ({
            label: monitor.label || monitor.url,
            value: monitor.id,
            description: monitor.url,
          })),
        ]);
      })
      .catch((error) => {
        if (!cancelled) {
          setMonitorError(error?.data?.error ?? error?.statusText ?? 'Could not load monitors from Oh Dear');
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoadingMonitors(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [datasource]);

  useEffect(() => {
    if (!query.queryType) {
      onChange({ ...DEFAULT_QUERY, ...query, queryType });
    }
  }, [query, queryType, onChange]);

  const onQueryTypeChange = (option: ComboboxOption<OhDearQueryType>) => {
    onChange({ ...query, queryType: option.value });
    onRunQuery();
  };

  const onMonitorChange = (option: ComboboxOption<number>) => {
    onChange({ ...query, monitorId: option.value });
    onRunQuery();
  };

  const onGroupByChange = (option: ComboboxOption<GroupBy>) => {
    onChange({ ...query, groupBy: option.value });
    onRunQuery();
  };

  return (
    <div aria-busy={loadingMonitors}>
      <Stack direction="column" gap={1}>
        <Stack gap={0} wrap="wrap">
          <InlineField label="Query type" labelWidth={16}>
            <Combobox
              id="query-editor-query-type"
              options={queryTypeOptions}
              value={queryType ?? null}
              onChange={onQueryTypeChange}
              width={28}
            />
          </InlineField>
          {usesMonitor(queryType) && (
            <InlineField label="Monitor" labelWidth={16} tooltip="Leave as All monitors to query every monitor">
              <Combobox
                id="query-editor-monitor"
                options={monitors}
                value={query.monitorId ?? ALL_MONITORS}
                onChange={onMonitorChange}
                loading={loadingMonitors}
                width={32}
              />
            </InlineField>
          )}
          {usesGroupBy(queryType) && (
            <InlineField label="Group by" labelWidth={16}>
              <Combobox
                id="query-editor-group-by"
                options={groupByOptions}
                value={query.groupBy ?? 'hour'}
                onChange={onGroupByChange}
                width={16}
              />
            </InlineField>
          )}
        </Stack>
        {monitorError && (
          <Alert title="Oh Dear" severity="warning">
            {monitorError}
          </Alert>
        )}
      </Stack>
    </div>
  );
}
