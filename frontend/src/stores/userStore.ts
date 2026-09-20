import { getMe, updateProfile } from '../api/user';
import type { User } from '../types';
import { useAuth } from './authStore';

// 用户资料 store：负责拉取与更新当前用户信息
export function useUserStore() {
  const { user, updateUser } = useAuth();

  async function refresh(): Promise<User> {
    const me = await getMe();
    updateUser(me);
    return me;
  }

  async function save(data: { name?: string; avatar?: string }): Promise<User> {
    const updated = await updateProfile(data);
    updateUser(updated);
    return updated;
  }

  return { user, refresh, save };
}
