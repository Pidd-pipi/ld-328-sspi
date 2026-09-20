import { useEffect, useState } from 'react';
import { getFamilyDetail } from '../api/family';
import { useFamilyStore } from '../stores/familyStore';
import { useAuth } from '../stores/authStore';
import type { FamilyMember } from '../types';

// 家庭角色 hook：返回当前用户在当前家庭中的成员记录与是否家庭管理员（审批权限）。
export function useFamilyRole() {
  const { currentFamily } = useFamilyStore();
  const { user } = useAuth();
  const [members, setMembers] = useState<FamilyMember[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!currentFamily) {
      setMembers([]);
      return;
    }
    let active = true;
    setLoading(true);
    getFamilyDetail(currentFamily.id)
      .then((detail) => {
        if (active) setMembers(detail.members);
      })
      .catch(() => undefined)
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [currentFamily]);

  const myMember = members.find((m) => m.user_id === user?.id);
  const isFamilyAdmin = myMember?.role === 'admin';
  return { members, myMember, isFamilyAdmin, loading };
}
