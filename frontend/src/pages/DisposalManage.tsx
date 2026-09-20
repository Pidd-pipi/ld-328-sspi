import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Badge, Button, Card, Col, Drawer, Form, Input, InputNumber, Modal, Row, Select, Space, Table, Tabs, Tag, message,
} from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { FormInstance } from 'antd';
import { useSearchParams } from 'react-router-dom';
import { useFamilyStore } from '../stores/familyStore';
import { listFoods } from '../api/foodItem';
import {
  approveDisposal, createDisposal, listDisposals, rejectDisposal,
} from '../api/disposal';
import type { DisposalRequest, FoodItem } from '../types';
import DisposalStatusBadge from '../components/common/DisposalStatusBadge';
import FreshnessBadge from '../components/common/FreshnessBadge';
import EmptyState from '../components/common/EmptyState';
import {
  DisposalMethodLabels, DisposalMethods, DisposalStatus, DisposalStatusLabels,
  FoodCategoryLabels, FreshnessStatus,
} from '../constants/food';
import { computeFreshness } from '../utils/calculateRemainingDays';
import { formatDateTime } from '../utils/dateFormat';
import { usePagination } from '../hooks/usePagination';
import { useFamilyRole } from '../hooks/useFamilyRole';

const PAGE_SIZE = 10;

export default function DisposalManage() {
  const { currentFamily } = useFamilyStore();
  const { isFamilyAdmin } = useFamilyRole();
  const [searchParams, setSearchParams] = useSearchParams();
  const presetFoodId = Number(searchParams.get('food_id') || 0);
  const [tab, setTab] = useState<string>(DisposalStatus.PENDING);
  const [statusFilter, setStatusFilter] = useState('');
  const [list, setList] = useState<DisposalRequest[]>([]);
  const [total, setTotal] = useState(0);
  const [pendingCount, setPendingCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [rejectTarget, setRejectTarget] = useState<DisposalRequest | null>(null);
  const [rejectRemark, setRejectRemark] = useState('');
  const [approving, setApproving] = useState<number | null>(null);
  const { pagination, setTotal: setPageTotal, onPageChange } = usePagination(1, PAGE_SIZE);
  const [form] = Form.useForm();

  const status = tab === DisposalStatus.PENDING ? DisposalStatus.PENDING : statusFilter;

  const load = useCallback(async (page = pagination.page) => {
    if (!currentFamily) return;
    setLoading(true);
    try {
      const data = await listDisposals({
        family_id: currentFamily.id,
        status: status || undefined,
        page,
        page_size: PAGE_SIZE,
      });
      setList(data.list);
      setTotal(data.total);
      setPageTotal(data.total);
      if (tab === DisposalStatus.PENDING) setPendingCount(data.total);
    } finally {
      setLoading(false);
    }
  }, [currentFamily, status, pagination.page, setPageTotal, tab]);

  // 待处理数量独立刷新（全部 Tab 下也要回读角标）
  const refreshPendingCount = useCallback(async () => {
    if (!currentFamily) return;
    const data = await listDisposals({ family_id: currentFamily.id, status: DisposalStatus.PENDING, page: 1, page_size: 1 });
    setPendingCount(data.total);
  }, [currentFamily]);

  useEffect(() => {
    load().catch(() => undefined);
    refreshPendingCount().catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [load]);

  // 从食品管理页带 food_id 跳转时自动打开申请抽屉并预选食品
  useEffect(() => {
    if (presetFoodId > 0 && currentFamily) {
      form.setFieldsValue({ food_item_id: presetFoodId, quantity: undefined, method: 'discard' });
      setCreateOpen(true);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [presetFoodId, currentFamily]);

  async function onCreate(values: { food_item_id: number; quantity: number; method: string; reason?: string }) {
    await createDisposal(values);
    message.success('处置申请已提交，等待家庭管理员审批');
    setCreateOpen(false);
    form.resetFields();
    await load(1);
    await refreshPendingCount();
  }

  async function onApprove(record: DisposalRequest) {
    setApproving(record.id);
    try {
      await approveDisposal(record.id);
      message.success('已批准：余量已扣减并生成消耗记录');
    } catch {
      // 并发结案/超量/已消耗失败时不改库存，刷新回读真实状态
    } finally {
      setApproving(null);
      await load(1);
      await refreshPendingCount();
    }
  }

  async function onRejectOk() {
    if (!rejectTarget) return;
    try {
      await rejectDisposal(rejectTarget.id, rejectRemark);
      message.success('已驳回申请');
      setRejectTarget(null);
      setRejectRemark('');
    } catch {
      // 并发结案失败，刷新回读
    }
    await load(1);
    await refreshPendingCount();
  }

  const columns: ColumnsType<DisposalRequest> = [
    {
      title: '食品',
      render: (_, r) => (
        <Space direction="vertical" size={0}>
          <b>{r.food_item?.name ?? `#${r.food_item_id}`}</b>
          <span style={{ color: '#999', fontSize: 12 }}>
            {r.food_item ? FoodCategoryLabels[r.food_item.category] ?? r.food_item.category : ''}
          </span>
        </Space>
      ),
    },
    { title: '申请数量', render: (_, r) => `${r.quantity} ${r.food_item?.unit ?? ''}` },
    { title: '处置方式', render: (_, r) => <Tag>{DisposalMethodLabels[r.method] ?? r.method}</Tag> },
    {
      title: '食品状态',
      render: (_, r) => (r.food_item
        ? <FreshnessBadge status={r.food_item.status} expiryDate={r.food_item.expiry_date} />
        : '-'),
    },
    { title: '申请人', render: (_, r) => r.applicant?.name || r.applicant?.phone || `#${r.applicant_id}` },
    { title: '申请时间', dataIndex: 'created_at', render: (v) => formatDateTime(v) },
    {
      title: '申请状态',
      render: (_, r) => (
        <Space direction="vertical" size={0}>
          <DisposalStatusBadge status={r.status} />
          {r.status !== DisposalStatus.PENDING && (
            <span style={{ color: '#999', fontSize: 12 }}>
              {r.status === DisposalStatus.APPROVED ? '批准' : '驳回'}人：
              {r.reviewer?.name || r.reviewer?.phone || `#${r.reviewer_id ?? '-'}`}
            </span>
          )}
          {r.status === DisposalStatus.REJECTED && r.review_remark && (
            <span style={{ color: '#999', fontSize: 12 }}>备注：{r.review_remark}</span>
          )}
        </Space>
      ),
    },
    {
      title: '操作',
      render: (_, r) => {
        if (r.status !== DisposalStatus.PENDING) {
          return r.status === DisposalStatus.APPROVED && r.consumption_record_id
            ? <span style={{ color: '#52c41a' }}>已生成消耗记录 #{r.consumption_record_id}</span>
            : <span style={{ color: '#999' }}>已结案</span>;
        }
        if (!isFamilyAdmin) {
          return <span style={{ color: '#999' }}>等待管理员审批</span>;
        }
        return (
          <Space>
            <Button
              type="link" size="small" style={{ padding: 0 }}
              loading={approving === r.id}
              onClick={() => Modal.confirm({
                title: '确认批准该处置申请？',
                content: `将扣减 ${r.food_item?.name ?? ''} ${r.quantity} ${r.food_item?.unit ?? ''}，余量归零则标记已消耗。`,
                okText: '批准',
                cancelText: '取消',
                onOk: () => onApprove(r),
              })}
            >
              批准
            </Button>
            <Button type="link" size="small" danger style={{ padding: 0 }} onClick={() => setRejectTarget(r)}>
              驳回
            </Button>
          </Space>
        );
      },
    },
  ];

  const tabItems = useMemo(() => [
    {
      key: DisposalStatus.PENDING,
      label: <Badge count={pendingCount} size="small" offset={[8, 0]}>待处理</Badge>,
    },
    { key: 'all', label: '全部申请' },
  ], [pendingCount]);

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Space wrap>
          <Button type="primary" icon={<PlusOutlined />} disabled={!currentFamily} onClick={() => setCreateOpen(true)}>
            提交处置申请
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => { load().catch(() => undefined); refreshPendingCount().catch(() => undefined); }}>
            刷新
          </Button>
          {tab === 'all' && (
            <Select
              allowClear placeholder="申请状态" style={{ width: 140 }}
              value={statusFilter || undefined}
              options={Object.entries(DisposalStatusLabels).map(([value, label]) => ({ value, label }))}
              onChange={(v) => { setStatusFilter(v ?? ''); onPageChange(1, PAGE_SIZE); }}
            />
          )}
          {!isFamilyAdmin && <Tag color="orange">仅家庭管理员可批准/驳回</Tag>}
        </Space>
      </Card>

      <Card size="small" title="临期处置申请">
        <Tabs
          activeKey={tab}
          items={tabItems}
          onChange={(k) => {
            setTab(k);
            setStatusFilter('');
            onPageChange(1, PAGE_SIZE);
          }}
        />
        <Table
          rowKey="id"
          columns={columns}
          dataSource={list}
          loading={loading}
          locale={{ emptyText: <EmptyState description={tab === DisposalStatus.PENDING ? '暂无待处理申请' : '暂无处置申请'} /> }}
          pagination={{
            current: pagination.page,
            pageSize: pagination.pageSize,
            total,
            showSizeChanger: false,
            onChange: (p) => onPageChange(p, PAGE_SIZE),
          }}
        />
      </Card>

      <CreateDisposalDrawer
        open={createOpen}
        form={form}
        familyId={currentFamily?.id ?? 0}
        presetFoodId={presetFoodId}
        onClose={() => { setCreateOpen(false); form.resetFields(); if (searchParams.has('food_id')) setSearchParams({}); }}
        onSubmit={async (values) => { await onCreate(values); if (searchParams.has('food_id')) setSearchParams({}); }}
      />

      <Modal
        open={!!rejectTarget}
        title="驳回处置申请"
        okText="确认驳回"
        cancelText="取消"
        okButtonProps={{ danger: true }}
        onOk={onRejectOk}
        onCancel={() => { setRejectTarget(null); setRejectRemark(''); }}
      >
        <p>驳回只关闭申请，不会修改食品库存。</p>
        <Input.TextArea
          rows={3}
          maxLength={255}
          showCount
          placeholder="驳回备注（可选）"
          value={rejectRemark}
          onChange={(e) => setRejectRemark(e.target.value)}
        />
      </Modal>
    </Space>
  );
}

interface DrawerProps {
  open: boolean;
  form: FormInstance;
  familyId: number;
  presetFoodId: number;
  onClose: () => void;
  onSubmit: (values: { food_item_id: number; quantity: number; method: string; reason?: string }) => Promise<void>;
}

// 提交申请表抽屉：仅可选择临期/过期且未消耗的食品
function CreateDisposalDrawer({ open, form, familyId, presetFoodId, onClose, onSubmit }: DrawerProps) {
  const [foods, setFoods] = useState<FoodItem[]>([]);
  const [pendingIds, setPendingIds] = useState<Set<number>>(new Set());
  const [loadingFoods, setLoadingFoods] = useState(false);

  useEffect(() => {
    if (!open || !familyId) return;
    let active = true;
    (async () => {
      setLoadingFoods(true);
      try {
        const [foodData, pendingData] = await Promise.all([
          listFoods({ family_id: familyId, page: 1, page_size: 200 }),
          listDisposals({ family_id: familyId, status: DisposalStatus.PENDING, page: 1, page_size: 200 }),
        ]);
        if (!active) return;
        setFoods(foodData.list);
        setPendingIds(new Set(pendingData.list.map((r) => r.food_item_id)));
        const preset = foodData.list.find((f) => f.id === presetFoodId);
        if (preset) {
          form.setFieldsValue({ food_item_id: preset.id, quantity: preset.quantity >= 1 ? 1 : preset.quantity, method: 'discard' });
        }
      } finally {
        if (active) setLoadingFoods(false);
      }
    })();
    return () => { active = false; };
  }, [open, familyId, presetFoodId, form]);

  const eligibleFoods = useMemo(() => foods.filter((f) => {
    const fresh = computeFreshness(f.status, f.expiry_date);
    return fresh === FreshnessStatus.EXPIRING || fresh === FreshnessStatus.EXPIRED;
  }), [foods]);

  const selectedFoodId = Form.useWatch('food_item_id', form);
  const selectedFood = eligibleFoods.find((f) => f.id === selectedFoodId);

  return (
    <Drawer open={open} title="提交临期处置申请" width={420} onClose={onClose} destroyOnClose>
      <Form
        form={form}
        layout="vertical"
        onFinish={onSubmit}
        initialValues={{ method: 'discard', quantity: 1 }}
      >
        <Form.Item
          name="food_item_id"
          label="临期/过期食品"
          rules={[{ required: true, message: '请选择食品' }]}
        >
          <Select
            placeholder={eligibleFoods.length ? '请选择食品' : '暂无可处置的临期/过期食品'}
            loading={loadingFoods}
            showSearch
            optionFilterProp="label"
            options={eligibleFoods.map((f) => {
              const pending = pendingIds.has(f.id);
              const fresh = computeFreshness(f.status, f.expiry_date);
              const freshText = fresh === FreshnessStatus.EXPIRED ? '已过期' : '临期';
              return {
                value: f.id,
                disabled: pending,
                label: `${f.name}（余量 ${f.quantity} ${f.unit} · ${freshText}${pending ? ' · 已有待处理申请' : ''}）`,
              };
            })}
          />
        </Form.Item>
        <Row gutter={12}>
          <Col span={12}>
            <Form.Item name="quantity" label="处置数量" rules={[{ required: true, message: '请输入数量' }]}>
              <InputNumber
                min={0.001}
                max={selectedFood?.quantity}
                precision={3}
                style={{ width: '100%' }}
                addonAfter={selectedFood?.unit}
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="method" label="处置方式" rules={[{ required: true }]}>
              <Select options={DisposalMethods.map((m) => ({ value: m, label: DisposalMethodLabels[m] }))} />
            </Form.Item>
          </Col>
        </Row>
        <Form.Item name="reason" label="申请说明">
          <Input.TextArea rows={3} maxLength={255} showCount placeholder="如：包装破损、准备捐赠等（可选）" />
        </Form.Item>
        {selectedFood && (
          <p style={{ color: '#faad14' }}>
            当前余量 {selectedFood.quantity} {selectedFood.unit}，超量或审批时已消耗将整单拒绝。
          </p>
        )}
        <Button type="primary" htmlType="submit" block disabled={!eligibleFoods.length}>
          提交申请
        </Button>
      </Form>
    </Drawer>
  );
}
