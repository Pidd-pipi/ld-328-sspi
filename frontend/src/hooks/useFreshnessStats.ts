import { useMemo } from 'react';
import type { FoodItem } from '../types';
import { computeFreshness } from '../utils/calculateRemainingDays';
import { FreshnessStatus } from '../constants/food';

export interface FreshnessStats {
  fresh: number;
  expiring: number;
  expired: number;
  consumed: number;
  total: number;
  expiringRatio: number;
  expiredRatio: number;
}

// 新鲜度统计与看板着色
export function useFreshnessStats(items: FoodItem[]): FreshnessStats {
  return useMemo(() => {
    const stats: FreshnessStats = { fresh: 0, expiring: 0, expired: 0, consumed: 0, total: items.length, expiringRatio: 0, expiredRatio: 0 };
    for (const item of items) {
      const status = computeFreshness(item.status, item.expiry_date);
      if (status === FreshnessStatus.EXPIRING) stats.expiring++;
      else if (status === FreshnessStatus.EXPIRED) stats.expired++;
      else if (status === FreshnessStatus.CONSUMED) stats.consumed++;
      else stats.fresh++;
    }
    stats.expiringRatio = stats.total ? Math.round((stats.expiring / stats.total) * 100) : 0;
    stats.expiredRatio = stats.total ? Math.round((stats.expired / stats.total) * 100) : 0;
    return stats;
  }, [items]);
}
