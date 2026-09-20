import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Col, Drawer, Form, Input, InputNumber, Modal, Row, Select, Space, Table, Upload, message } from 'antd';
import { PlusOutlined, UploadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useFamilyStore } from '../stores/familyStore';
import { consumeFood, createFood, deleteFood, getFoodDetail, importFoodsCSV, listFoods, updateFood } from '../api/foodItem';
import type { ConsumptionRecord, FoodItem } from '../types';
import FreshnessBadge from '../components/common/FreshnessBadge';
import RemainingDaysBar from '../components/common/RemainingDaysBar';
import { FoodCategories, FoodCategoryLabels, StorageLocationLabels } from '../constants/food';
import { formatDateTime } from '../utils/dateFormat';
import { usePagination } from '../hooks/usePagination';
import { useFoodStore } from '../stores/foodStore';

export default function FoodManage() {
  const { currentFamily } = useFamilyStore();
  const { items, total, loading, fetchList } = useFoodStore();
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const [filters, setFilters] = useState({ category: '', status: '', storage_location: '', keyword: '' });
  const [form] = Form.useForm();
  const [editing, setEditing] = useState<FoodItem | null>(null);
  const [open, setOpen] = useState(false);
  const [detail, setDetail] = useState<{ item: FoodItem; consumption_records: ConsumptionRecord[] } | null>(null);
  const [consumeQty, setConsumeQty] = useState(1);
  const [csvText, setCsvText] = useState('');

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    if (!currentFamily) return;
    const data = await fetchList({ family_id: currentFamily.id, page, page_size: size, ...filters });
    setTotal(data.total);
  }, [currentFamily, filters, fetchList, pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  async function onSave(values: any) {
    if (!currentFamily) return;
    const payload = { ...values, family_id: currentFamily.id, quantity: values.quantity ?? 1, shelf_life_days: values.shelf_life_days ?? 7 };
    if (editing) await updateFood(editing.id, payload);
    else await createFood(payload);
    message.success('保存成功');
    setOpen(false);
    load();
  }

  async function onConsume(item: FoodItem) {
    await consumeFood(item.id, consumeQty);
    message.success('消耗成功');
    setDetail(null);
    load();
  }

  async function onImport() {
    if (!currentFamily) return;
    await importFoodsCSV(currentFamily.id, csvText);
    message.success('CSV 导入成功');
    setCsvText('');
    load();
  }

  async function onDelete(item: FoodItem) {
    await deleteFood(item.id);
    message.success('删除成功');
    load();
  }

  const columns: ColumnsType<FoodItem> = [
    { title: '食品', dataIndex: 'name' },
    { title: '类别', dataIndex: 'category', render: (v) => FoodCategoryLabels[v] ?? v },
    { title: '数量', render: (_, r) => `${r.quantity} ${r.unit}` },
    { title: '存放位置', dataIndex: 'storage_location', render: (v) => StorageLocationLabels[v] ?? v },
    { title: '状态', render: (_, r) => <FreshnessBadge status={r.status} expiryDate={r.expiry_date} /> },
    { title: '剩余', render: (_, r) => <RemainingDaysBar expiryDate={r.expiry_date} /> },
    { title: '操作', render: (_, r) => (
      <Space>
        <a onClick={() => getFoodDetail(r.id).then(setDetail)}>详情</a>
        <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true); }}>编辑</a>
        <a onClick={() => { setDetail({ item: r, consumption_records: [] }); setConsumeQty(1); }}>消耗</a>
        <a style={{ color: '#ff4d4f' }} onClick={() => Modal.confirm({ title: '确认删除？', onOk: () => onDelete(r) })}>删除</a>
      </Space>
    ) },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Space wrap>
          <Input.Search placeholder="搜索食品" allowClear style={{ width: 200 }} onSearch={(v) => setFilters((f) => ({ ...f, keyword: v }))} />
          <Select placeholder="类别" allowClear style={{ width: 120 }} options={FoodCategories.map((c) => ({ value: c, label: FoodCategoryLabels[c] }))} onChange={(v) => setFilters((f) => ({ ...f, category: v ?? '' }))} />
          <Select placeholder="状态" allowClear style={{ width: 120 }} options={[['fresh','充裕'],['expiring','临期'],['expired','已过期'],['consumed','已消耗']].map(([value,label]) => ({ value, label }))} onChange={(v) => setFilters((f) => ({ ...f, status: v ?? '' }))} />
          <Select placeholder="存放位置" allowClear style={{ width: 130 }} options={Object.entries(StorageLocationLabels).map(([value,label]) => ({ value, label }))} onChange={(v) => setFilters((f) => ({ ...f, storage_location: v ?? '' }))} />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true); }}>新增食品</Button>
          <Button icon={<UploadOutlined />} onClick={() => setCsvText('示例：\n牛奶,dairy,2,盒,5,fridge\n苹果,fresh,3,个,14,pantry')}>CSV 批量导入</Button>
        </Space>
      </Card>

      <Card size="small" title="食品列表">
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, showSizeChanger: true, onChange: onPageChange }} />
      </Card>

      <Drawer open={open} title={editing ? '编辑食品' : '新增食品'} width={420} onClose={() => setOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={onSave} initialValues={{ quantity: 1, unit: '份', shelf_life_days: 7, storage_location: 'fridge', category: 'fresh' }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="category" label="类别" rules={[{ required: true }]}><Select options={FoodCategories.map((c) => ({ value: c, label: FoodCategoryLabels[c] }))} /></Form.Item>
          <Form.Item name="quantity" label="数量" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="unit" label="单位"><Input /></Form.Item>
          <Form.Item name="shelf_life_days" label="保质期（天）"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="storage_location" label="存放位置"><Select options={Object.entries(StorageLocationLabels).map(([value,label]) => ({ value, label }))} /></Form.Item>
          <Button type="primary" htmlType="submit" block>保存</Button>
        </Form>
      </Drawer>

      <Modal open={!!detail} title={detail?.item.name} footer={null} onCancel={() => setDetail(null)} width={560}>
        {detail && (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Row gutter={16}>
              <Col span={8}>数量：{detail.item.quantity} {detail.item.unit}</Col>
              <Col span={8}>类别：{FoodCategoryLabels[detail.item.category]}</Col>
              <Col span={8}><FreshnessBadge status={detail.item.status} expiryDate={detail.item.expiry_date} /></Col>
            </Row>
            <Space>
              <InputNumber min={0.1} value={consumeQty} onChange={(v) => setConsumeQty(v ?? 1)} />
              <Button type="primary" onClick={() => onConsume(detail.item)}>记录消耗</Button>
            </Space>
            <Card size="small" title="历史消耗记录">
              {detail.consumption_records.length ? detail.consumption_records.map((r) => (
                <div key={r.id}>{formatDateTime(r.consumed_at)} - {r.quantity} {detail.item.unit}（{r.user?.name || r.user?.phone}）</div>
              )) : <span style={{ color: '#999' }}>暂无消耗记录</span>}
            </Card>
          </Space>
        )}
      </Modal>

      <Modal open={!!csvText} title="CSV 批量导入" onOk={onImport} onCancel={() => setCsvText('')} okText="导入">
        <Input.TextArea rows={8} value={csvText} onChange={(e) => setCsvText(e.target.value)} placeholder="name,category,quantity,unit,shelf_life_days,storage_location" />
      </Modal>
    </Space>
  );
}
