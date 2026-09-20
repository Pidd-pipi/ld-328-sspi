# CyFreshFood（易腐食品保质期追踪）

> 项目类型：全栈 Web 应用（农业与生活服务）

面向家庭和小型餐饮的食品保质期管理工具，帮助用户记录食品入库信息、自动计算保质期剩余天数、及时提醒临期食品，减少食物浪费。支持家庭多成员共享、消耗记录、分类统计报表与智能食谱推荐。

## 快速启动（Docker Compose 一键部署，首选）

```bash
# 1. 首次启动前复制环境变量
cp .env.example .env

# 2. 启动全部服务（前端 + 后端 + PostgreSQL + Redis）
docker compose up -d

# 3. 查看健康状态
docker compose ps
```

访问地址：

- 前端：http://localhost:18628
- 后端 API：http://localhost:19628
- 健康检查：http://localhost:19628/healthz

演示账号：

| 角色 | 账号 | 密码 |
| --- | --- | --- |
| 管理员 | 13800000001 | admin123 |
| 成员 | 13800000002 | member123 |

## 本地开发

```bash
# 前端（React 18 + TS + Vite + Ant Design + ECharts）
cd frontend
npm install
npm run dev        # http://localhost:18628

# 后端（Go）
cd backend
go mod tidy
go run ./cmd/server
```

## 技术栈

| 层次 | 技术 |
| --- | --- |
| 前端 | React 18 + TypeScript + Vite + Ant Design 5 |
| 图表 | ECharts（echarts-for-react） |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 15 |
| 缓存/任务 | Redis（go-redis/v9，临期扫描分布式锁） |
| 认证 | JWT (github.com/golang-jwt/jwt/v5) + RBAC |
| 校验 | github.com/go-playground/validator/v10 |
| PDF | github.com/jung-kurt/gofpdf（统计报表导出） |

## 项目目录结构

```
ld-328/
├── docker-compose.yml        # 前端/后端/PostgreSQL/Redis 编排
├── .env.example              # 环境变量示例
├── database/init.sql         # 建表 + 种子数据（首次启动自动执行）
├── frontend/
│   ├── Dockerfile            # Node 构建 + Nginx 托管
│   ├── nginx.conf            # SPA 路由 + /api 反代到 backend:8080
│   └── src/
│       ├── api/              # user/family/foodItem/consumption/notification/recipe/stats
│       ├── stores/           # authStore/userStore/familyStore/foodStore/notificationStore
│       ├── components/common/# FreshnessBadge/RemainingDaysBar/FoodCard/MemberAvatar/EmptyState/RoleGuard/ErrorBoundary
│       ├── hooks/            # useFreshnessStats/usePagination
│       ├── pages/            # Dashboard/FoodManage/ConsumptionManage/Statistics/FamilyManage/Recommendations/Profile/Login
│       ├── router/           # index.tsx + guards.tsx
│       ├── utils/            # calculateRemainingDays/dateFormat/request
│       └── constants/        # food/user/errorCodes
└── backend/
    ├── cmd/server/main.go    # 入口：装配依赖、启动 Gin 与临期扫描
    └── internal/
        ├── config/           # 环境变量解析
        ├── model/            # user/family_group/family_member/food_item/consumption_record/notification/recipe/stats
        ├── repository/       # 按实体分文件
        ├── service/          # 按实体分文件 + reminder_scheduler/notification_sender/stats
        ├── handler/          # 按实体分文件
        ├── router/           # router.go + 按实体分文件
        ├── middleware/       # auth/rbac/error_handler/rate_limiter/request_logger
        ├── dto/              # 请求/响应结构体
        ├── constants/        # food/user/reminder/error_codes/log_templates/messages
        └── util/             # jwt/logger/formatters/food_calculator/scheduler/app_error/response/pdf
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | cyfreshfood | Docker Compose 项目短名 |
| DB_NAME | cyfreshfood_db | 数据库名 |
| DB_USER | cyfreshfood_user | 数据库用户 |
| DB_PASSWORD | cyfreshfood_pwd | 数据库密码 |
| REDIS_HOST | redis | Redis 服务名 |
| REDIS_PORT | 6379 | Redis 端口 |
| JWT_SECRET | change_me_to_a_long_random_string | JWT 签名密钥（生产替换） |
| APP_CORS_ORIGINS | http://localhost:18628 | CORS 来源白名单（逗号分隔，生产请显式配置，不允许 `*`） |
| FRONTEND_PORT | 18628 | 前端对外端口 |
| BACKEND_PORT | 19628 | 后端对外端口 |
| DB_PORT | 5432 | 数据库对外端口 |

## Docker 部署说明

- 端口映射：前端 `18628:80`、后端 `19628:8080`、数据库 `5432:5432`、Redis `6379:6379`
- 数据卷：`db_data` 命名卷持久化 PostgreSQL 数据
- 依赖顺序：backend `depends_on` db/redis 健康后启动；frontend 等待 backend 健康
- 常见问题：
  - 端口冲突：修改 `.env` 中的 `FRONTEND_PORT` / `BACKEND_PORT` / `DB_PORT`
  - 数据重置：`docker compose down -v` 后重新 `docker compose up -d`
  - 中文目录：容器名使用 `COMPOSE_PROJECT_NAME` 前缀，不依赖目录名

## API 接口清单

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/healthz` | 健康检查 |
| POST | `/api/v1/auth/register` | 注册（返回 JWT） |
| POST | `/api/v1/auth/login` | 登录（返回 JWT） |
| GET | `/api/v1/users/me` | 当前用户资料 |
| PUT | `/api/v1/users/me` | 更新当前用户资料 |
| POST | `/api/v1/family-groups` | 创建家庭组 |
| GET | `/api/v1/family-groups` | 我加入的家庭组列表 |
| GET | `/api/v1/family-groups/:id` | 家庭组详情 |
| POST | `/api/v1/family-groups/join` | 邀请码加入家庭组 |
| POST | `/api/v1/family-groups/:id/invite` | 邀请成员 |
| PUT | `/api/v1/family-groups/:id/members/:memberId/role` | 修改成员家庭角色 |
| DELETE | `/api/v1/family-groups/:id/members/:memberId` | 移除成员 |
| POST | `/api/v1/foods` | 录入食品 |
| GET | `/api/v1/foods` | 查询食品列表 |
| POST | `/api/v1/foods/csv-import` | CSV 批量导入 |
| GET | `/api/v1/foods/:id` | 食品详情 |
| PUT | `/api/v1/foods/:id` | 编辑食品 |
| DELETE | `/api/v1/foods/:id` | 删除食品 |
| POST | `/api/v1/foods/:id/consume` | 记录食品消耗 |
| GET | `/api/v1/consumptions` | 家庭消耗记录分页列表 |
| GET | `/api/v1/consumptions/analysis` | 消耗频率分析 |
| GET | `/api/v1/notifications` | 通知列表 |
| GET | `/api/v1/notifications/unread-count` | 未读通知数量 |
| PUT | `/api/v1/notifications/read-all` | 全部标记已读 |
| PUT | `/api/v1/notifications/:id/read` | 单条标记已读 |
| GET | `/api/v1/recipes/recommendations` | 智能食谱推荐 |
| GET | `/api/v1/recipes` | 食谱列表 |
| GET | `/api/v1/recipes/:id` | 食谱详情 |
| GET | `/api/v1/stats/dashboard` | 看板统计 |
| GET | `/api/v1/stats/statistics` | 分类统计 |
| GET | `/api/v1/stats/statistics/export` | 统计报表 PDF 导出 |

> 除 `/healthz` 与注册/登录外，其余接口需携带 `Authorization: Bearer <JWT>`；响应统一为 `{code, message, data}`，请求日志与响应头包含 `X-Request-ID`。

## 枚举出现位置清单

### FoodCategory（食品类别：fresh/dairy/cooked/bakery/frozen/other）

- 后端：`backend/internal/constants/food.go`（定义）、`backend/internal/model/food_item.go`（模型）、`backend/internal/service/food_item_service.go`（校验/导入）、`backend/internal/service/reminder_service.go`（通知文案）、`backend/internal/util/formatters.go`（类别文本）、`backend/internal/constants/log_templates.go`（日志）、`backend/internal/constants/messages.go`（文案）、`backend/internal/constants/error_codes.go`（错误码）
- 前端：`frontend/src/constants/food.ts`（定义）、`frontend/src/types/index.ts`（类型）、`frontend/src/components/common/FoodCard.tsx`（展示）、`frontend/src/components/common/FreshnessBadge.tsx`（状态）、`frontend/src/pages/Dashboard.tsx`（看板筛选）、`frontend/src/pages/FoodManage.tsx`（表单/筛选）、`frontend/src/pages/Statistics.tsx`（统计分组）、`frontend/src/pages/ConsumptionManage.tsx`（分析）、`frontend/src/pages/Recommendations.tsx`（推荐）、`frontend/src/api/foodItem.ts`（查询参数）

### FreshnessStatus（新鲜度状态：fresh/expiring/expired/consumed）

- 后端：`backend/internal/constants/food.go`（定义）、`backend/internal/model/food_item.go`（模型）、`backend/internal/service/food_item_service.go`（状态机）、`backend/internal/service/reminder_service.go`（扫描状态流转）、`backend/internal/util/food_calculator.go`（计算）、`backend/internal/util/formatters.go`（状态文本）、`backend/internal/constants/log_templates.go`（日志）、`backend/internal/constants/error_codes.go`（错误码）、`backend/internal/repository/food_item_repository.go`（查询）、`backend/internal/service/stats_service.go`（看板分组）
- 前端：`frontend/src/constants/food.ts`（定义）、`frontend/src/types/index.ts`（类型）、`frontend/src/components/common/FreshnessBadge.tsx`（着色）、`frontend/src/components/common/RemainingDaysBar.tsx`（进度条）、`frontend/src/components/common/FoodCard.tsx`（展示）、`frontend/src/hooks/useFreshnessStats.ts`（统计）、`frontend/src/utils/calculateRemainingDays.ts`（计算）、`frontend/src/pages/Dashboard.tsx`（看板着色）、`frontend/src/pages/FoodManage.tsx`（筛选）、`frontend/src/pages/Statistics.tsx`（浪费统计）

### UserRole（用户角色：admin/member）

- 后端：`backend/internal/constants/user.go`、`backend/internal/model/user.go`、`backend/internal/model/family_member.go`、`backend/internal/middleware/rbac.go`、`backend/internal/router/family_groups.go`、`backend/internal/service/family_group_service.go`、`backend/internal/util/formatters.go`、`backend/internal/constants/error_codes.go`
- 前端：`frontend/src/constants/user.ts`、`frontend/src/stores/authStore.ts`、`frontend/src/router/guards.tsx`、`frontend/src/components/common/RoleGuard.tsx`、`frontend/src/pages/FamilyManage.tsx`

## License

MIT
