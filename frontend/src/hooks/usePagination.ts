import { useCallback, useState } from 'react';

interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}

// 通用分页 hook
export function usePagination(initialPage = 1, initialPageSize = 10) {
  const [pagination, setPagination] = useState<PaginationState>({
    page: initialPage,
    pageSize: initialPageSize,
    total: 0,
  });

  const setTotal = useCallback((total: number) => {
    setPagination((prev) => ({ ...prev, total }));
  }, []);

  const onPageChange = useCallback((page: number, pageSize: number) => {
    setPagination({ page, pageSize, total: pagination.total });
  }, [pagination.total]);

  return { pagination, setTotal, onPageChange };
}
