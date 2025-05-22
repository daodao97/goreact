import React, { useState, useEffect } from 'react';
import { DropdownMenu, Button } from '@radix-ui/themes';
import { showLoginModal } from './Login';
import { getTranslations, t, getUrlWithLang, matchPath, getWebsite } from '../lib/i18n';
import { getUser, isAuthenticated, logout } from '../lib/auth';
import { HamburgerMenuIcon } from '@radix-ui/react-icons';

// 语言切换功能
const switchLang = (lang: string, supportedLangs: string[]) => {
    // 更新 HTML lang 属性
    document.documentElement.lang = lang;

    // 重定向到相应语言的页面
    const currentPath = window.location.pathname;
    const pathSegments = currentPath.split('/').filter(Boolean);

    // 移除路径中可能存在的语言代码
    if (pathSegments.length > 0 && supportedLangs.includes(pathSegments[0])) {
        pathSegments.shift();
    }

    // 构建新的 URL 路径，去除末尾的斜线
    let newPath = '';
    if (pathSegments.length > 0) {
        newPath = `/${lang}/${pathSegments.join('/')}`;
    } else {
        newPath = `/${lang}`;
    }

    // 确保 URL 不以斜线结尾
    if (newPath.endsWith('/') && newPath.length > 1) {
        newPath = newPath.slice(0, -1);
    }

    window.location.href = newPath;
}

// 主题管理器Hook
function useThemeManager() {
    const [isDarkTheme, setIsDarkTheme] = useState(true);

    useEffect(() => {
        // 设置默认为dark模式，如果localStorage中没有设置主题，就默认使用dark模式
        if (localStorage.theme === 'light') {
            document.documentElement.classList.remove('dark');
            setIsDarkTheme(false);
        } else {
            // 默认使用dark模式
            document.documentElement.classList.add('dark');
            localStorage.theme = 'dark';
            setIsDarkTheme(true);
        }

        // 监听系统主题变化
        const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        const handleDarkModeChange = (e: MediaQueryListEvent) => {
            // 只有在用户没有手动设置主题时才跟随系统
            if (!('theme' in localStorage)) {
                if (e.matches) {
                    document.documentElement.classList.add('dark');
                    setIsDarkTheme(true);
                } else {
                    // 即使系统是亮色模式，我们也保持暗色模式作为默认值
                    document.documentElement.classList.add('dark');
                    localStorage.theme = 'dark';
                    setIsDarkTheme(true);
                }
            }
        };

        // 添加事件监听
        darkModeMediaQuery.addEventListener('change', handleDarkModeChange);

        // 清理函数
        return () => {
            darkModeMediaQuery.removeEventListener('change', handleDarkModeChange);
        };
    }, []);

    const toggleTheme = () => {
        if (isDarkTheme) {
            document.documentElement.classList.remove('dark');
            localStorage.theme = 'light';
            setIsDarkTheme(false);
        } else {
            document.documentElement.classList.add('dark');
            localStorage.theme = 'dark';
            setIsDarkTheme(true);
        }
    };

    const switchToAutoTheme = () => {
        // 删除localStorage中的主题设置
        localStorage.removeItem('theme');

        // 跟随系统设置
        const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
        if (isDarkMode) {
            document.documentElement.classList.add('dark');
            setIsDarkTheme(true);
        } else {
            document.documentElement.classList.remove('dark');
            setIsDarkTheme(false);
        }
    };

    return { isDarkTheme, toggleTheme, switchToAutoTheme };
}

// 主题切换按钮组件
const ThemeToggleButton = ({ isDarkTheme, toggleTheme, isScrolled }: { isDarkTheme: boolean; toggleTheme: () => void; isScrolled: boolean }) => (
    <button
        onClick={toggleTheme}
        className={`${isScrolled ? 'text-gray-600 dark:text-gray-400' : 'text-gray-400 dark:text-gray-600'}  hover:text-amber-500 transition-colors`}
        aria-label={isDarkTheme ? t('theme.light', 'Light Mode') : t('theme.dark', 'Dark Mode')}
    >
        {isDarkTheme ? (
            <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" />
            </svg>
        ) : (
            <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
            </svg>
        )}
    </button>
);

// Logo组件
const Logo = ({ logo, title, isScrolled }: { logo?: string; title: string; isScrolled: boolean }) => (
    <div className="flex-shrink-0 flex items-center">
        <a href="/" className="flex items-center">
            {logo && <img src={logo} alt="logo" className="h-8 w-8" />}
            <span className={logo ? 'ml-2' : '' + ' text-xl font-bold text-amber-600'}>{t(title, title)}</span>
        </a>
    </div>
);

// 导航链接组件
const NavLinks = ({ navItems, isLoggedIn, isScrolled }: { navItems: any[]; isLoggedIn: boolean; isScrolled: boolean }) => (
    <div className="hidden md:flex items-center justify-center flex-1">
        <div className="flex space-x-8">
            {(navItems || []).filter((nav) => !nav.IsLogin || isLoggedIn).map((nav) => (
                <a key={nav.text} href={getUrlWithLang(nav.url)}
                    className={`px-3 py-2 font-medium text-sm transition-colors rounded-md
                    ${matchPath(nav.url, window.location.pathname)
                            ? isScrolled ? 'text-amber-500 bg-white/5' : 'text-white bg-white/10'
                            : isScrolled ? 'text-gray-400 hover:text-gray-100 hover:bg-white/5' : 'text-white/90 hover:text-white hover:bg-white/10'}`}>
                    {t(nav.text, nav.text)}
                </a>
            ))}
        </div>
    </div>
);

// 语言切换下拉菜单
const LanguageDropdown = ({
    supportedLangs,
    supportedLangsMap,
    currentLang,
    language,
    changeLanguage,
    isScrolled
}: {
    supportedLangs: string[];
    supportedLangsMap: Record<string, string>;
    currentLang: string;
    language: string;
    changeLanguage: (lang: string) => void;
    isScrolled: boolean;
}) => (
    <DropdownMenu.Root>
        <DropdownMenu.Trigger>
            <button
                type="button"
                className={`flex items-center ${isScrolled ? 'text-gray-700 dark:text-gray-300 hover:text-amber-600' : 'text-gray-400 dark:text-gray-600 hover:text-white'} transition-colors`}
            >
                <span className='mr-1'>{language}</span>
                <DropdownMenu.TriggerIcon />
            </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content variant="soft" align="end">
            {supportedLangs.map((lang) => (
                <DropdownMenu.Item
                    key={lang}
                    onClick={() => changeLanguage(lang)}
                >
                    <div className='flex w-full justify-end'>
                        {currentLang === lang && (
                            <svg className="inline-block h-4 w-4 mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
                            </svg>
                        )}
                        {supportedLangsMap[lang]}
                    </div>
                </DropdownMenu.Item>
            ))}
        </DropdownMenu.Content>
    </DropdownMenu.Root>
);

// 用户菜单组件
const UserMenu = ({ isLoggedIn, userInfo, handleLogout, isScrolled }: { isLoggedIn: boolean; userInfo: any; handleLogout: () => void; isScrolled: boolean }) => (
    isLoggedIn ? (
        <DropdownMenu.Root>
            <DropdownMenu.Trigger>
                <button
                    type="button"
                    className={`flex items-center ${isScrolled ? 'text-gray-700 hover:text-amber-600' : 'text-white/80 hover:text-white'} transition-colors`}
                >
                    {userInfo?.avatar_url ? (
                        <img
                            src={userInfo.avatar_url}
                            alt="avatar"
                            className="h-8 w-8 rounded-full"
                        />
                    ) : (
                        <div className="h-8 w-8 rounded-full bg-amber-600 flex items-center justify-center text-white">
                            {(userInfo?.user_name || '').charAt(0).toUpperCase()}
                        </div>
                    )}
                </button>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content variant="soft" align="end">
                <DropdownMenu.Item>
                    <span className="text-sm font-medium">{userInfo.user_name || userInfo.email}</span>
                </DropdownMenu.Item>
                <DropdownMenu.Separator />
                <DropdownMenu.Item onClick={handleLogout}>
                    {t('root.logout_button', 'Logout')}
                </DropdownMenu.Item>
            </DropdownMenu.Content>
        </DropdownMenu.Root>
    ) : (
        <Button variant="outline" color={isScrolled ? "gray" : "amber"} onClick={() => showLoginModal()}>
            {t('root.login_button', 'Login')}
        </Button>
    )
);

// 移动端菜单组件
const MobileMenu = ({
    isOpen,
    onClose,
    isDarkTheme,
    toggleTheme,
    navItems,
    isLoggedIn,
    supportedLangs,
    supportedLangsMap,
    currentLang,
    changeLanguage,
    handleLogout,
    userInfo
}: {
    isOpen: boolean;
    onClose: () => void;
    isDarkTheme: boolean;
    toggleTheme: () => void;
    navItems: any[];
    isLoggedIn: boolean;
    supportedLangs: string[];
    supportedLangsMap: Record<string, string>;
    currentLang: string;
    changeLanguage: (lang: string) => void;
    handleLogout: () => void;
    userInfo: any;
}) => {
    const multiLang = supportedLangs.length > 1;

    // 创建抽屉的容器元素，确保其附加到body上，避免受到父元素样式影响
    useEffect(() => {
        // 仅在客户端执行
        if (isOpen) {
            // 防止滚动
            document.body.style.overflow = 'hidden';
        } else {
            // 恢复滚动
            document.body.style.overflow = '';
        }

        return () => {
            // 清理
            document.body.style.overflow = '';
        };
    }, [isOpen]);

    // 如果没有打开，不渲染菜单内容，避免水合不匹配
    if (!isOpen && typeof window !== 'undefined') {
        return null;
    }

    return (
        <>
            {/* 遮罩层，仅在isOpen=true时显示 */}
            {isOpen && (
                <div
                    className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[9999]"
                    onClick={onClose}
                />
            )}

            {/* 侧边抽屉，使用CSS控制可见性 */}
            <div className={`fixed top-0 right-0 bottom-0 w-[280px] h-screen ${isDarkTheme ? 'bg-gray-900' : 'bg-white'} 
                ${isOpen ? 'translate-x-0' : 'translate-x-full'} transition-transform duration-300 ease-in-out z-[10000] 
                shadow-[-4px_0_10px_rgba(0,0,0,0.1)] flex flex-col`}>

                {/* 抽屉头部 */}
                <div className={`flex items-center justify-between p-4 ${isDarkTheme ? 'border-gray-700' : 'border-gray-200'} border-b`}>
                    <h2 className={`text-lg font-semibold ${isDarkTheme ? 'text-white' : 'text-gray-900'}`}>
                        {t('menu', 'Menu')}
                    </h2>
                    <button
                        onClick={onClose}
                        className={`p-2 rounded-md ${isDarkTheme ? 'text-gray-400 hover:text-gray-300' : 'text-gray-500 hover:text-gray-700'} 
                            cursor-pointer bg-transparent border-none`}
                        aria-label="Close Menu"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                {/* 抽屉内容 */}
                <div className="flex-1 overflow-y-auto p-4">
                    <div className="flex flex-col gap-6">
                        {/* 主题切换 */}
                        <button
                            onClick={() => {
                                toggleTheme();
                                onClose();
                            }}
                            className={`flex items-center w-full ${isDarkTheme ? 'text-gray-200' : 'text-gray-600'} 
                                bg-transparent border-none py-2 cursor-pointer text-left`}
                        >
                            {isDarkTheme ? (
                                <>
                                    <svg className="w-5 h-5 mr-3" fill="currentColor" viewBox="0 0 20 20">
                                        <path d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" />
                                    </svg>
                                    {t('theme.light', 'Light Mode')}
                                </>
                            ) : (
                                <>
                                    <svg className="w-5 h-5 mr-3" fill="currentColor" viewBox="0 0 20 20">
                                        <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
                                    </svg>
                                    {t('theme.dark', 'Dark Mode')}
                                </>
                            )}
                        </button>

                        {/* 导航链接 */}
                        <div className="flex flex-col gap-1">
                            {(navItems || []).filter((nav) => !nav.IsLogin || isLoggedIn).map((nav) => {
                                const isActive = matchPath(nav.url, window.location.pathname);
                                return (
                                    <a
                                        key={nav.text}
                                        href={getUrlWithLang(nav.url)}
                                        onClick={onClose}
                                        className={`block px-3 py-2 rounded-md transition-colors 
                                            ${isActive
                                                ? (isDarkTheme ? 'bg-amber-700/10 text-amber-600' : 'bg-amber-50 text-amber-600')
                                                : (isDarkTheme ? 'text-gray-200 hover:bg-gray-800' : 'text-gray-600 hover:bg-gray-100')}`}
                                    >
                                        {t(nav.text, nav.text)}
                                    </a>
                                );
                            })}
                        </div>

                        {/* 语言切换 */}
                        {multiLang && (
                            <div className={`py-4 ${isDarkTheme ? 'border-gray-700' : 'border-gray-200'} border-t`}>
                                <h3 className={`text-sm font-medium ${isDarkTheme ? 'text-gray-400' : 'text-gray-500'} mb-3`}>
                                    {t('language', 'Language')}
                                </h3>
                                <div className="flex flex-col gap-2">
                                    {supportedLangs.map((lang) => {
                                        const isActive = currentLang === lang;
                                        return (
                                            <button
                                                key={lang}
                                                onClick={() => {
                                                    changeLanguage(lang);
                                                    onClose();
                                                }}
                                                className={`flex items-center justify-between w-full px-3 py-2 rounded-md 
                                                    ${isActive
                                                        ? (isDarkTheme ? 'bg-amber-700/10 text-amber-600' : 'bg-amber-50 text-amber-600')
                                                        : (isDarkTheme ? 'text-gray-200 hover:bg-gray-800' : 'text-gray-600 hover:bg-gray-100')}
                                                    border-none text-left cursor-pointer`}
                                            >
                                                <span>{supportedLangsMap[lang]}</span>
                                                {isActive && (
                                                    <svg className="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                                                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                                                    </svg>
                                                )}
                                            </button>
                                        );
                                    })}
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* 抽屉底部 */}
                <div className={`p-4 ${isDarkTheme ? 'border-gray-700' : 'border-gray-200'} border-t`}>
                    {isLoggedIn ? (
                        <div>
                            <div className="flex items-center gap-2 mb-3">
                                {userInfo?.avatar_url ? (
                                    <img
                                        src={userInfo.avatar_url}
                                        alt="avatar"
                                        className="h-8 w-8 rounded-full"
                                    />
                                ) : (
                                    <div className="h-8 w-8 rounded-full bg-amber-600 flex items-center justify-center text-white">
                                        {(userInfo?.user_name || '').charAt(0).toUpperCase()}
                                    </div>
                                )}
                                <span className="text-sm font-medium">{userInfo.user_name || userInfo.email}</span>
                            </div>
                            <button
                                onClick={() => {
                                    handleLogout();
                                    onClose();
                                }}
                                className={`w-full py-2 px-4 rounded-md ${isDarkTheme ? 'bg-gray-800 text-gray-200' : 'bg-gray-100 text-gray-600'} 
                                    border-none cursor-pointer text-sm font-medium text-center`}
                            >
                                {t('root.logout_button', 'Logout')}
                            </button>
                        </div>
                    ) : (
                        <button
                            onClick={() => {
                                showLoginModal();
                                onClose();
                            }}
                            className={`w-full py-2 px-4 rounded-md ${isDarkTheme ? 'bg-amber-600' : 'bg-amber-500'} text-white border-none cursor-pointer text-sm font-medium text-center shadow`}
                        >
                            {t('root.login_button', 'Login')}
                        </button>
                    )}
                </div>
            </div>
        </>
    );
};

// 主头部组件
export function Header() {
    const [language, setLanguage] = useState('');
    const [currentLang, setCurrentLang] = useState('');
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [userInfo, setUserInfo] = useState<any>(null);
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
    const [isScrolled, setIsScrolled] = useState(false);

    const { isDarkTheme, toggleTheme, switchToAutoTheme } = useThemeManager();

    const headerInfo = getTranslations("root.header", {
        logo: "",
        title: "",
        nav: [],
    });
    const website = getWebsite();
    const supportedLangs = website.SupportLang || [];
    const supportedLangsMap = website.LangMap || {};
    const multiLang = supportedLangs.length > 1;

    useEffect(() => {
        if (multiLang) {
            const htmlLang = document.documentElement.lang || 'en';
            setCurrentLang(htmlLang);
            setLanguage(supportedLangsMap[htmlLang] || 'English');
        }

        const authenticated = isAuthenticated();
        setIsLoggedIn(authenticated);

        if (authenticated) {
            const user = getUser();
            setUserInfo(user);
        }

        // 添加滚动监听
        const handleScroll = () => {
            if (window.scrollY > 10) {
                setIsScrolled(true);
            } else {
                setIsScrolled(false);
            }
        };

        // 页面加载时立即检查滚动位置，处理带锚点刷新的情况
        handleScroll();

        // 添加滚动事件监听器
        window.addEventListener('scroll', handleScroll);

        return () => {
            window.removeEventListener('scroll', handleScroll);
        };
    }, []);

    const changeLanguage = (lang: string) => {
        // 更新语言状态
        setLanguage(supportedLangsMap[lang] || 'English');
        setCurrentLang(lang);
        switchLang(lang, supportedLangs);
    };

    const handleLogout = () => {
        setIsLoggedIn(false);
        setUserInfo(null);
        logout();
    };

    const toggleMobileMenu = () => {
        setMobileMenuOpen(!mobileMenuOpen);
    };

    const closeMobileMenu = () => {
        setMobileMenuOpen(false);
    };

    return (
        <nav className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${isScrolled ? 'bg-white/80 dark:bg-gray-900/80 backdrop-blur-md' : 'bg-transparent border-transparent'}`}>
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                {/* 移动端头部 */}
                <div className="md:!hidden flex justify-between h-16 items-center px-4">
                    <Logo logo={headerInfo.logo} title={headerInfo.title} isScrolled={isScrolled} />

                    <div className="flex items-center space-x-2">
                        <ThemeToggleButton isDarkTheme={isDarkTheme} toggleTheme={toggleTheme} isScrolled={isScrolled} />
                        <Button variant="outline" color={isScrolled ? "gray" : "amber"} onClick={toggleMobileMenu}>
                            <HamburgerMenuIcon />
                        </Button>
                    </div>
                </div>

                {/* 移动端菜单 */}
                <MobileMenu
                    isOpen={mobileMenuOpen}
                    onClose={closeMobileMenu}
                    isDarkTheme={isDarkTheme}
                    toggleTheme={toggleTheme}
                    navItems={headerInfo.nav || []}
                    isLoggedIn={isLoggedIn}
                    supportedLangs={supportedLangs}
                    supportedLangsMap={supportedLangsMap}
                    currentLang={currentLang}
                    changeLanguage={changeLanguage}
                    handleLogout={handleLogout}
                    userInfo={userInfo}
                />

                {/* 桌面端头部 */}
                <div className="hidden md:flex justify-between h-16 items-center px-4">
                    <Logo logo={headerInfo.logo} title={headerInfo.title} isScrolled={isScrolled} />
                    <NavLinks navItems={headerInfo.nav || []} isLoggedIn={isLoggedIn} isScrolled={isScrolled} />

                    {/* 右侧按钮组 */}
                    <div className="flex items-center space-x-4">
                        <ThemeToggleButton isDarkTheme={isDarkTheme} toggleTheme={toggleTheme} isScrolled={isScrolled} />

                        {multiLang && (
                            <LanguageDropdown
                                supportedLangs={supportedLangs}
                                supportedLangsMap={supportedLangsMap}
                                currentLang={currentLang}
                                language={language}
                                changeLanguage={changeLanguage}
                                isScrolled={isScrolled}
                            />
                        )}

                        <UserMenu isLoggedIn={isLoggedIn} userInfo={userInfo} handleLogout={handleLogout} isScrolled={isScrolled} />
                    </div>
                </div>
            </div>
        </nav>
    );
}
