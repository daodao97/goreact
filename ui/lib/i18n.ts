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
    const lang = window.LANG;
    const defaultLang = getWebsite()?.Lang;
    if (lang === defaultLang) {
        return path;
    }
    return `/${lang}${path}`;
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