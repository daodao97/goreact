import './polyfill';
import React from "react";
import { renderToString } from "react-dom/server";
import { MyTheme } from "../layout/Theme";

type ComponentProps = Record<string, any>;

interface RenderOptions {
    Component: React.ComponentType<any>;
    transformProps?: (globalProps: any) => ComponentProps;
}

export function createServerRenderer({ Component, transformProps }: RenderOptions) {
    return function Render() {
        const defaultTransform = (globalProps: any) => globalProps || {};
        const propsTransformer = transformProps || defaultTransform;

        const props = propsTransformer(window.INITIAL_PROPS);

        const userInfo = window.USER_INFO || null;
        const propsWithUser = {
            ...props,
            userInfo
        };

        return renderToString(
            <MyTheme>
                <Component {...propsWithUser} />
            </MyTheme>
        );
    };
} 