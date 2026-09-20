import type { ReactNode } from 'react';
import { Result } from 'antd';
import { useAuth } from '../../stores/authStore';

interface Props {
  roles: string[];
  children: ReactNode;
}

// 角色守卫：按 UserRole 控制按钮/页面可见性
export default function RoleGuard({ roles, children }: Props) {
  const { user } = useAuth();
  if (!user || !roles.includes(user.role)) {
    return <Result status="403" title="403" subTitle="当前角色无权访问该功能" />;
  }
  return <>{children}</>;
}
