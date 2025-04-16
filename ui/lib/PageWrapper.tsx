import React, { ComponentType } from "react";
import { hydrateRoot } from "react-dom/client";
import { MyTheme } from "../layout/Theme";

interface PageWrapperProps {
    Component: ComponentType<any>;
    containerId?: string;
    errorHandler?: (error: Error) => void;
}

/**
 * 统一的页面包装器组件
 * 
 * @param Component 要渲染的页面组件
 * @param containerId 容器ID，默认为"react-app"
 * @param errorHandler 错误处理函数
 */
export function renderPage<T>({
    Component,
    containerId = "react-app",
    errorHandler = (error: Error) => console.error("Root render failed:", error)
}: PageWrapperProps): void {
    const container = document.getElementById(containerId);

    if (!container) {
        console.error(`React app container with id "${containerId}" not found`);
        return;
    }

    const props = (window as any).INITIAL_PROPS || {};

    try {
        hydrateRoot(
            container,
            <MyTheme>
                <Component {...props} />
            </MyTheme>
        );
    } catch (error) {
        errorHandler(error as Error);
    }
} 