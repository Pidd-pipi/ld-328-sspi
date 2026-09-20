import type { FreshnessStatusValue } from '../constants/food';

export interface User {
  id: number;
  phone: string;
  name: string;
  avatar: string;
  role: string;
  created_at: string;
}

export interface FamilyGroup {
  id: number;
  name: string;
  owner_id: number;
  invite_code: string;
  created_at: string;
  owner?: User;
}

export interface FamilyMember {
  id: number;
  family_id: number;
  user_id: number;
  role: string;
  joined_at: string;
  user?: User;
}

export interface FoodItem {
  id: number;
  family_id: number;
  name: string;
  category: string;
  production_date?: string | null;
  shelf_life_days: number;
  quantity: number;
  unit: string;
  storage_location: string;
  opened_at?: string | null;
  expiry_date?: string | null;
  status: FreshnessStatusValue | string;
  image_url: string;
  creator_id: number;
  created_at: string;
  updated_at: string;
}

export interface ConsumptionRecord {
  id: number;
  food_item_id: number;
  quantity: number;
  consumed_at: string;
  user_id: number;
  user?: User;
  food_item?: FoodItem;
}

export interface Notification {
  id: number;
  family_id: number;
  food_item_id: number;
  type: string;
  title: string;
  content: string;
  is_read: boolean;
  send_at: string;
  read_at?: string | null;
}

export interface Recipe {
  id: number;
  name: string;
  ingredients: string;
  description: string;
  suitable_category: string;
}

export interface PageData<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}
