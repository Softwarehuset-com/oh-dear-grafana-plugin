import React from 'react';
import { render, screen } from '@testing-library/react';
import { QueryEditor } from './QueryEditor';
import type { DataSource } from '../datasource';
import { OhDearQuery, QueryTypes } from '../types';

const setup = (query: Partial<OhDearQuery> = {}) => {
  const onChange = jest.fn();
  const onRunQuery = jest.fn();
  const getMonitors = jest.fn().mockResolvedValue([]);

  render(
    <QueryEditor
      query={{ refId: 'A', queryType: QueryTypes.Uptime, ...query }}
      datasource={{ getMonitors } as unknown as DataSource}
      onChange={onChange}
      onRunQuery={onRunQuery}
    />
  );

  return { onChange, onRunQuery, getMonitors };
};

describe('QueryEditor', () => {
  it('shows the monitor and group by pickers for uptime queries', () => {
    const { getMonitors } = setup();

    expect(screen.getByRole('combobox', { name: 'Query type' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Monitor' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Group by' })).toBeInTheDocument();
    expect(getMonitors).toHaveBeenCalledTimes(1);
  });

  it('hides the monitor and group by pickers for monitor queries', () => {
    setup({ queryType: QueryTypes.Monitors });

    expect(screen.queryByRole('combobox', { name: 'Monitor' })).not.toBeInTheDocument();
    expect(screen.queryByRole('combobox', { name: 'Group by' })).not.toBeInTheDocument();
  });

  it('hides the group by picker for downtime queries', () => {
    setup({ queryType: QueryTypes.Downtime });

    expect(screen.getByRole('combobox', { name: 'Monitor' })).toBeInTheDocument();
    expect(screen.queryByRole('combobox', { name: 'Group by' })).not.toBeInTheDocument();
  });

  it('defaults the pickers to all monitors grouped by hour', () => {
    setup();

    expect(screen.getByDisplayValue('All monitors')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Hour')).toBeInTheDocument();
  });

  it('hides the group by picker for lighthouse queries', () => {
    setup({ queryType: QueryTypes.Lighthouse });

    expect(screen.getByRole('combobox', { name: 'Monitor' })).toBeInTheDocument();
    expect(screen.queryByRole('combobox', { name: 'Group by' })).not.toBeInTheDocument();
  });
});
