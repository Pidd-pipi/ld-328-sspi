import { Avatar, Tooltip } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import type { User } from '../../types';

interface Props {
  user?: User | null;
  size?: number;
}

// 家庭成员头像
export default function MemberAvatar({ user, size = 32 }: Props) {
  const name = user?.name || user?.phone || '成员';
  return (
    <Tooltip title={name}>
      <Avatar size={size} src={user?.avatar || undefined} icon={<UserOutlined />}>
        {!user?.avatar && name.slice(0, 1)}
      </Avatar>
    </Tooltip>
  );
}
