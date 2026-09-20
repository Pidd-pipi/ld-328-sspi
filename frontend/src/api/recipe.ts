import request from '../utils/request';
import type { FoodItem, Recipe } from '../types';

export interface Recommendation {
  food_items: FoodItem[];
  recipes: Recipe[];
  summary: { expiring_count: number; expired_count: number; categories: string[] };
}

export function getRecommendations(familyId: number): Promise<Recommendation> {
  return request.get('/recipes/recommendations', { params: { family_id: familyId } });
}

export function listRecipes(): Promise<Recipe[]> {
  return request.get('/recipes');
}
