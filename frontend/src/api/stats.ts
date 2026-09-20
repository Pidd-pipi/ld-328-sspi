import request from '../utils/request';
import type { FoodItem, Notification } from '../types';

export interface DashboardData {
  total_items: number;
  expiring_count: number;
  expired_count: number;
  consumed_count: number;
  unread_notify: number;
  member_count: number;
  by_category: { category: string; count: number; total_quantity: number }[];
  expiring_items: FoodItem[];
  expired_items: FoodItem[];
  recent_notify: Notification[];
}

export interface StatisticsData {
  category_share: { category: string; count: number; total_quantity: number }[];
  consumption_share: { category: string; count: number; total_quantity: number }[];
  top_purchased: { food_item_id: number; name: string; count: number; quantity: number }[];
  top_wasted: { food_item_id: number; name: string; count: number; quantity: number }[];
  waste_amount: number;
  month: string;
}

export function getDashboard(familyId: number): Promise<DashboardData> {
  return request.get('/stats/dashboard', { params: { family_id: familyId } });
}

export function getStatistics(familyId: number, month?: string): Promise<StatisticsData> {
  return request.get('/stats/statistics', { params: { family_id: familyId, month } });
}

export function exportStatisticsPDF(familyId: number, month?: string): string {
  const token = localStorage.getItem('cyfreshfood_token');
  const params = new URLSearchParams({ family_id: String(familyId) });
  if (month) params.set('month', month);
  return `/api/v1/stats/statistics/export?${params.toString()}&token=${token ?? ''}`;
}
