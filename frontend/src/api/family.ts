import request from '../utils/request';
import type { FamilyGroup, FamilyMember } from '../types';

export function createFamily(name: string): Promise<FamilyGroup> {
  return request.post('/family-groups', { name });
}

export function listMyFamilies(): Promise<FamilyGroup[]> {
  return request.get('/family-groups');
}

export function getFamilyDetail(id: number): Promise<{ group: FamilyGroup; members: FamilyMember[] }> {
  return request.get(`/family-groups/${id}`);
}

export function joinFamily(inviteCode: string): Promise<FamilyGroup> {
  return request.post('/family-groups/join', { invite_code: inviteCode });
}

export function inviteMember(familyId: number, userId: number): Promise<FamilyMember> {
  return request.post(`/family-groups/${familyId}/invite`, { user_id: userId });
}

export function setMemberRole(familyId: number, memberId: number, role: string): Promise<void> {
  return request.put(`/family-groups/${familyId}/members/${memberId}/role`, { role });
}

export function removeMember(familyId: number, memberId: number): Promise<void> {
  return request.delete(`/family-groups/${familyId}/members/${memberId}`);
}
