-- up migrations
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    description TEXT,
    currency currency_code NOT NULL,
    category expense_category NOT NULL,
    date TEXT NOT NULL,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expenses_user_id
ON expenses(user_id);

CREATE INDEX idx_expenses_created_at
ON expenses(user_id, created_at DESC);

---- create above / drop below ----
DROP TABLE IF EXISTS expenses;
