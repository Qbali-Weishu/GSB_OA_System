CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(64) NOT NULL,
    email         VARCHAR(128) NOT NULL UNIQUE,
    phone         VARCHAR(20),
    dept_id       BIGINT NOT NULL,
    manager_id    BIGINT REFERENCES users(id),
    role          VARCHAR(32) NOT NULL DEFAULT 'employee',
    is_active     BOOLEAN NOT NULL DEFAULT true,
    join_date     DATE NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_dept_id ON users(dept_id);
CREATE INDEX idx_users_manager_id ON users(manager_id);

-- 初始管理员账号，密码：Admin@123（bcrypt hash）
INSERT INTO users (username, password_hash, name, email, dept_id, role, join_date)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    '系统管理员',
    'admin@company.com',
    1,
    'admin',
    '2020-01-01'
) ON CONFLICT DO NOTHING;
