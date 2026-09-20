import { useCallback, useState } from 'react';
import { listNotifications, unreadCount } from '../api/notification';
import type { Notification } from '../types';

// 通知 store：临期/过期提醒
export function useNotificationStore() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unread, setUnread] = useState(0);
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async (familyId: number) => {
    if (!familyId) return;
    setLoading(true);
    try {
      const data = await listNotifications({ family_id: familyId, page_size: 10 });
      setNotifications(data.list);
      const c = await unreadCount(familyId);
      setUnread(c.unread);
    } finally {
      setLoading(false);
    }
  }, []);

  return { notifications, unread, loading, refresh };
}
