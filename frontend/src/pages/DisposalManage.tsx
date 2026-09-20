import { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Modal, Space, Table, Tabs, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useFamilyStore } from '../stores/familyStore';
import { useDisposalStore } from '../stores/disposalStore';
import { usePagination } from '../hooks/usePagination';
import { approveDisposal, rejectDisposal } from '../api/disposal';
import { getFamilyDetail } from '../api/family';
import { useAuth } from '../stores/authStore';
import EmptyState from '../components/common/EmptyState';
import FreshnessBadge from '../components/common/FreshnessBadge';
import {
  DisposalMethodLabels,
  DisposalStatus,
  DisposalStatusLabels,
} from '../constants/food';
import { formatDateTime } from '../utils/dateFormat';
import type { DisposalApplication, FamilyMember } from '../types';

const statusColor: Record<string, string> = {
  [DisposalStatus.PENDING]: 'orange',
  [DisposalStatus.APPROVED]: 'green',
  [DisposalStatus.REJECTED]: 'red',
};

export default function DisposalManage() {
  const { user } = useAuth();
  const { currentFamily } = useFamilyStore();
  const { applications, total, loading, fetchList } = useDisposalStore();
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const [tab, setTab] = useState<string>(DisposalStatus.PENDING);
  const [members, setMembers] = useState<FamilyMember[]>([]);
  const [reviewTarget, setReviewTarget] = useState<{ record: DisposalApplication; action: 'approve' | 'reject' } | null>(null);
  const [reviewForm] = Form.useForm<{ note: string }>();

  // 当前用户在该家庭的角色：仅家庭管理员可批准/驳回
  const isFamilyAdmin = useMemo(
    () => members.some((m) => m.user_id === user?.id && m.role === 'admin'),
    [members, user?.id]
  );

  const load = useCallback(
    async (page = pagination.page, size = pagination.pageSize) => {
      if (!currentFamily) return;
      const data = await fetchList({
        family_id: currentFamily.id,
        status: tab === 'all' ? undefined : tab,
        page,
        page_size: size,
      });
      setTotal(data.total);
    },
    [currentFamily, tab, fetchList, pagination.page, pagination.pageSize, setTotal]
  );

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!currentFamily) return;
    getFamilyDetail(currentFamily.id)
      .then((detail) => setMembers(detail.members))
      .catch(() => undefined);
  }, [currentFamily, applications]);

  async function onReview() {
    if (!reviewTarget) return;
    const { note } = reviewForm.getFieldsValue();
    if (reviewTarget.action === 'approve') {
      await approveDisposal(reviewTarget.record.id, note ?? '');
      message.success('已批准：余量已扣减并生成消耗记录');
    } else {
      await rejectDisposal(reviewTarget.record.id, note ?? '');
      message.success('已驳回：申请已关闭，库存不变');
    }
    reviewForm.resetFields();
    setReviewTarget(null);
    load();
  }

  const columns: ColumnsType<DisposalApplication> = [
    {
      title: '食品',
      render: (_, r) => (
        <Space>
          {r.food_item?.name ?? `#${r.food_item_id}`}
          {r.food_item && <FreshnessBadge status={r.food_item.status} expiryDate={r.food_item.expiry_date} />}
        </Space>
      ),
    },
    {
      title: '申请数量',
      render: (_, r) => `${r.quantity} ${r.food_item?.unit ?? ''}`.trim(),
    },
    {
      title: '处置方式',
      dataIndex: 'method',
      render: (v: string) => <Tag>{DisposalMethodLabels[v] ?? v}</Tag>,
    },
    { title: '申请人', render: (_, r) => r.applicant?.name || r.applicant?.phone || `#${r.applicant_id}` },
    {
      title: '状态',
      dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'default'}>{DisposalStatusLabels[v] ?? v}</Tag>,
    },
    { title: '提交时间', dataIndex: 'created_at', render: (v: string) => formatDateTime(v) },
    {
      title: '审批人/备注',
      render: (_, r) =>
        r.status === DisposalStatus.PENDING ? (
          <span style={{ color: '#999' }}>待处理</span>
        ) : (
          <Space direction="vertical" size={0}>
            <span>{r.reviewer?.name || r.reviewer?.phone || (r.reviewer_id ? `#${r.reviewer_id}` : '-')}</span>
            {r.review_note && <span style={{ color: '#999' }}>{r.review_note}</span>}
            <span style={{ color: '#bbb', fontSize: 12 }}>{formatDateTime(r.reviewed_at)}</span>
          </Space>
        ),
    },
    {
      title: '操作',
      render: (_, r) =>
        r.status === DisposalStatus.PENDING ? (
          isFamilyAdmin ? (
            <Space>
              <a onClick={() => { reviewForm.resetFields(); setReviewTarget({ record: r, action: 'approve' }); }}>批准</a>
              <a
                style={{ color: '#ff4d4f' }}
                onClick={() => { reviewForm.resetFields(); setReviewTarget({ record: r, action: 'reject' }); }}
              >
                驳回
              </a>
            </Space>
          ) : (
            <span style={{ color: '#999' }}>仅家庭管理员可审批</span>
          )
        ) : (
          <span style={{ color: '#bbb' }}>已结案</span>
        ),
    },
  ];

  return (
    <Card
      size="small"
      title="临期处置申请"
      extra={<Button onClick={() => load()}>刷新</Button>}
    >
      <Tabs
        activeKey={tab}
        onChange={(key) => {
          setTab(key);
          onPageChange(1, pagination.pageSize);
        }}
        items={[
          { key: DisposalStatus.PENDING, label: '待处理' },
          { key: DisposalStatus.APPROVED, label: '已批准' },
          { key: DisposalStatus.REJECTED, label: '已驳回' },
          { key: 'all', label: '全部' },
        ]}
      />
      <Table
        rowKey="id"
        columns={columns}
        dataSource={applications}
        loading={loading}
        locale={{ emptyText: <EmptyState description="暂无处置申请" /> }}
        pagination={{
          current: pagination.page,
          pageSize: pagination.pageSize,
          total,
          showSizeChanger: true,
          onChange: onPageChange,
        }}
      />

      <Modal
        open={!!reviewTarget}
        title={reviewTarget?.action === 'approve' ? '批准处置申请' : '驳回处置申请'}
        okText={reviewTarget?.action === 'approve' ? '确认批准' : '确认驳回'}
        okButtonProps={reviewTarget?.action === 'reject' ? { danger: true } : undefined}
        onOk={() => reviewForm.submit()}
        onCancel={() => { reviewForm.resetFields(); setReviewTarget(null); }}
        destroyOnClose
      >
        {reviewTarget && (
          <Space direction="vertical" size={8} style={{ width: '100%' }}>
            <div>
              食品：<b>{reviewTarget.record.food_item?.name ?? `#${reviewTarget.record.food_item_id}`}</b>
              {'　'}数量：{reviewTarget.record.quantity} {reviewTarget.record.food_item?.unit ?? ''}
              {'　'}方式：{DisposalMethodLabels[reviewTarget.record.method] ?? reviewTarget.record.method}
            </div>
            {reviewTarget.action === 'approve' ? (
              <span style={{ color: '#999' }}>批准后将扣减余量并生成消耗记录，余量归零则食品标记为已消耗。</span>
            ) : (
              <span style={{ color: '#999' }}>驳回仅关闭申请，不扣减余量、不生成消耗记录。</span>
            )}
            <Form form={reviewForm} layout="vertical" onFinish={onReview} preserve={false}>
              <Form.Item name="note" label="备注" rules={[{ max: 255 }]}>
                <Input.TextArea rows={2} maxLength={255} placeholder={reviewTarget.action === 'approve' ? '批准备注（可选）' : '驳回原因（可选）'} />
              </Form.Item>
            </Form>
          </Space>
        )}
      </Modal>
    </Card>
  );
}
