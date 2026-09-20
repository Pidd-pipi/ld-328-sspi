import request from '../utils/request';
import type { DisposalApplication, PageData } from '../types';

export interface DisposalQuery {
  family_id: number;
  status?: string;
  food_item_id?: number;
  page?: number;
  page_size?: number;
}

// 处置申请列表（待处理与全部申请状态，刷新后可回读）
export function listDisposals(params: DisposalQuery): Promise<PageData<DisposalApplication>> {
  return request.get('/disposal-applications', { params });
}

export function getDisposalDetail(id: number): Promise<DisposalApplication> {
  return request.get(`/disposal-applications/${id}`);
}

// 家庭成员提交临期处置申请（数量 + 丢弃/食用/捐赠）
export function createDisposal(data: {
  food_item_id: number;
  quantity: number;
  method: string;
}): Promise<DisposalApplication> {
  return request.post('/disposal-applications', data);
}

// 家庭管理员批准：扣减余量、生成消耗记录并结案
export function approveDisposal(id: number, note?: string): Promise<DisposalApplication> {
  return request.post(`/disposal-applications/${id}/approve`, { note: note ?? '' });
}

// 家庭管理员驳回：仅关闭申请
export function rejectDisposal(id: number, note?: string): Promise<DisposalApplication> {
  return request.post(`/disposal-applications/${id}/reject`, { note: note ?? '' });
}
