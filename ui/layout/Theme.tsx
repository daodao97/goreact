import { Theme, ThemeProps } from "@radix-ui/themes";
import React, { useEffect } from "react";
import { Header } from "./Header";
import { Footer } from "./Footer";
import { LoginModalContainer } from "./Login";

const defaultThemeConfig: ThemeProps = {
    accentColor: "indigo",
    grayColor: "slate",
    radius: "medium",
    scaling: "100%",
    panelBackground: "translucent",
    hasBackground: true
}

export function MyTheme({
    children,
    ...themeConfig
}: {
    children: React.ReactNode;
} & ThemeProps) {
    if (themeConfig === undefined) {
        themeConfig = defaultThemeConfig;
    }
    return (
        <Theme
            {...themeConfig}
        >
            <LoginModalContainer />
            <div className="flex flex-col min-h-screen h-screen">
                <Header />
                <main className="flex-grow bg-gray-50 dark:bg-gray-900">
                    {children}
                </main>
                <Footer />
            </div>
        </Theme>
    );
}
