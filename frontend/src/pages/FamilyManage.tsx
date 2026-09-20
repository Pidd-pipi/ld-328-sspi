import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Descriptions, Form, Input, InputNumber, Modal, Select, Space, Table, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useFamilyStore } from '../stores/familyStore';
import { createFamily, getFamilyDetail, inviteMember, joinFamily, removeMember, setMemberRole } from '../api/family';
import type { FamilyMember } from '../types';
import MemberAvatar from '../components/common/MemberAvatar';
import EmptyState from '../components/common/EmptyState';
import { UserRoleLabels } from '../constants/user';
import { formatDateTime } from '../utils/dateFormat';
import { useAuth } from '../stores/authStore';

export default function FamilyManage() {
  const { user } = useAuth();
  const { families, currentFamily, refresh, selectFamily, setCurrent } = useFamilyStore();
  const [members, setMembers] = useState<FamilyMember[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [joinOpen, setJoinOpen] = useState(false);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [form] = Form.useForm();
  const [inviteForm] = Form.useForm();

  const loadMembers = useCallback(async () => {
    if (!currentFamily) return;
    const detail = await getFamilyDetail(currentFamily.id);
    setMembers(detail.members);
  }, [currentFamily]);

  useEffect(() => { loadMembers().catch(() => undefined); }, [loadMembers]);

  async function onCreate(values: { name: string }) {
    const g = await createFamily(values.name);
    message.success('家庭组创建成功');
    setCreateOpen(false);
    const data = await refresh();
    const created = data.find((x) => x.id === g.id) ?? g;
    setCurrent(created);
  }

  async function onJoin(values: { invite_code: string }) {
    const g = await joinFamily(values.invite_code);
    message.success('加入成功');
    setJoinOpen(false);
    const data = await refresh();
    const joined = data.find((x) => x.id === g.id) ?? g;
    setCurrent(joined);
  }

  async function onInvite(values: { user_id: number }) {
    if (!currentFamily) return;
    await inviteMember(currentFamily.id, values.user_id);
    message.success('邀请成功');
    setInviteOpen(false);
    inviteForm.resetFields();
    loadMembers();
  }

  async function onRoleChange(member: FamilyMember, role: string) {
    if (!currentFamily) return;
    await setMemberRole(currentFamily.id, member.id, role);
    message.success('权限已更新');
    loadMembers();
  }

  async function onRemove(member: FamilyMember) {
    if (!currentFamily) return;
    await removeMember(currentFamily.id, member.id);
    message.success('已移除成员');
    loadMembers();
  }

  const columns: ColumnsType<FamilyMember> = [
    { title: '成员', render: (_, r) => <Space><MemberAvatar user={r.user} />{r.user?.name || r.user?.phone}</Space> },
    { title: '角色', dataIndex: 'role', render: (v) => UserRoleLabels[v] ?? v },
    { title: '加入时间', dataIndex: 'joined_at', render: (v) => formatDateTime(v) },
    { title: '操作', render: (_, r) => (
      <Space>
        {r.user_id !== user?.id && (
          <Select size="small" style={{ width: 100 }} value={r.role} options={[['admin','管理员'],['member','成员']].map(([value,label]) => ({ value, label }))} onChange={(v) => onRoleChange(r, v)} />
        )}
        {r.user_id !== user?.id && <a style={{ color: '#ff4d4f' }} onClick={() => Modal.confirm({ title: '确认移除该成员？', onOk: () => onRemove(r) })}>移除</a>}
      </Space>
    ) },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Space wrap>
          <Button type="primary" onClick={() => setCreateOpen(true)}>创建家庭组</Button>
          <Button onClick={() => setJoinOpen(true)}>加入家庭组</Button>
          <Button disabled={!currentFamily} onClick={() => setInviteOpen(true)}>邀请成员</Button>
          <Select style={{ width: 220 }} placeholder="切换家庭组" value={currentFamily?.id} options={families.map((f) => ({ value: f.id, label: f.name }))} onChange={(id) => { const f = families.find((x) => x.id === id); if (f) selectFamily(f); }} />
        </Space>
      </Card>

      {currentFamily ? (
        <>
          <Card size="small" title="家庭信息">
            <Descriptions column={3} size="small">
              <Descriptions.Item label="家庭组">{currentFamily.name}</Descriptions.Item>
              <Descriptions.Item label="邀请码"><b>{currentFamily.invite_code}</b></Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(currentFamily.created_at)}</Descriptions.Item>
            </Descriptions>
          </Card>
          <Card size="small" title="家庭成员">
            <Table rowKey="id" columns={columns} dataSource={members} pagination={false} locale={{ emptyText: <EmptyState description="暂无成员" /> }} />
          </Card>
        </>
      ) : <EmptyState description="请先创建或加入家庭组" />}

      <Modal open={createOpen} title="创建家庭组" onOk={() => form.submit()} onCancel={() => setCreateOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={onCreate}>
          <Form.Item name="name" label="家庭组名称" rules={[{ required: true }]}><Input /></Form.Item>
        </Form>
      </Modal>

      <Modal open={joinOpen} title="加入家庭组" onOk={() => form.submit()} onCancel={() => setJoinOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={onJoin}>
          <Form.Item name="invite_code" label="邀请码" rules={[{ required: true }]}><Input placeholder="如 FAMILY01" /></Form.Item>
        </Form>
      </Modal>

      <Modal open={inviteOpen} title="邀请成员（按用户 ID）" onOk={() => inviteForm.submit()} onCancel={() => setInviteOpen(false)} destroyOnClose>
        <Form form={inviteForm} layout="vertical" onFinish={onInvite}>
          <Form.Item name="user_id" label="用户 ID" rules={[{ required: true }]}><InputNumberW /></Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}

function InputNumberW() {
  return <InputNumber min={1} style={{ width: '100%' }} />;
}
