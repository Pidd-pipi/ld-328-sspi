import request from '../utils/request';
import type { ConsumptionRecord, FoodItem, PageData } from '../types';

export interface FoodQuery {
  family_id: number;
  category?: string;
  status?: string;
  storage_location?: string;
  keyword?: string;
  page?: number;
  page_size?: number;
}

export function listFoods(params: FoodQuery): Promise<PageData<FoodItem>> {
  return request.get('/foods', { params });
}

export function getFoodDetail(id: number): Promise<{ item: FoodItem; consumption_records: ConsumptionRecord[] }> {
  return request.get(`/foods/${id}`);
}

export function createFood(data: Partial<FoodItem>): Promise<FoodItem> {
  return request.post('/foods', data);
}

export function updateFood(id: number, data: Partial<FoodItem>): Promise<FoodItem> {
  return request.put(`/foods/${id}`, data);
}

export function deleteFood(id: number): Promise<void> {
  return request.delete(`/foods/${id}`);
}

export function consumeFood(id: number, quantity: number): Promise<ConsumptionRecord> {
  return request.post(`/foods/${id}/consume`, { quantity });
}

export function importFoodsCSV(familyId: number, csvText: string): Promise<{ imported: number }> {
  return request.post('/foods/csv-import', { family_id: familyId, csv_text: csvText });
}
