-- up migrations

CREATE TYPE expense_category AS ENUM (
    'food',
    'transport',
    'education',
    'utilities',
    'entertainment',
    'healthcare',
    'shopping',
    'other'
);

CREATE TYPE user_roles AS ENUM (
    'user',
    'admin'
);

---- create above / drop below ----

DROP TYPE IF EXISTS user_roles;
DROP TYPE IF EXISTS expense_category;
