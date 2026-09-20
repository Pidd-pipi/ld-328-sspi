-- ld-328 CyFreshFood 易腐食品保质期追踪 初始化脚本（容器首次启动自动执行）
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    phone VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(50),
    avatar VARCHAR(255) DEFAULT '',
    role VARCHAR(20) DEFAULT 'member',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS family_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    owner_id BIGINT NOT NULL,
    invite_code VARCHAR(16) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT fk_family_owner FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS family_members (
    id BIGSERIAL PRIMARY KEY,
    family_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT fk_member_family FOREIGN KEY (family_id) REFERENCES family_groups(id),
    CONSTRAINT fk_member_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT uq_family_user UNIQUE (family_id, user_id)
);

CREATE TABLE IF NOT EXISTS food_items (
    id BIGSERIAL PRIMARY KEY,
    family_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(20) NOT NULL,
    production_date TIMESTAMPTZ,
    shelf_life_days INT DEFAULT 0,
    quantity DOUBLE PRECISION DEFAULT 0,
    unit VARCHAR(20) DEFAULT '份',
    storage_location VARCHAR(20) DEFAULT 'fridge',
    opened_at TIMESTAMPTZ,
    expiry_date TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'fresh',
    image_url VARCHAR(255) DEFAULT '',
    creator_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT fk_food_family FOREIGN KEY (family_id) REFERENCES family_groups(id)
);
CREATE INDEX IF NOT EXISTS idx_food_family_status ON food_items(family_id, status);

CREATE TABLE IF NOT EXISTS consumption_records (
    id BIGSERIAL PRIMARY KEY,
    food_item_id BIGINT NOT NULL,
    quantity DOUBLE PRECISION DEFAULT 0,
    consumed_at TIMESTAMPTZ DEFAULT now(),
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT fk_consume_food FOREIGN KEY (food_item_id) REFERENCES food_items(id),
    CONSTRAINT fk_consume_user FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_consume_food ON consumption_records(food_item_id);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    family_id BIGINT NOT NULL,
    food_item_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,
    title VARCHAR(100) DEFAULT '',
    content VARCHAR(500) DEFAULT '',
    is_read BOOLEAN DEFAULT FALSE,
    send_at TIMESTAMPTZ DEFAULT now(),
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT fk_notify_food FOREIGN KEY (food_item_id) REFERENCES food_items(id)
);
CREATE INDEX IF NOT EXISTS idx_notify_family ON notifications(family_id, is_read);

CREATE TABLE IF NOT EXISTS recipes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    ingredients VARCHAR(1000) DEFAULT '',
    description VARCHAR(2000) DEFAULT '',
    suitable_category VARCHAR(20) DEFAULT 'other',
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 种子数据（密码均为明文密码的 bcrypt 哈希：admin123 / member123）
INSERT INTO users (phone, password_hash, name, role) VALUES
('13800000001', '$2a$10$lZmZzCvRaYR.pAKTsPaXnuqcYHGKshT.v9FnKf2a7YCSMd6PmOPq.', '管理员', 'admin'),
('13800000002', '$2a$10$20Jb0QOClTzn6vCDWoduFe1qkmONUiykQToqNo7Q9kw0JqGzABq8.', '家庭成员', 'member')
ON CONFLICT (phone) DO NOTHING;

INSERT INTO family_groups (name, owner_id, invite_code)
SELECT '幸福之家', id, 'FAMILY01' FROM users WHERE phone='13800000001'
ON CONFLICT (invite_code) DO NOTHING;

INSERT INTO family_members (family_id, user_id, role)
SELECT g.id, u.id, CASE WHEN u.phone='13800000001' THEN 'admin' ELSE 'member' END
FROM users u JOIN family_groups g ON g.invite_code='FAMILY01'
WHERE u.phone IN ('13800000001','13800000002')
ON CONFLICT (family_id, user_id) DO NOTHING;

INSERT INTO food_items (family_id, name, category, quantity, unit, shelf_life_days, storage_location, expiry_date, status, creator_id)
SELECT g.id, f.name, f.category, f.quantity, f.unit, f.shelf_life_days, f.storage_location, now() + (f.days || ' days')::interval, f.status, u.id
FROM family_groups g
CROSS JOIN (VALUES
    ('鲜牛奶','dairy',2,'盒',5,'fridge',2,'fresh'),
    ('吐司面包','bakery',1,'袋',3,'pantry',1,'fresh'),
    ('鸡胸肉','fresh',3,'块',10,'freezer',7,'fresh'),
    ('熟食卤味','cooked',1,'份',2,'fridge',-1,'fresh')
) AS f(name, category, quantity, unit, shelf_life_days, storage_location, days, status)
JOIN users u ON u.phone='13800000001'
WHERE g.invite_code='FAMILY01';

INSERT INTO notifications (family_id, food_item_id, type, title, content)
SELECT g.id, fi.id, n.type, n.title, n.content
FROM family_groups g
CROSS JOIN (VALUES
    ('鲜牛奶','expiring','食品临近过期','鲜牛奶 即将过期，请及时处理。'),
    ('熟食卤味','expired','食品已过期','熟食卤味 已过期，请及时处理。')
) AS n(food_name, type, title, content)
JOIN food_items fi ON fi.family_id = g.id AND fi.name = n.food_name
WHERE g.invite_code='FAMILY01'
ON CONFLICT DO NOTHING;

INSERT INTO recipes (name, ingredients, description, suitable_category) VALUES
('牛奶燕麦粥', '牛奶、燕麦、蜂蜜', '将燕麦煮开后加入牛奶与蜂蜜，营养早餐。', 'dairy'),
('蒜香吐司', '吐司、黄油、蒜末', '吐司抹黄油蒜末烤至金黄。', 'bakery'),
('香煎鸡胸', '鸡胸肉、黑胡椒、橄榄油', '鸡胸肉腌制后煎熟，低脂高蛋白。', 'fresh'),
('凉拌卤味', '卤味、黄瓜、香菜', '卤味切片配黄瓜香菜凉拌。', 'cooked');
