import * as z from 'zod';

import { ZExpense, ZExpenseCategory } from './expense.js';

export const ZCategoryTotals = z.object({
    category: ZExpenseCategory,

    totalCents: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description:
                'Total expense amount for the category in cents. For example, 45000 represents 450.00 units of the applicable currency.',
            examples: [45000],
        }),

    count: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description: 'Number of expenses recorded in the category.',
            examples: [23],
        }),

    percentage: z
        .number()
        .min(0)
        .max(100)
        .meta({
            description: 'Percentage of total expenses represented by the category.',
            examples: [35.42],
        }),
});

export const ZMonthlyTotals = z.object({
    month: z
        .string()
        .regex(/^\d{4}-(0[1-9]|1[0-2])$/)
        .meta({
            description: 'Month represented in YYYY-MM format.',
            examples: ['2026-08'],
        }),

    totalCents: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description:
                'Total expense amount for the month in cents. For example, 125000 represents 1250.00 units of the applicable currency.',
            examples: [125000],
        }),

    count: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description: 'Number of expenses recorded during the month.',
            examples: [42],
        }),
});

export const ZDashboardStats = z.object({
    highestExpense: ZExpense.meta({
        description: 'The highest-value expense recorded by the user.',
    }),

    lowestExpense: ZExpense.meta({
        description: 'The lowest-value expense recorded by the user.',
    }),

    totalExpensesCents: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description:
                'Total amount of all expenses in cents. For example, 125000 represents 1250.00 units of the applicable currency.',
            examples: [125000],
        }),

    expenseCount: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description: 'Total number of expenses recorded by the user.',
            examples: [42],
        }),

    averageExpenseAmountCents: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description:
                'Average expense amount in cents. For example, 2976 represents 29.76 units of the applicable currency.',
            examples: [2976],
        }),

    currentMonthTotalCents: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description:
                'Total expense amount for the current month in cents. For example, 85000 represents 850.00 units of the applicable currency.',
            examples: [85000],
        }),

    lastMonthTotalCents: z
        .number()
        .int()
        .nonnegative()
        .meta({
            description:
                'Total expense amount for the previous month in cents. For example, 72000 represents 720.00 units of the applicable currency.',
            examples: [72000],
        }),

    monthlyNetChangeCents: z
        .number()
        .int()
        .meta({
            description:
                'Net change in expense spending between the current and previous month in cents. Positive values indicate an increase, while negative values indicate a decrease.',
            examples: [13000],
        }),
});
