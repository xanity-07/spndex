export const ExpenseCategory = {
    FOOD: 'food',
    TRANSPORT: 'transport',
    UTILITIES: 'utilities',
    ENTERTAINMENT: 'entertainment',
    HEALTHCARE: 'healthcare',
    SHOPPING: 'shopping',
    EDUCATION: 'education',
    OTHER: 'other',
} as const;
export type ExpenseCategory = (typeof ExpenseCategory)[keyof typeof ExpenseCategory];

export interface Expenses {
    id: string;
    userId: string;
    amount: number;
    category: ExpenseCategory;
    description: string;
    date: Date;
    createdAt: Date;
    updatedAt: Date;
}

export interface User {
    id: string;
    name: string;
    email: string;
    password: string;
    createdAt: Date;
    updatedAt: Date;
}
