import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { Button, Card, Col, DatePicker, Row, Statistic, Table, message } from 'antd';
import { DownloadOutlined } from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';

import { useFamilyStore } from '../stores/familyStore';
import { exportStatisticsPDF, getStatistics, type StatisticsData } from '../api/stats';
import EmptyState from '../components/common/EmptyState';
import { FoodCategoryLabels } from '../constants/food';
import dayjs from 'dayjs';
import { currentMonth } from '../utils/dateFormat';

export default function Statistics() {
  const { currentFamily } = useFamilyStore();
  const [data, setData] = useState<StatisticsData | null>(null);
  const [month, setMonth] = useState(currentMonth());

  const load = useCallback(async (m: string) => {
    if (!currentFamily) return;
    const d = await getStatistics(currentFamily.id, m);
    setData(d);
  }, [currentFamily]);

  useEffect(() => { load(month).catch(() => undefined); }, [load, month]);

  function download() {
    if (!currentFamily) return;
    window.open(exportStatisticsPDF(currentFamily.id, month), '_blank');
  }

  const categoryPie = {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [{
      type: 'pie', radius: ['35%', '65%'],
      data: (data?.category_share ?? []).map((row) => ({ name: FoodCategoryLabels[row.category] ?? row.category, value: row.count })),
    }],
  };

  const columns = [
    { title: '食品名称', dataIndex: 'name' },
    { title: '次数', dataIndex: 'count' },
    { title: '数量', dataIndex: 'quantity' },
  ];

  return (
    <SpaceV>
      <Row gutter={16}>
        <Col span={8}><Card><Statistic title="浪费金额估算（元）" value={data?.waste_amount ?? 0} precision={2} valueStyle={{ color: '#ff4d4f' }} /></Card></Col>
        <Col span={8}><Card><Statistic title="过期食品数" value={data?.top_wasted?.length ?? 0} /></Card></Col>
        <Col span={8}><Card><Statistic title="消耗次数" value={(data?.top_purchased ?? []).reduce((s, t) => s + t.count, 0)} /></Card></Col>
      </Row>
      <Card
        size="small"
        title="分类统计报表"
        extra={<Space2><DatePicker picker="month" value={dayjs(month)} onChange={(d) => d && setMonth(d.format('YYYY-MM'))} /><Button type="primary" icon={<DownloadOutlined />} onClick={download}>导出 PDF</Button></Space2>}
      >
        <Row gutter={16}>
          <Col span={12}><ReactECharts option={categoryPie} style={{ height: 300 }} /></Col>
          <Col span={12}>
            <div style={{ fontWeight: 600, marginBottom: 8 }}>最常购买 Top 10</div>
            <Table size="small" rowKey={(r) => String(r.food_item_id)} columns={columns} dataSource={data?.top_purchased ?? []} pagination={false} locale={{ emptyText: <EmptyState description="暂无数据" /> }} />
            <div style={{ fontWeight: 600, margin: '12px 0 8px' }}>最常浪费 Top 10</div>
            <Table size="small" rowKey={(r) => String(r.food_item_id)} columns={columns} dataSource={data?.top_wasted ?? []} pagination={false} locale={{ emptyText: <EmptyState description="暂无数据" /> }} />
          </Col>
        </Row>
      </Card>
    </SpaceV>
  );
}

function SpaceV({ children }: { children: ReactNode }) {
  return <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>{children}</div>;
}
function Space2({ children }: { children: ReactNode }) {
  return <span style={{ display: 'inline-flex', gap: 8, alignItems: 'center' }}>{children}</span>;
}
