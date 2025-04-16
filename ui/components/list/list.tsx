import React from 'react';
import Pagination from './pagination';
import clsx from 'clsx';
import { isAuthenticated } from '@/core/lib/auth';
import { showLoginModal } from '@/core/layout/Login';

// 合并类名的工具函数
const cn = (...classes) => {
    return clsx(classes);
};

interface Record {
    [key: string]: any;
}

interface TableData {
    h2: string;
    total: number;
    page: number;
    data: Record[];
}

interface SelectOption {
    label: string;
    value: string;
    class?: string;
}

interface SelectFilter {
    label: string;
    field: string;
    value: string;
    options: SelectOption[];
}

interface Header {
    label: string;
    field: string;
    class?: string;
    cellClass?: string;
    render?: (fields: Header[], index: number, field: Header, row: Record) => React.ReactNode;
}

interface TableHeaderData {
    h2: string;
    total: number;
    page: number;
}


// 过滤器组件
const FilterComponent: React.FC<{ filters: SelectFilter[] }> = ({ filters }) => {
    const getQueryValue = (field: string): string => {
        const params = new URLSearchParams(window.location.search);
        return params.get(field) || '';
    };

    const buildUrl = (field: string, value: string | null): string => {
        const params = new URLSearchParams(window.location.search);
        if (value === null) {
            params.delete(field);
        } else {
            params.set(field, value);
        }
        return `${window.location.pathname}?${params.toString()}`;
    };

    return (
        <>
            {filters.map((filter, filterIndex) => (
                <React.Fragment key={filterIndex}>
                    {filter.options.map((option, optionIndex) => {
                        const currentValue = getQueryValue(filter.field);
                        const isSelected = option.value === currentValue;
                        const jumpUrl = isSelected ? buildUrl(filter.field, null) : buildUrl(filter.field, option.value);

                        return (
                            <a
                                key={`${filterIndex}-${optionIndex}`}
                                href={jumpUrl}
                                className={cn(
                                    option.class || '',
                                    'px-3 py-1 text-sm rounded-full transition-colors mb-2 sm:mb-0 flex items-center gap-2',
                                    isSelected ? 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200' : ''
                                )}
                            >
                                {option.label}
                                {isSelected && (
                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                )}
                            </a>
                        );
                    })}
                </React.Fragment>
            ))}
        </>
    );
};

// 表格行组件
const TableRow: React.FC<{ schema: Header[]; row: Record }> = ({ schema, row }) => {
    return (
        <tr className="group hover:bg-gray-50 dark:hover:bg-gray-700">
            {schema.map((item, index) => (
                <td
                    key={index}
                    className={cn(
                        item.cellClass || '',
                        'px-3 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300'
                    )}
                >
                    {item.render ? (
                        item.render(schema, index, item, row)
                    ) : (
                        row[item.field]
                    )}
                </td>
            ))}
        </tr>
    );
};

// 表格组件
const Table: React.FC<{
    schema: Header[];
    data: TableHeaderData;
    filters: SelectFilter[];
    children: React.ReactNode;
}> = ({ schema, data, filters, children }) => {
    return (
        <div className="px-4 sm:px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="flex items-center space-x-4">
                    <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                        {data.h2}
                        {/* <span className="text-sm text-gray-500 ml-2">
                            ({`共 ${data.total} 条记录，第 ${data.page} 页`})
                        </span> */}
                    </h2>
                </div>
                <div className="flex items-center gap-2 flex-wrap">
                    <FilterComponent filters={filters} />
                </div>
            </div>

            <div className="mt-4 -mx-4 sm:mx-0 overflow-x-auto">
                <div className="inline-block min-w-full align-middle">
                    <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                        <thead className="bg-gray-50 dark:bg-gray-700">
                            <tr>
                                {schema.map((header, index) => (
                                    <th
                                        key={index}
                                        scope="col"
                                        className={cn(
                                            header.class || '',
                                            'px-3 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider'
                                        )}
                                    >
                                        {header.label}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                            {children}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

// 主数据表格组件
const DataTable: React.FC<{
    schema: Header[];
    data: TableData;
    filters: SelectFilter[];
    minNoNeedLoginPage: number;
}> = ({ schema, data, filters, minNoNeedLoginPage = -1 }) => {
    return (
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm overflow-hidden">
            <Table
                schema={schema}
                data={{
                    h2: data.h2,
                    total: data.total,
                    page: data.page,
                }}
                filters={filters}
            >
                {data.data.map((record, index) => (
                    <TableRow key={index} schema={schema} row={record} />
                ))}
            </Table>

            <Pagination
                currentPage={data.page}
                totalPages={Math.ceil(data.total / 20)} // 每页20条数据
                onPageChange={(page) => {
                    console.log({ minNoNeedLoginPage, page, isAuthenticated: isAuthenticated(), tag: minNoNeedLoginPage > 0 && !isAuthenticated() && page > minNoNeedLoginPage });
                    if (minNoNeedLoginPage > 0 && !isAuthenticated() && page > minNoNeedLoginPage) {
                        showLoginModal();
                        return;
                    }
                    // 处理页码变化
                    const params = new URLSearchParams(window.location.search);
                    params.set('page', page.toString());
                    window.location.href = `${window.location.pathname}?${params.toString()}`;
                }}
            />
        </div>
    );
};

export default DataTable;
export type { TableData, SelectFilter, SelectOption, Header };
