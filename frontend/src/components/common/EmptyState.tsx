import { Empty } from 'antd';

interface Props {
  description?: string;
}

// 列表空状态（消耗记录/统计/家庭 等共用）
export default function EmptyState({ description = '暂无数据' }: Props) {
  return <Empty description={description} style={{ padding: 24 }} />;
}
