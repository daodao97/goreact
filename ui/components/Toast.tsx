import React, { createContext, useState, useContext, useEffect } from 'react';
import * as ToastPrimitive from '@radix-ui/react-toast';

type ToastType = 'success' | 'error' | 'info';
type ToastVerticalPosition = 'top' | 'center' | 'bottom';
type ToastHorizontalPosition = 'left' | 'center' | 'right';

interface ToastContextType {
    showToast: (title: string, description: string, options?: {
        type?: ToastType;
        verticalPosition?: ToastVerticalPosition;
        horizontalPosition?: ToastHorizontalPosition;
    }) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

export const useToast = () => {
    const context = useContext(ToastContext);
    if (!context) {
        throw new Error('useToast must be used within a ToastProvider');
    }
    return context;
};

export function ToastProvider({ children }: { children: React.ReactNode }) {
    const [toast, setToast] = useState<{
        open: boolean;
        title: string;
        description: string;
        type: ToastType;
        verticalPosition: ToastVerticalPosition;
        horizontalPosition: ToastHorizontalPosition;
    }>({
        open: false,
        title: '',
        description: '',
        type: 'info',
        verticalPosition: 'top',
        horizontalPosition: 'right',
    });

    const showToast = (
        title: string,
        description: string,
        options?: {
            type?: ToastType;
            verticalPosition?: ToastVerticalPosition;
            horizontalPosition?: ToastHorizontalPosition;
        }
    ) => {
        setToast({
            open: true,
            title,
            description,
            type: options?.type || 'info',
            verticalPosition: options?.verticalPosition || 'top',
            horizontalPosition: options?.horizontalPosition || 'right',
        });
        setTimeout(() => setToast((prev) => ({ ...prev, open: false })), 5000);
    };

    // 为全局fetchApi提供showToast方法
    useEffect(() => {
        window.__showToast = (title: string, description: string, type: ToastType = 'info') => {
            showToast(title, description, { type });
        };
        return () => {
            delete window.__showToast;
        };
    }, []);

    // 根据位置生成定位样式
    const getPositionClasses = (): string => {
        let classes = 'fixed ';

        // 垂直位置
        switch (toast.verticalPosition) {
            case 'top':
                classes += 'top-4 ';
                break;
            case 'center':
                classes += 'top-1/2 -translate-y-1/2 ';
                break;
            case 'bottom':
            default:
                classes += 'bottom-4 ';
                break;
        }

        // 水平位置
        switch (toast.horizontalPosition) {
            case 'left':
                classes += 'left-4 ';
                break;
            case 'center':
                classes += 'left-1/2 -translate-x-1/2 ';
                break;
            case 'right':
            default:
                classes += 'right-4 ';
                break;
        }

        return classes;
    };

    return (
        <ToastContext.Provider value={{ showToast }}>
            <ToastPrimitive.Provider swipeDirection="right">
                {children}
                <ToastPrimitive.Root
                    className={`${getPositionClasses()} p-4 rounded-lg shadow-lg max-w-sm ${toast.type === 'error'
                        ? 'bg-red-50 border-l-4 border-red-500'
                        : toast.type === 'success'
                            ? 'bg-green-50 border-l-4 border-green-500'
                            : 'bg-blue-50 border-l-4 border-blue-500'
                        } transition-all transform ${toast.open ? 'opacity-100' : 'opacity-0 pointer-events-none'
                        } z-50`}
                    open={toast.open}
                    onOpenChange={(open) => setToast((prev) => ({ ...prev, open }))}
                >
                    <ToastPrimitive.Title
                        className={`font-medium mb-1 ${toast.type === 'error'
                            ? 'text-red-800'
                            : toast.type === 'success'
                                ? 'text-green-800'
                                : 'text-blue-800'
                            }`}
                    >
                        {toast.title}
                    </ToastPrimitive.Title>
                    <ToastPrimitive.Description
                        className={`text-sm ${toast.type === 'error'
                            ? 'text-red-600'
                            : toast.type === 'success'
                                ? 'text-green-600'
                                : 'text-blue-600'
                            }`}
                    >
                        {toast.description}
                    </ToastPrimitive.Description>
                    <ToastPrimitive.Close className="absolute top-2 right-2 p-1 rounded-full hover:bg-black/5">
                        <svg
                            width="14"
                            height="14"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                        >
                            <path
                                d="M12 4L4 12M4 4L12 12"
                                stroke="currentColor"
                                strokeWidth="1.5"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                            />
                        </svg>
                    </ToastPrimitive.Close>
                </ToastPrimitive.Root>
                <ToastPrimitive.Viewport />
            </ToastPrimitive.Provider>
        </ToastContext.Provider>
    );
} 