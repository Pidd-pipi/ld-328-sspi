import { useCallback, useState } from 'react';
import { listFoods, type FoodQuery } from '../api/foodItem';
import type { FoodItem } from '../types';

// 食品 store：分页缓存食品列表
export function useFoodStore() {
  const [items, setItems] = useState<FoodItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  const fetchList = useCallback(async (params: FoodQuery) => {
    setLoading(true);
    try {
      const data = await listFoods(params);
      setItems(data.list);
      setTotal(data.total);
      return data;
    } finally {
      setLoading(false);
    }
  }, []);

  return { items, total, loading, fetchList };
}
