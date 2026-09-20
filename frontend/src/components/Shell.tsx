import { useEffect } from 'react';
import { Layout, Menu, Badge, Dropdown, Avatar, Spin } from 'antd';
import { DashboardOutlined, ShoppingOutlined, HistoryOutlined, BarChartOutlined, TeamOutlined, BulbOutlined, UserOutlined, LogoutOutlined } from '@ant-design/icons';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../stores/authStore';
import { useFamilyStore } from '../stores/familyStore';
import { useNotificationStore } from '../stores/notificationStore';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '保质期看板' },
  { key: '/foods', icon: <ShoppingOutlined />, label: '食品管理' },
  { key: '/consumptions', icon: <HistoryOutlined />, label: '消耗记录' },
  { key: '/statistics', icon: <BarChartOutlined />, label: '分类统计' },
  { key: '/family', icon: <TeamOutlined />, label: '家庭管理' },
  { key: '/recommendations', icon: <BulbOutlined />, label: '智能推荐' },
  { key: '/profile', icon: <UserOutlined />, label: '个人中心' },
];

export default function Shell() {
  const { user, logout } = useAuth();
  const { families, currentFamily, refresh, selectFamily } = useFamilyStore();
  const { unread, refresh: refreshNotify } = useNotificationStore();
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    if (!families.length) refresh().catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (currentFamily) refreshNotify(currentFamily.id).catch(() => undefined);
  }, [currentFamily, refreshNotify]);

  if (!user) return <Spin style={{ margin: '40vh 50%' }} />;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth={64}>
        <div style={{ color: '#fff', padding: 16, fontWeight: 600, fontSize: 15 }}>CyFreshFood</div>
        <Menu theme="dark" mode="inline" selectedKeys={[location.pathname]} items={menuItems} onClick={({ key }) => navigate(key)} />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <Dropdown
              menu={{
                items: families.map((f) => ({ key: String(f.id), label: f.name })),
                onClick: ({ key }) => {
                  const f = families.find((x) => String(x.id) === key);
                  if (f) selectFamily(f);
                },
              }}
              disabled={families.length === 0}
            >
              <span style={{ cursor: 'pointer' }}>
                {currentFamily ? `当前家庭：${currentFamily.name}` : '暂无家庭组'}
                {families.length > 1 ? ' ▾' : ''}
              </span>
            </Dropdown>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <Badge count={unread} size="small">
              <span>临期提醒</span>
            </Badge>
            <Dropdown
              menu={{
                items: [{ key: 'logout', icon: <LogoutOutlined />, label: '退出登录' }],
                onClick: ({ key }) => {
                  if (key === 'logout') {
                    logout();
                    navigate('/login');
                  }
                },
              }}
            >
              <Avatar size="small" src={user.avatar || undefined} icon={<UserOutlined />} style={{ cursor: 'pointer' }}>
                {!user.avatar && (user.name || 'U').slice(0, 1)}
              </Avatar>
            </Dropdown>
          </div>
        </Header>
        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
