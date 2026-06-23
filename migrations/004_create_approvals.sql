CREATE TABLE IF NOT EXISTS approvals (
    id            BIGSERIAL PRIMARY KEY,
    leave_id      BIGINT NOT NULL REFERENCES leaves(id),
    approver_id   BIGINT NOT NULL,
    approver_type VARCHAR(32) NOT NULL,
    step_order    INT NOT NULL DEFAULT 1,
    status        VARCHAR(32) NOT NULL DEFAULT 'pending',
    comment       TEXT NOT NULL DEFAULT '',
    acted_at      TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_approvals_leave_id     ON approvals(leave_id);
CREATE INDEX idx_approvals_approver_id  ON approvals(approver_id);
CREATE INDEX idx_approvals_status       ON approvals(status);

CREATE TABLE IF NOT EXISTS notifications (
    id          BIGSERIAL PRIMARY KEY,
    receiver_id BIGINT NOT NULL REFERENCES users(id),
    sender_id   BIGINT REFERENCES users(id),
    ref_id      BIGINT NOT NULL,
    ref_type    VARCHAR(32) NOT NULL,
    type        VARCHAR(64) NOT NULL,
    title       VARCHAR(255) NOT NULL,
    content     TEXT NOT NULL,
    status      VARCHAR(32) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    sent_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_receiver_id ON notifications(receiver_id);
CREATE INDEX idx_notifications_status      ON notifications(status);
