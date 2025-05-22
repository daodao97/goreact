import { get } from "lodash-es";

/**
 * 国际化翻译函数
 * @param path 翻译键路径
 * @param defaultValue 默认值
 * @returns 翻译后的文本
 */
export function t(path: string, defaultValue: string): string {
    return get(window.TRANSLATIONS, path, defaultValue || path);
}

/**
 * 获取当前网站配置
 * @returns Website
 */
export function getWebsite() {
    return window.WEBSITE;
}


export function getUrlWithLang(path: string) {
    // 如果是完整的URL（以http://或https://开头），直接返回不修改
    if (path.startsWith('http://') || path.startsWith('https://')) {
        return path;
    }

    const lang = window.LANG;
    const defaultLang = getWebsite()?.Lang;

    // 确保路径以/开头
    const normalizedPath = path.startsWith('/') ? path : `/${path}`;

    if (lang === defaultLang) {
        return normalizedPath;
    }

    return `/${lang}${normalizedPath}`;
}

export function matchPath(path: string, currentPath: string): boolean {
    const lang = window.LANG;
    currentPath = currentPath.replace(`/${lang}`, '');
    if (currentPath === "") {
        currentPath = "/";
    }
    return currentPath === path;
}

export function getTranslations<T>(path: string, defaultValue: T): T {
    return get(window.TRANSLATIONS, path, defaultValue);
}