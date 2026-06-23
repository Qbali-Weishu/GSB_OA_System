CREATE TABLE IF NOT EXISTS departments (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    parent_id   BIGINT REFERENCES departments(id),
    manager_id  BIGINT REFERENCES users(id),
    head_count  INT NOT NULL DEFAULT 0,
    min_on_site INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_departments_parent_id ON departments(parent_id);

-- 初始化测试部门数据
INSERT INTO departments (id, name, parent_id, head_count, min_on_site) VALUES
    (1, '总公司',    NULL, 0, 1),
    (2, '技术部',    1,    5, 2),
    (3, '产品部',    1,    4, 2),
    (4, '市场部',    1,    3, 2)
ON CONFLICT DO NOTHING;

-- 测试用员工（3人部门，min_on_site=2，验证并发请假场景）
INSERT INTO users (username, password_hash, name, email, dept_id, role, join_date) VALUES
    ('emp_a', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '员工甲', 'empa@company.com', 4, 'employee', '2022-03-01'),
    ('emp_b', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '员工乙', 'empb@company.com', 4, 'employee', '2022-03-01'),
    ('emp_c', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '员工丙', 'empc@company.com', 4, 'employee', '2022-03-01')
ON CONFLICT DO NOTHING;
