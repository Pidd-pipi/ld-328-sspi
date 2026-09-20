import { Progress } from 'antd';
import { ExpiringThresholdDays } from '../../constants/food';
import { calculateRemainingDays, remainingDaysText } from '../../utils/calculateRemainingDays';

interface Props {
  expiryDate?: string | null;
}

// 剩余天数进度条（看板专用）
export default function RemainingDaysBar({ expiryDate }: Props) {
  const days = calculateRemainingDays(expiryDate);
  const pct = days < 0 ? 0 : Math.min(100, Math.round((days / 30) * 100));
  const color = days < 0 ? '#ff4d4f' : days <= ExpiringThresholdDays ? '#faad14' : '#52c41a';
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <Progress percent={pct} size="small" strokeColor={color} style={{ flex: 1, marginBottom: 0 }} />
      <span style={{ fontSize: 12, color }}>{remainingDaysText(days)}</span>
    </div>
  );
}
