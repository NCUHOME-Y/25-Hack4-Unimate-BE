import apiClient from './apiClient';
import type { User, ApiResponse } from '../lib/types';

/**
 * 用户服务 - 处理所有用户相关的 API 请求
 */
export const userService = {
  /**
   * 获取用户信息
   */
  getProfile: async (): Promise<User> => {
    const response = await apiClient.get<ApiResponse<User>>('/user/profile');
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    throw new Error(response.data.message || '获取用户信息失败');
  },

  /**
   * 更新用户信息
   */
  updateProfile: async (data: Partial<User>): Promise<User> => {
    const response = await apiClient.put<ApiResponse<User>>('/user/profile', data);
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    throw new Error(response.data.message || '更新用户信息失败');
  },
};
