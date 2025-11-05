import apiClient from './apiClient';
import type { ApiResponse } from '../lib/types';
import { STORAGE_KEYS } from '../lib/constants';

interface LoginCredentials {
  email: string;
  password: string;
}

interface RegisterData {
  username: string;
  email: string;
  password: string;
}

interface AuthResponse {
  token: string;
  user: {
    id: string;
    username: string;
    email: string;
  };
}

/**
 * 认证服务 - 处理登录、注册等认证相关操作
 */
export const authService = {
  /**
   * 用户登录
   */
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/login', credentials);
    if (response.data.success && response.data.data) {
      // 保存 token 到本地存储
      localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, response.data.data.token);
      return response.data.data;
    }
    throw new Error(response.data.message || '登录失败');
  },

  /**
   * 用户注册
   */
  register: async (data: RegisterData): Promise<AuthResponse> => {
    const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data);
    if (response.data.success && response.data.data) {
      localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, response.data.data.token);
      return response.data.data;
    }
    throw new Error(response.data.message || '注册失败');
  },

  /**
   * 用户登出
   */
  logout: async (): Promise<void> => {
    try {
      await apiClient.post('/auth/logout');
    } finally {
      // 无论请求成功与否，都清除本地 token
      localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
      localStorage.removeItem(STORAGE_KEYS.USER_INFO);
    }
  },

  /**
   * 检查是否已登录
   */
  isAuthenticated: (): boolean => {
    return !!localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
  },
};
