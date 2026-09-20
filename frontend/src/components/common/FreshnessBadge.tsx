import { Tag } from 'antd';
import { FreshnessStatus, FreshnessStatusLabels } from '../../constants/food';
import { computeFreshness } from '../../utils/calculateRemainingDays';

interface Props {
  status?: string;
  expiryDate?: string | null;
}

const colorMap: Record<string, string> = {
  [FreshnessStatus.FRESH]: 'green',
  [FreshnessStatus.EXPIRING]: 'orange',
  [FreshnessStatus.EXPIRED]: 'red',
  [FreshnessStatus.CONSUMED]: 'default',
};

// 看板/食品管理/消耗记录/推荐 共用组件
export default function FreshnessBadge({ status, expiryDate }: Props) {
  const final = computeFreshness(status ?? 'fresh', expiryDate);
  return <Tag color={colorMap[final] ?? 'default'}>{FreshnessStatusLabels[final] ?? final}</Tag>;
}
