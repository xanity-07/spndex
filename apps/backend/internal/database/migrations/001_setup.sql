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

CREATE TYPE currency_code AS ENUM (
    'USD',
    'EUR',
    'GBP',
    'CAD',
    'AUD'
);

---- create above / drop below ----

DROP TYPE IF EXISTS user_roles;
DROP TYPE IF EXISTS expense_category;
DROP TYPE IF EXISTS currency_code;
