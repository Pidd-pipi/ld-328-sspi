import { useEffect, useState } from 'react';
import { Alert, Card, Col, Row, Space, Statistic, Spin } from 'antd';
import { ShoppingOutlined, ExclamationCircleOutlined, StopOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { useFamilyStore } from '../stores/familyStore';
import { getDashboard, type DashboardData } from '../api/stats';
import FoodCard from '../components/common/FoodCard';
import FreshnessBadge from '../components/common/FreshnessBadge';
import RemainingDaysBar from '../components/common/RemainingDaysBar';
import { FoodCategoryLabels } from '../constants/food';

export default function Dashboard() {
  const { currentFamily } = useFamilyStore();
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!currentFamily) return;
    setLoading(true);
    getDashboard(currentFamily.id)
      .then(setData)
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, [currentFamily]);

  if (!currentFamily) return <Alert type="info" message="请先创建或加入家庭组" showIcon />;

  return (
    <Spin spinning={loading}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Row gutter={16}>
          <Col span={6}><Card><Statistic title="食品总数" value={data?.total_items ?? 0} prefix={<ShoppingOutlined />} /></Card></Col>
          <Col span={6}><Card><Statistic title="临期食品" value={data?.expiring_count ?? 0} prefix={<ExclamationCircleOutlined />} valueStyle={{ color: '#faad14' }} /></Card></Col>
          <Col span={6}><Card><Statistic title="已过期" value={data?.expired_count ?? 0} prefix={<StopOutlined />} valueStyle={{ color: '#ff4d4f' }} /></Card></Col>
          <Col span={6}><Card><Statistic title="已消耗" value={data?.consumed_count ?? 0} prefix={<CheckCircleOutlined />} /></Card></Col>
        </Row>

        {(data?.expiring_items?.length || data?.expired_items?.length) ? (
          <Alert
            type={data!.expired_items.length ? 'error' : 'warning'}
            showIcon
            message="临期提醒"
            description={`当前有 ${data!.expiring_items.length} 件临期、${data!.expired_items.length} 件过期食品，请及时处理。`}
          />
        ) : null}

        <Row gutter={16}>
          <Col span={12}>
            <Card title="临期食品" size="small">
              {data?.expiring_items?.length ? (
                <Row gutter={[12, 12]}>{data.expiring_items.map((it) => (
                  <Col span={12} key={it.id}><FoodCard item={it} /></Col>
                ))}</Row>
              ) : <div style={{ color: '#999' }}>暂无临期食品</div>}
            </Card>
          </Col>
          <Col span={12}>
            <Card title="已过期食品" size="small">
              {data?.expired_items?.length ? (
                <Row gutter={[12, 12]}>{data.expired_items.map((it) => (
                  <Col span={12} key={it.id}>
                    <Card size="small" title={<Space>{it.name}<FreshnessBadge status={it.status} expiryDate={it.expiry_date} /></Space>}>
                      <Space direction="vertical" size={2}>
                        <span>类别：{FoodCategoryLabels[it.category] ?? it.category}</span>
                        <RemainingDaysBar expiryDate={it.expiry_date} />
                      </Space>
                    </Card>
                  </Col>
                ))}</Row>
              ) : <div style={{ color: '#999' }}>暂无过期食品</div>}
            </Card>
          </Col>
        </Row>

        <Card title="最近通知" size="small">
          {data?.recent_notify?.length ? data.recent_notify.map((n) => (
            <Alert key={n.id} type={n.type === 'expired' ? 'error' : 'warning'} showIcon message={n.title} description={n.content} style={{ marginBottom: 8 }} />
          )) : <div style={{ color: '#999' }}>暂无通知</div>}
        </Card>
      </Space>
    </Spin>
  );
}
