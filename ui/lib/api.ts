// 封装全局API请求工具

// 定义请求选项类型
export interface RequestOptions extends RequestInit {
    showError?: boolean;
}

// 确保Window中定义了__showToast方法
declare global {
    interface Window {
        __showToast?: (title: string, description: string, type: 'success' | 'error' | 'info') => void;
    }
}

/**
 * 封装的fetch API，提供统一的错误处理和请求配置
 * @param url 请求地址
 * @param options 请求选项
 * @returns 返回解析后的JSON数据
 */
export const fetchApi = async <T extends unknown>(url: string, options: RequestOptions = {}): Promise<T> => {
    const { showError = true, ...fetchOptions } = options;

    try {
        const response = await fetch(url, {
            headers: {
                'Content-Type': 'application/json',
                ...fetchOptions.headers,
            },
            ...fetchOptions,
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: '请求失败' }));
            const error = new Error(errorData.message || `请求失败: ${response.status}`);
            throw error;
        }

        return await response.json() as T;
    } catch (error: any) {
        throw error;
    }
};

/**
 * 显示成功消息
 * @param message 成功消息
 */
export const showSuccess = (message: string) => {
    if (window.__showToast) {
        window.__showToast('成功', message, 'success');
    }
};

/**
 * 显示错误消息
 * @param message 错误消息
 */
export const showError = (message: string) => {
    if (window.__showToast) {
        window.__showToast('错误', message, 'error');
    }
};

/**
 * 显示信息提示
 * @param message 信息内容
 */
export const showInfo = (message: string) => {
    if (window.__showToast) {
        window.__showToast('提示', message, 'info');
    }
}; 