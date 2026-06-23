CREATE TABLE IF NOT EXISTS leaves (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL REFERENCES users(id),
    dept_id             BIGINT NOT NULL REFERENCES departments(id),
    leave_type          VARCHAR(32) NOT NULL,
    start_date          DATE NOT NULL,
    end_date            DATE NOT NULL,
    working_days        INT NOT NULL DEFAULT 0,
    reason              TEXT NOT NULL,
    status              VARCHAR(32) NOT NULL DEFAULT 'pending',
    required_approvals  INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_leave_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_leaves_user_id     ON leaves(user_id);
CREATE INDEX idx_leaves_dept_id     ON leaves(dept_id);
CREATE INDEX idx_leaves_status      ON leaves(status);
CREATE INDEX idx_leaves_start_date  ON leaves(start_date);

-- 假期余额表
CREATE TABLE IF NOT EXISTS leave_balances (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    year        INT NOT NULL,
    leave_type  VARCHAR(32) NOT NULL,
    total_days  NUMERIC(5,1) NOT NULL DEFAULT 0,
    used_days   NUMERIC(5,1) NOT NULL DEFAULT 0,
    pending_days NUMERIC(5,1) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, year, leave_type)
);

CREATE INDEX idx_leave_balances_user_year ON leave_balances(user_id, year);

-- 为测试用员工初始化年假余额（每人10天年假）
INSERT INTO leave_balances (user_id, year, leave_type, total_days, used_days)
SELECT u.id, EXTRACT(YEAR FROM NOW())::INT, 'annual', 10, 0
FROM users u
WHERE u.username IN ('emp_a', 'emp_b', 'emp_c')
ON CONFLICT DO NOTHING;
