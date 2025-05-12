
import React from "react";
import { twMerge } from "tailwind-merge";
import { ToastProvider } from "../components/Toast";

export default function Layout({ children, className }: { children: React.ReactNode, className?: string }) {
    return (
        <ToastProvider>
            <div className={twMerge(
                "max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-12",
                className
            )}>
                {children}
            </div>
        </ToastProvider>
    );
}
