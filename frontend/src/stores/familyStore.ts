import { useCallback, useSyncExternalStore } from 'react';
import { listMyFamilies } from '../api/family';
import type { FamilyGroup } from '../types';

// 家庭组 store：模块级共享状态（useSyncExternalStore），Shell 与各页面读取同一份数据
let families: FamilyGroup[] = [];
let currentFamily: FamilyGroup | null = null;
const listeners = new Set<() => void>();

function emit(): void {
  listeners.forEach((l) => l());
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getFamilies(): FamilyGroup[] {
  return families;
}

function getCurrent(): FamilyGroup | null {
  return currentFamily;
}

export function useFamilyStore() {
  const familiesSnap = useSyncExternalStore(subscribe, getFamilies, getFamilies);
  const currentSnap = useSyncExternalStore(subscribe, getCurrent, getCurrent);

  const refresh = useCallback(async (): Promise<FamilyGroup[]> => {
    const data = await listMyFamilies();
    families = data;
    if (data.length > 0 && !currentFamily) {
      currentFamily = data[0];
    }
    emit();
    return data;
  }, []);

  const selectFamily = useCallback((f: FamilyGroup) => {
    currentFamily = f;
    emit();
  }, []);

  const setCurrent = useCallback((f: FamilyGroup) => {
    currentFamily = f;
    emit();
  }, []);

  return { families: familiesSnap, currentFamily: currentSnap, refresh, selectFamily, setCurrent };
}
