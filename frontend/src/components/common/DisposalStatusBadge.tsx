import { Tag } from 'antd';
import { DisposalStatus, DisposalStatusLabels } from '../../constants/food';

const colorMap: Record<string, string> = {
  [DisposalStatus.PENDING]: 'processing',
  [DisposalStatus.APPROVED]: 'success',
  [DisposalStatus.REJECTED]: 'default',
};

// 处置申请状态徽标：食品管理与处置申请页共用
export default function DisposalStatusBadge({ status }: { status: string }) {
  return <Tag color={colorMap[status] ?? 'default'}>{DisposalStatusLabels[status] ?? status}</Tag>;
}
