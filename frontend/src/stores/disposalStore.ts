import { useCallback, useState } from 'react';
import { listDisposals, type DisposalQuery } from '../api/disposal';
import type { DisposalApplication } from '../types';

// 临期处置申请 store：分页缓存申请列表（待处理与全部申请状态，刷新后可回读）
export function useDisposalStore() {
  const [applications, setApplications] = useState<DisposalApplication[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  const fetchList = useCallback(async (params: DisposalQuery) => {
    setLoading(true);
    try {
      const data = await listDisposals(params);
      setApplications(data.list);
      setTotal(data.total);
      return data;
    } finally {
      setLoading(false);
    }
  }, []);

  return { applications, total, loading, fetchList };
}
