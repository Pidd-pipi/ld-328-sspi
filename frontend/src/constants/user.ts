// 与 backend/internal/constants/user.go 保持一致
export const UserRole = {
  ADMIN: 'admin',
  MEMBER: 'member',
} as const;
export type UserRoleValue = typeof UserRole[keyof typeof UserRole];

export const UserRoleLabels: Record<string, string> = {
  [UserRole.ADMIN]: '管理员',
  [UserRole.MEMBER]: '成员',
};

export const FamilyMemberRole = {
  ADMIN: 'admin',
  MEMBER: 'member',
} as const;
