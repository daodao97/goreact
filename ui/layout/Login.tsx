import React, { useState, useEffect } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { Cross2Icon } from '@radix-ui/react-icons';
import { getWebsite, t } from '../lib/i18n';
import { saveUserInfo } from '../lib/auth';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { GoogleLogin } from '@react-oauth/google';

// 登录模态框组件
export function LoginModal({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {

    const website = getWebsite();
    const authProviders = website.AuthProvider || [];

    return (
        <Dialog.Root open={open} onOpenChange={onOpenChange}>
            <Dialog.Portal>
                <Dialog.Overlay className="fixed inset-0 bg-black/50 data-[state=open]:animate-overlayShow" />
                <Dialog.Content className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg p-6 shadow-lg max-w-md w-full data-[state=open]:animate-contentShow">
                    <Dialog.Title className="text-xl font-bold mb-4">{t('root.login.title', 'Login')}</Dialog.Title>
                    <Dialog.Description className="text-sm text-gray-500 mb-4">
                        {t('root.login.description', 'Please select the following methods to login to your account')}
                    </Dialog.Description>

                    <div className="space-y-4">
                        {authProviders.map((provider) => {
                            if (provider.Provider === 'github') {
                                return <GithubLogin key={provider.Provider} data={provider} />;
                            } else if (provider.Provider === 'google') {
                                return <MyGoogleLogin key={provider.Provider} data={provider} />;
                            } else {
                                return (
                                    <button
                                        key={provider.Provider}
                                        type="button"
                                        className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                                    >
                                        {provider.Provider}
                                    </button>
                                );
                            }
                        })}
                    </div>

                    <Dialog.Close asChild>
                        <button
                            className="absolute top-4 right-4 inline-flex items-center justify-center rounded-full h-6 w-6 hover:bg-gray-200 focus:outline-none"
                            aria-label="关闭"
                        >
                            <Cross2Icon />
                        </button>
                    </Dialog.Close>
                </Dialog.Content>
            </Dialog.Portal>
        </Dialog.Root>
    );
}

// github登录组件
function GithubLogin({ data }: { data: any }) {
    return (
        <a
            href={`https://github.com/login/oauth/authorize?client_id=${data.ClientID}&redirect_uri=${data.CallbackURL}&scope=user`}
            className="w-full flex items-center justify-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
        >
            <svg className="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.87 1.52 2.34 1.07 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.92 0-1.11.38-2 1.03-2.71-.1-.25-.45-1.29.1-2.64 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.35.2 2.39.1 2.64.65.71 1.03 1.6 1.03 2.71 0 3.82-2.34 4.66-4.57 4.91.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0012 2z"></path>
            </svg>
            {t('root.login.github', 'Login with GitHub')}
        </a>
    );
}

// google登录组件
function MyGoogleLogin({ data }: { data: any }) {
    const onSuccess = (credentialResponse: any) => {
        console.log("credentialResponse", credentialResponse)
        fetch('/login/google', {
            method: 'POST',
            body: JSON.stringify(credentialResponse),
        }).then(res => {
            if (res.redirected) {
                window.location.href = res.url;
                return;
            }
            return res.json();
        }).then(data => {
            if (data) {
                saveUserInfo(data);
                window.location.reload();
            }
        });
    }
    const onError = () => {
        console.log("error login")
    }
    return <GoogleOAuthProvider clientId={data.ClientID}>
        <GoogleLogin onSuccess={onSuccess} onError={onError} />
    </GoogleOAuthProvider>
}

// 导出打开登录模态框的方法
let setLoginModalOpen: React.Dispatch<React.SetStateAction<boolean>> | null = null;

export function showLoginModal() {
    if (setLoginModalOpen) {
        setLoginModalOpen(true);
    } else {
        console.warn('登录模态框尚未初始化');
    }
}

// 登录模态框容器组件，用于在应用中初始化模态框
export function LoginModalContainer() {
    const [open, setOpen] = useState(false);

    // 保存 setOpen 函数的引用，以便可以从外部调用
    setLoginModalOpen = setOpen;

    return <LoginModal open={open} onOpenChange={setOpen} />;
}