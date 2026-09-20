// 与 backend/internal/constants/food.go 保持一致（屎山耦合：新增枚举需前后端同步 ≥10 处）
export const FoodCategory = {
  FRESH: 'fresh',
  DAIRY: 'dairy',
  COOKED: 'cooked',
  BAKERY: 'bakery',
  FROZEN: 'frozen',
  OTHER: 'other',
} as const;
export type FoodCategoryValue = typeof FoodCategory[keyof typeof FoodCategory];

export const FoodCategories: string[] = [
  FoodCategory.FRESH, FoodCategory.DAIRY, FoodCategory.COOKED,
  FoodCategory.BAKERY, FoodCategory.FROZEN, FoodCategory.OTHER,
];

export const FoodCategoryLabels: Record<string, string> = {
  [FoodCategory.FRESH]: '生鲜',
  [FoodCategory.DAIRY]: '乳制品',
  [FoodCategory.COOKED]: '熟食',
  [FoodCategory.BAKERY]: '烘焙',
  [FoodCategory.FROZEN]: '冷冻',
  [FoodCategory.OTHER]: '其他',
};

export const FreshnessStatus = {
  FRESH: 'fresh',
  EXPIRING: 'expiring',
  EXPIRED: 'expired',
  CONSUMED: 'consumed',
} as const;
export type FreshnessStatusValue = typeof FreshnessStatus[keyof typeof FreshnessStatus];

export const FreshnessStatusLabels: Record<string, string> = {
  [FreshnessStatus.FRESH]: '充裕',
  [FreshnessStatus.EXPIRING]: '临期',
  [FreshnessStatus.EXPIRED]: '已过期',
  [FreshnessStatus.CONSUMED]: '已消耗',
};

export const ExpiringThresholdDays = 3;

export const StorageLocationLabels: Record<string, string> = {
  fridge: '冰箱',
  pantry: '储藏室',
  freezer: '冷冻室',
  counter: '台面',
  other: '其他',
};

// 临期处置方式：与 backend/internal/constants/disposal.go 保持一致
export const DisposalMethod = {
  DISCARD: 'discard',
  EAT: 'eat',
  DONATE: 'donate',
} as const;
export type DisposalMethodValue = typeof DisposalMethod[keyof typeof DisposalMethod];

export const DisposalMethods: string[] = [
  DisposalMethod.DISCARD, DisposalMethod.EAT, DisposalMethod.DONATE,
];

export const DisposalMethodLabels: Record<string, string> = {
  [DisposalMethod.DISCARD]: '丢弃',
  [DisposalMethod.EAT]: '食用',
  [DisposalMethod.DONATE]: '捐赠',
};

// 处置申请状态：待处理 / 已批准 / 已驳回（终态）
export const DisposalStatus = {
  PENDING: 'pending',
  APPROVED: 'approved',
  REJECTED: 'rejected',
} as const;
export type DisposalStatusValue = typeof DisposalStatus[keyof typeof DisposalStatus];

export const DisposalStatusLabels: Record<string, string> = {
  [DisposalStatus.PENDING]: '待处理',
  [DisposalStatus.APPROVED]: '已批准',
  [DisposalStatus.REJECTED]: '已驳回',
};

// 可发起处置申请的新鲜度状态：仅临期 / 已过期
export const DisposableFreshnessStatuses: string[] = [
  FreshnessStatus.EXPIRING, FreshnessStatus.EXPIRED,
];
