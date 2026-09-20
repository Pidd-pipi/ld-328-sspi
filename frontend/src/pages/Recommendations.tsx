import { useCallback, useEffect, useState } from 'react';
import { Alert, Card, Col, Row, Space, Spin, Tag } from 'antd';
import { useFamilyStore } from '../stores/familyStore';
import { getRecommendations, type Recommendation } from '../api/recipe';
import FoodCard from '../components/common/FoodCard';
import FreshnessBadge from '../components/common/FreshnessBadge';
import EmptyState from '../components/common/EmptyState';
import { FoodCategoryLabels } from '../constants/food';

export default function Recommendations() {
  const { currentFamily } = useFamilyStore();
  const [data, setData] = useState<Recommendation | null>(null);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    if (!currentFamily) return;
    setLoading(true);
    try {
      setData(await getRecommendations(currentFamily.id));
    } finally {
      setLoading(false);
    }
  }, [currentFamily]);

  useEffect(() => { load().catch(() => undefined); }, [load]);

  return (
    <Spin spinning={loading}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        {data?.summary?.expiring_count ? (
          <Alert type="warning" showIcon message="优先食用建议" description={`当前有 ${data.summary.expiring_count} 件临期食品，建议按以下顺序优先食用。`} />
        ) : null}
        <Row gutter={16}>
          <Col span={14}>
            <Card size="small" title="即将过期食品（优先食用）">
              {data?.food_items?.length ? (
                <Row gutter={[12, 12]}>{data.food_items.map((it) => (
                  <Col span={12} key={it.id}><FoodCard item={it} /></Col>
                ))}</Row>
              ) : <EmptyState description="暂无临期食品，库存很健康" />}
            </Card>
          </Col>
          <Col span={10}>
            <Card size="small" title="简易食谱搭配建议">
              {data?.recipes?.length ? data.recipes.map((r) => (
                <Card key={r.id} size="small" style={{ marginBottom: 12 }}>
                  <Space direction="vertical" size={4}>
                    <b>{r.name} <FreshnessBadge status="fresh" /></b>
                    <span>适用类别：<Tag color="blue">{FoodCategoryLabels[r.suitable_category] ?? r.suitable_category}</Tag></span>
                    <span>食材：{r.ingredients}</span>
                    <span style={{ color: '#666' }}>{r.description}</span>
                  </Space>
                </Card>
              )) : <EmptyState description="暂无食谱建议" />}
            </Card>
          </Col>
        </Row>
      </Space>
    </Spin>
  );
}
