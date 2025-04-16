import React from 'react';

interface PaginationProps {
    currentPage: number;
    totalPages: number;
    currentTag?: string;
    onPageChange: (page: number) => void;
}

// 获取要显示的页码范围
const getPageRange = (current: number, total: number): number[] => {
    if (total <= 7) {
        // 如果总页数小于等于7，显示所有页码
        return Array.from({ length: total }, (_, i) => i + 1);
    }

    // 否则使用省略号显示
    const pages: number[] = [1];

    if (current <= 3) {
        // 当前页靠近开始
        pages.push(2, 3, 4, -1, total);
    } else if (current >= total - 2) {
        // 当前页靠近结束
        pages.push(-1, total - 3, total - 2, total - 1, total);
    } else {
        // 当前页在中间
        pages.push(-1, current - 1, current, current + 1, -1, total);
    }

    return pages;
};

const PrevButton: React.FC<{ currentPage: number; onClick: () => void; disabled?: boolean }> = ({
    currentPage,
    onClick,
    disabled
}) => {
    if (disabled) {
        return (
            <button disabled
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-400 bg-gray-50 cursor-not-allowed dark:bg-gray-800 dark:border-gray-700 dark:text-gray-500">
                <svg className="mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
                </svg>
                上一页
            </button>
        );
    }

    return (
        <button onClick={onClick}
            className="relative inline-flex items-center px-4 py-2 border border-amber-300 text-sm font-medium rounded-md text-amber-700 bg-white hover:bg-amber-50 transition-colors dark:bg-gray-800 dark:text-amber-400 dark:hover:bg-gray-700">
            <svg className="mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
            </svg>
            上一页
        </button>
    );
};

const NextButton: React.FC<{ currentPage: number; onClick: () => void; disabled?: boolean }> = ({
    currentPage,
    onClick,
    disabled
}) => {
    if (disabled) {
        return (
            <button disabled
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-400 bg-gray-50 cursor-not-allowed dark:bg-gray-800 dark:border-gray-700 dark:text-gray-500">
                下一页
                <svg className="ml-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                </svg>
            </button>
        );
    }

    return (
        <button onClick={onClick}
            className="relative inline-flex items-center px-4 py-2 border border-amber-300 text-sm font-medium rounded-md text-amber-700 bg-white hover:bg-amber-50 transition-colors dark:bg-gray-800 dark:text-amber-400 dark:hover:bg-gray-700">
            下一页
            <svg className="ml-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
            </svg>
        </button>
    );
};

const Pagination: React.FC<PaginationProps> = ({
    currentPage,
    totalPages,
    onPageChange
}) => {
    return (
        <div className="px-3 sm:px-6 py-4 flex items-center justify-between border-t border-gray-200 dark:border-gray-700">
            <div className="flex-1 flex justify-between sm:justify-center gap-4">
                <PrevButton
                    currentPage={currentPage}
                    onClick={() => onPageChange(currentPage - 1)}
                    disabled={currentPage <= 1}
                />

                {/* 页码显示 (仅在中等屏幕及以上显示) */}
                <div className="hidden sm:flex items-center gap-2">
                    {getPageRange(currentPage, totalPages).map((pageNum, index) => {
                        if (pageNum === -1) {
                            return (
                                <span key={`ellipsis-${index}`} className="px-2 text-gray-500 dark:text-gray-400">
                                    ...
                                </span>
                            );
                        }

                        if (pageNum === currentPage) {
                            return (
                                <span
                                    key={pageNum}
                                    className="relative inline-flex items-center px-4 py-2 border border-amber-300 text-sm font-medium rounded-md bg-amber-50 text-amber-700 dark:bg-amber-900 dark:text-amber-200"
                                >
                                    {pageNum}
                                </span>
                            );
                        }

                        return (
                            <button
                                key={pageNum}
                                onClick={() => onPageChange(pageNum)}
                                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 dark:border-gray-600"
                            >
                                {pageNum}
                            </button>
                        );
                    })}
                </div>

                <NextButton
                    currentPage={currentPage}
                    onClick={() => onPageChange(currentPage + 1)}
                    disabled={currentPage >= totalPages}
                />
            </div>
        </div>
    );
};

export default Pagination;
