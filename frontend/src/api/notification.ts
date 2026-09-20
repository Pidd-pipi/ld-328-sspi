import request from '../utils/request';
import type { Notification, PageData } from '../types';

export function listNotifications(params: { family_id: number; unread_only?: boolean; page?: number; page_size?: number }): Promise<PageData<Notification>> {
  return request.get('/notifications', { params });
}

export function markRead(id: number): Promise<void> {
  return request.put(`/notifications/${id}/read`);
}

export function markAllRead(familyId: number): Promise<void> {
  return request.put('/notifications/read-all', null, { params: { family_id: familyId } });
}

export function unreadCount(familyId: number): Promise<{ unread: number }> {
  return request.get('/notifications/unread-count', { params: { family_id: familyId } });
}
