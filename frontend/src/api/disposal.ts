import request from '../utils/request';
import type { DisposalRequest, PageData } from '../types';

export interface DisposalQuery {
  family_id: number;
  food_item_id?: number;
  status?: string;
  page?: number;
  page_size?: number;
}

// 家庭成员提交临期/过期食品处置申请
export function createDisposal(data: {
  food_item_id: number;
  quantity: number;
  method: string;
  reason?: string;
}): Promise<DisposalRequest> {
  return request.post('/disposals', data);
}

// 处置申请列表（待处理与全部申请状态，刷新后可回读）
export function listDisposals(params: DisposalQuery): Promise<PageData<DisposalRequest>> {
  return request.get('/disposals', { params });
}

export function getDisposalDetail(id: number): Promise<DisposalRequest> {
  return request.get(`/disposals/${id}`);
}

// 仅家庭管理员：批准（扣减余量、生成消耗记录并结案）
export function approveDisposal(id: number): Promise<DisposalRequest> {
  return request.post(`/disposals/${id}/approve`, {});
}

// 仅家庭管理员：驳回（只关闭申请，不改库存）
export function rejectDisposal(id: number, remark?: string): Promise<DisposalRequest> {
  return request.post(`/disposals/${id}/reject`, { remark });
}
