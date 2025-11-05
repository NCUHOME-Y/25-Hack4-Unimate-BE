// API 相关常量
export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/auth/login',
    LOGOUT: '/auth/logout',
    REGISTER: '/auth/register',
    REFRESH: '/auth/refresh',
  },
  USER: {
    PROFILE: '/user/profile',
    UPDATE: '/user/update',
  },
  // 根据需要添加更多端点
} as const;

// 应用配置
export const APP_CONFIG = {
  APP_NAME: import.meta.env.VITE_APP_NAME || 'My App',
  API_BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  ENABLE_DEBUG: import.meta.env.VITE_ENABLE_DEBUG === 'true',
} as const;

// 本地存储键名
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'authToken',
  USER_INFO: 'userInfo',
  THEME: 'theme',
} as const;

// 路由路径
export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  PROFILE: '/profile',
} as const;
