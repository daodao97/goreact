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
    handleLogout
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
}) => {
    if (!isOpen) return null;

    const multiLang = supportedLangs.length > 1;

    return (
        <div className="md:hidden bg-black/90 backdrop-blur-md shadow-lg fixed left-0 right-0 top-16 bottom-0 z-50 py-4 px-6 overflow-y-auto">
            <div className="flex flex-col space-y-4">
                <div className="flex justify-end">
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-amber-600 transition-colors"
                        aria-label="关闭菜单"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                {/* 主题切换选项 */}
                <button
                    onClick={() => {
                        toggleTheme();
                        onClose();
                    }}
                    className="flex items-center text-gray-100 hover:text-amber-600 py-2 font-medium text-sm transition-colors"
                >
                    {isDarkTheme ? (
                        <>
                            <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                                <path d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" />
                            </svg>
                            {t('theme.light', '切换为亮色模式')}
                        </>
                    ) : (
                        <>
                            <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                                <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
                            </svg>
                            {t('theme.dark', '切换为暗色模式')}
                        </>
                    )}
                </button>

                {/* 导航链接 */}
                {(navItems || []).filter((nav) => !nav.IsLogin || isLoggedIn).map((nav) => (
                    <a key={nav.Text}
                        href={getUrlWithLang(nav.URL)}
                        onClick={onClose}
                        className={`hover:text-amber-600 py-2 font-medium text-sm transition-colors ${matchPath(nav.URL, window.location.pathname) ? 'text-amber-600' : 'text-gray-100'}`}>
                        {t(nav.Text, nav.Text)}
                    </a>
                ))}

                {(navItems || []).length > 0 && (
                    <div className="border-t border-gray-200 my-2"></div>
                )}

                {/* 语言切换 */}
                {multiLang && (
                    <div className="py-2">
                        <p className="text-sm text-gray-400 mb-2">{t('language', 'Language')}</p>
                        <div className="flex flex-col space-y-2">
                            {supportedLangs.map((lang) => (
                                <button
                                    key={lang}
                                    onClick={() => {
                                        changeLanguage(lang);
                                        onClose();
                                    }}
                                    className={`text-left py-1 ${currentLang === lang ? 'text-amber-600 font-medium' : 'text-gray-300'}`}
                                >
                                    {supportedLangsMap[lang]}
                                    {currentLang === lang && (
                                        <span className="ml-2">✓</span>
                                    )}
                                </button>
                            ))}
                        </div>
                    </div>
                )}

                {multiLang && (
                    <div className="border-t border-gray-200 my-2"></div>
                )}

                {isLoggedIn ? (
                    <div className="flex flex-col space-y-2">
                        <button
                            onClick={() => {
                                handleLogout();
                                onClose();
                            }}
                            className="text-left text-gray-100 hover:text-amber-600 py-1"
                        >
                            {t('root.logout_button', 'Logout')}
                        </button>
                    </div>
                ) : (
                    <Button
                        variant="outline"
                        color="gray"
                        onClick={() => {
                            showLoginModal();
                            onClose();
                        }}
                        className="w-full mt-2"
                    >
                        {t('root.login_button', 'Login')}
                    </Button>
                )}
            </div>
        </div>
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

        // 确保在下一个渲染周期后再添加事件监听器
        setTimeout(() => {
            window.addEventListener('scroll', handleScroll);
        }, 100);

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
