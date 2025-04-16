export const getUser = (): any | null => {
    return window.USER_INFO;
};

// 修改isAuthenticated方法，支持服务端渲染
export const isAuthenticated = (): boolean => {
    const token = getUser();
    if (!token) return false;

    return true;
};

export const saveUserInfo = (userInfo: any): void => {
    localStorage.setItem('userInfo', JSON.stringify(userInfo));
};

// 清除用户信息和 token
export const logout = (): void => {
    location.href = "/logout";
}; 