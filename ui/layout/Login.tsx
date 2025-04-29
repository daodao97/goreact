import React, { useState, useEffect } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { Cross2Icon } from '@radix-ui/react-icons';
import { getWebsite, getTranslations, t } from '../lib/i18n';
import { saveUserInfo } from '../lib/auth';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { GoogleLogin } from '@react-oauth/google';

// 登录模态框组件
export function LoginModal({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {

    const website = getWebsite();
    const authProviders = website.AuthProvider || [];

    // 查找mail类型的登录方式
    const mailProvider = authProviders.find(provider => provider.Provider === 'mail');
    // 其他第三方登录方式
    const otherProviders = authProviders.filter(provider => provider.Provider !== 'mail');

    const title = getTranslations('root.login.title', '');
    const description = getTranslations('root.login.description', '');

    return (
        <Dialog.Root open={open} onOpenChange={onOpenChange}>
            <Dialog.Portal>
                <Dialog.Overlay className="fixed inset-0 bg-black/50 data-[state=open]:animate-overlayShow" />
                <Dialog.Content className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg p-6 shadow-lg max-w-md w-full data-[state=open]:animate-contentShow">
                    {title && <Dialog.Title className="text-xl font-bold mb-4">{title}</Dialog.Title>}
                    {description && <Dialog.Description className="text-sm text-gray-500 mb-4">{description}</Dialog.Description>}

                    <div className="space-y-4">
                        {/* 优先显示mail登录 */}
                        {mailProvider && <MailLogin key={mailProvider.Provider} data={mailProvider} />}

                        {/* 如果有其他登录方式，显示分隔线和其他登录方式 */}
                        {otherProviders.length > 0 && (
                            <>
                                <div className="relative my-4">
                                    <div className="absolute inset-0 flex items-center">
                                        <div className="w-full border-t border-gray-300" />
                                    </div>
                                    <div className="relative flex justify-center text-sm">
                                        <span className="px-2 bg-white text-gray-500">{t('root.login.orLoginWith', '或通过以下方式登录')}</span>
                                    </div>
                                </div>

                                {/* 显示其他登录方式 */}
                                {otherProviders.map((provider) => {
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
                            </>
                        )}
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

// mail登录组件
function MailLogin({ data }: { data: any }) {
    const [mode, setMode] = useState<'login' | 'register'>('login');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [verificationCode, setVerificationCode] = useState('');
    const [countdown, setCountdown] = useState(0);
    const [loading, setLoading] = useState(false);
    const [errorMsg, setErrorMsg] = useState('');
    const [verificationErrorMsg, setVerificationErrorMsg] = useState('');

    // 清除倒计时定时器
    useEffect(() => {
        let timer: ReturnType<typeof setTimeout> | null = null;

        if (countdown > 0) {
            timer = setInterval(() => {
                setCountdown(prev => {
                    if (prev <= 1) {
                        return 0;
                    }
                    return prev - 1;
                });
            }, 1000);
        }

        // 清除副作用
        return () => {
            if (timer) clearInterval(timer);
        };
    }, [countdown]);

    // 处理获取验证码
    const handleGetVerificationCode = () => {
        if (countdown > 0 || !email) return;

        // 清除之前的错误消息
        setVerificationErrorMsg('');

        // 简单的邮箱格式验证
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            setVerificationErrorMsg(t('root.login.emailInvalid', '请输入有效的邮箱地址'));
            return;
        }

        // 设置加载状态
        const sendingBtnText = t('root.login.sending', '发送中...');
        const originalBtnText = t('root.login.getVerificationCode', '获取验证码');
        const btn = document.querySelector('.verification-code-btn') as HTMLButtonElement;
        if (btn) btn.innerText = sendingBtnText;

        // 发送获取验证码请求
        fetch('/login/send-verification-code', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email }),
        })
            .then(res => res.json())
            .then(data => {
                console.log("data", data)
                if (data.message === "success") {
                    // 开始倒计时
                    setCountdown(60);
                } else {
                    // 恢复按钮文字
                    if (btn) btn.innerText = originalBtnText;
                    setVerificationErrorMsg(data.message || t('root.login.sendVerificationFailed', '验证码发送失败'));
                }
            })
            .catch(err => {
                console.error(err);
                // 恢复按钮文字
                if (btn) btn.innerText = originalBtnText;
                setVerificationErrorMsg(t('root.login.networkError', '网络错误，请稍后重试'));
            });
    };

    // 合并处理提交（登录或注册）
    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        // 清除错误消息
        setErrorMsg('');

        // 基本验证
        if (!email) {
            setErrorMsg(t('root.login.emailRequired', '请输入邮箱地址'));
            return;
        }

        // 邮箱格式验证
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            setErrorMsg(t('root.login.emailInvalid', '请输入有效的邮箱地址'));
            return;
        }

        if (!password) {
            setErrorMsg(t('root.login.passwordRequired', '请输入密码'));
            return;
        }

        // 密码长度验证
        if (password.length < 8 || password.length > 16) {
            setErrorMsg(t('root.login.passwordLengthInvalid', '密码长度必须在8-16个字符之间'));
            return;
        }

        if (mode === 'register' && !verificationCode) {
            setErrorMsg(t('root.login.verificationCodeRequired', '请输入验证码'));
            return;
        }

        console.log("handleSubmit", email, password, verificationCode, mode)
        if (loading) return;

        setLoading(true);
        fetch('/login/mail', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(
                mode === 'register'
                    ? { email, password, verificationCode, mode: 'register' }
                    : { email, password, mode: 'login' }
            ),
        })
            .then(res => {
                // 检查HTTP状态码
                if (res.status === 200) {
                    return res.json().then(data => {
                        setLoading(false);
                        saveUserInfo(data);
                        window.location.reload();
                    });
                } else {
                    return res.json().then(data => {
                        setLoading(false);
                        const msg = mode === 'register'
                            ? t('root.login.registerFailed', '注册失败')
                            : t('root.login.loginFailed', '登录失败');
                        setErrorMsg(data.message || msg);
                    });
                }
            })
            .catch(err => {
                setLoading(false);
                console.error(err);
                setErrorMsg(t('root.login.networkError', '网络错误，请稍后重试'));
            });
    };

    return (
        <div className="w-full">
            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                        {t('root.login.email', '邮箱')}
                    </label>
                    <input
                        id="email"
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                        placeholder={t('root.login.emailPlaceholder', '请输入邮箱')}
                        required
                    />
                </div>

                <div>
                    <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                        {t('root.login.password', '密码')}
                    </label>
                    <input
                        id="password"
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                        placeholder={t('root.login.passwordPlaceholder', '请输入密码')}
                        required
                    />
                </div>

                {mode === 'register' && (
                    <div>
                        <label htmlFor="verificationCode" className="block text-sm font-medium text-gray-700 mb-1">
                            {t('root.login.verificationCode', '验证码')}
                        </label>
                        <div className="flex">
                            <input
                                id="verificationCode"
                                type="text"
                                value={verificationCode}
                                onChange={(e) => setVerificationCode(e.target.value)}
                                className="flex-1 px-3 py-2 border border-gray-300 rounded-l-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                                placeholder={t('root.login.verificationCodePlaceholder', '请输入验证码')}
                                required
                            />
                            <button
                                type="button"
                                onClick={handleGetVerificationCode}
                                disabled={countdown > 0}
                                className={`verification-code-btn px-4 py-2 text-sm font-medium text-white rounded-r-md focus:outline-none ${countdown > 0 ? 'bg-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700'
                                    }`}
                            >
                                {countdown > 0 ? `${countdown}秒后重试` : t('root.login.getVerificationCode', '获取验证码')}
                            </button>
                        </div>
                        {verificationErrorMsg && (
                            <p className="mt-1 text-sm text-red-600">{verificationErrorMsg}</p>
                        )}
                    </div>
                )}

                {/* 全局错误信息显示 */}
                {errorMsg && (
                    <div className="py-2 px-3 bg-red-50 border border-red-200 rounded-md">
                        <p className="text-sm text-red-600">{errorMsg}</p>
                    </div>
                )}

                <button
                    type="submit"
                    disabled={loading}
                    onClick={handleSubmit}
                    className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-blue-400 disabled:cursor-not-allowed"
                >
                    {loading ? (
                        <span>{t('root.login.processing', '处理中...')}</span>
                    ) : mode === 'register' ? (
                        t('root.login.registerBtn', '注册')
                    ) : (
                        t('root.login.loginBtn', '登录')
                    )}
                </button>

                <div className="text-center text-sm">
                    <span className="text-gray-600">
                        {mode === 'login' ? t('root.login.noAccount', '还没有账号？') : t('root.login.hasAccount', '已有账号？')}
                    </span>
                    <button
                        type="button"
                        onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
                        className="ml-1 text-blue-600 hover:text-blue-800 focus:outline-none"
                    >
                        {mode === 'login' ? t('root.login.registerBtn', '注册') : t('root.login.loginBtn', '登录')}
                    </button>
                </div>
            </form>
        </div>
    );
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