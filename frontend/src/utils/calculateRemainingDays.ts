import { ExpiringThresholdDays, FreshnessStatus } from '../constants/food';

// 根据到期日计算剩余天数（与后端 util/food_calculator.go 保持一致）
export function calculateRemainingDays(expiryDate?: string | null): number {
  if (!expiryDate) return 365 * 10;
  const exp = new Date(expiryDate);
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const target = new Date(exp.getFullYear(), exp.getMonth(), exp.getDate());
  return Math.round((target.getTime() - today.getTime()) / 86400000);
}

export function computeFreshness(status: string, expiryDate?: string | null): string {
  if (status === FreshnessStatus.CONSUMED) return FreshnessStatus.CONSUMED;
  const days = calculateRemainingDays(expiryDate);
  if (days < 0) return FreshnessStatus.EXPIRED;
  if (days <= ExpiringThresholdDays) return FreshnessStatus.EXPIRING;
  return FreshnessStatus.FRESH;
}

export function remainingDaysText(days: number): string {
  if (days < 0) return `已过期 ${-days} 天`;
  if (days === 0) return '今天到期';
  if (days <= ExpiringThresholdDays) return `剩余 ${days} 天（临期）`;
  return `剩余 ${days} 天`;
}
