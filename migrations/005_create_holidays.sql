CREATE TABLE IF NOT EXISTS holidays (
    id          BIGSERIAL PRIMARY KEY,
    date        DATE NOT NULL UNIQUE,
    name        VARCHAR(64) NOT NULL,
    is_workday  BOOLEAN NOT NULL DEFAULT false,
    description VARCHAR(255) NOT NULL DEFAULT '',
    year        INT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_holidays_year ON holidays(year);

-- 初始化当前年度的法定节假日示例数据
INSERT INTO holidays (date, name, is_workday, year) VALUES
    (DATE_TRUNC('year', NOW())::DATE + INTERVAL '0 days',  '元旦',     false, EXTRACT(YEAR FROM NOW())::INT),
    (DATE_TRUNC('year', NOW())::DATE + INTERVAL '120 days','劳动节假期', false, EXTRACT(YEAR FROM NOW())::INT),
    (DATE_TRUNC('year', NOW())::DATE + INTERVAL '121 days','劳动节假期', false, EXTRACT(YEAR FROM NOW())::INT),
    (DATE_TRUNC('year', NOW())::DATE + INTERVAL '122 days','劳动节假期', false, EXTRACT(YEAR FROM NOW())::INT)
ON CONFLICT (date) DO NOTHING;
