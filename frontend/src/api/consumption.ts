import request from '../utils/request';
import type { ConsumptionRecord, PageData } from '../types';

export function listConsumptions(params: { family_id: number; page?: number; page_size?: number }): Promise<PageData<ConsumptionRecord>> {
  return request.get('/consumptions', { params });
}

export interface ConsumptionAnalysis {
  month: string;
  by_category: { category: string; count: number; total_quantity: number }[];
  top_consumed: { food_item_id: number; name: string; count: number; quantity: number }[];
}

export function consumptionAnalysis(params: { family_id: number; month?: string }): Promise<ConsumptionAnalysis> {
  return request.get('/consumptions/analysis', { params });
}
