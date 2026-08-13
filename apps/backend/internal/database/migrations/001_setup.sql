-- Write your migrate up statements here

CREATE TYPE expense_category AS ENUM (
    'food',
    'transport',
    'education',
    'utilities',
    'entertainment',
    'healthcare',
    'shopping'
    'other'
)
