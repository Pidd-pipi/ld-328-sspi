import { useCallback, useEffect, useState } from 'react';
import { Card, Col, Row, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useFamilyStore } from '../stores/familyStore';
import { consumptionAnalysis, listConsumptions, type ConsumptionAnalysis } from '../api/consumption';
import type { ConsumptionRecord } from '../types';
import FreshnessBadge from '../components/common/FreshnessBadge';
import EmptyState from '../components/common/EmptyState';
import { FoodCategoryLabels } from '../constants/food';
import { currentMonth, formatDateTime } from '../utils/dateFormat';
import { usePagination } from '../hooks/usePagination';

export default function ConsumptionManage() {
  const { currentFamily } = useFamilyStore();
  const [records, setRecords] = useState<ConsumptionRecord[]>([]);
  const [analysis, setAnalysis] = useState<ConsumptionAnalysis | null>(null);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    if (!currentFamily) return;
    setLoading(true);
    try {
      const data = await listConsumptions({ family_id: currentFamily.id, page, page_size: size });
      setRecords(data.list);
      setTotal(data.total);
      const a = await consumptionAnalysis({ family_id: currentFamily.id, month: currentMonth() });
      setAnalysis(a);
    } finally {
      setLoading(false);
    }
  }, [currentFamily, pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  const columns: ColumnsType<ConsumptionRecord> = [
    { title: '食品', render: (_, r) => r.food_item?.name ?? '-' },
    { title: '数量', dataIndex: 'quantity' },
    { title: '单位', render: (_, r) => r.food_item?.unit ?? '-' },
    { title: '状态', render: (_, r) => <FreshnessBadge status="consumed" /> },
    { title: '消耗时间', dataIndex: 'consumed_at', render: (v) => formatDateTime(v) },
    { title: '操作人', render: (_, r) => r.user?.name || r.user?.phone || '-' },
  ];

  return (
    <Row gutter={16}>
      <Col span={16}>
        <Card size="small" title="历史消耗记录">
          <Table rowKey="id" columns={columns} dataSource={records} loading={loading} pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }} locale={{ emptyText: <EmptyState description="暂无消耗记录" /> }} />
        </Card>
      </Col>
      <Col span={8}>
        <Card size="small" title={`消耗频率分析（${currentMonth()}）`}>
          {analysis?.by_category?.length ? analysis.by_category.map((row) => (
            <div key={row.category} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0' }}>
              <span>{FoodCategoryLabels[row.category] ?? row.category}</span>
              <span>{row.count} 次 / {row.total_quantity} 单位</span>
            </div>
          )) : <EmptyState description="本月暂无消耗数据" />}
          <div style={{ marginTop: 12, fontWeight: 600 }}>最常消耗 Top 10</div>
          {analysis?.top_consumed?.length ? analysis.top_consumed.map((t, idx) => (
            <div key={t.food_item_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0' }}>
              <span>{idx + 1}. {t.name}</span>
              <span>{t.count} 次</span>
            </div>
          )) : <EmptyState description="暂无排行数据" />}
        </Card>
      </Col>
    </Row>
  );
}
