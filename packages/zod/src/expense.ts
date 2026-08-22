import * as z from 'zod';

export const ZExpenseCategory = z.enum([
    'food',
    'transport',
    'utilities',
    'entertainment',
    'healthcare',
    'shopping',
    'education',
    'other',
]);

export const ZCurrencyCode = z.enum(['AUD', 'USD', 'EUR', 'GBP', 'CAD']);

export const ZExpense = z
    .object({
        id: z.uuid(),

        userId: z.uuid(),

        description: z.string().min(1).max(255).optional(),

        amountCents: z
            .number()
            .int()
            .positive()
            .meta({
                description:
                    'Expense amount in cents. For example, 3499 represents 34.99 units of the applicable currency.',
                examples: [3499],
            }),

        date: z.string(),

        category: ZExpenseCategory,

        currencyCode: ZCurrencyCode.optional(),
    })
    .meta({
        description: 'An expense belonging to the authenticated user.',
    });

export const ZCreateExpense = z
    .object({
        description: z.string().min(1).max(255).optional(),

        amountCents: z
            .number()
            .int()
            .positive()
            .meta({
                description:
                    'Expense amount in cents. For example, 2999 represents 29.99 units of the applicable currency.',
                examples: [2999],
            }),

        date: z.string().meta({
            description: 'Date the expense occurred in YYYY-MM-DD format.',
            examples: ['2026-08-29'],
        }),

        category: ZExpenseCategory,

        currencyCode: ZCurrencyCode.optional(),
    })
    .meta({
        description: 'Create a new expense.',
    });

export const ZUpdateExpense = z
    .object({
        description: z.string().min(1).max(255).optional(),

        amountCents: z
            .number()
            .int()
            .positive()
            .optional()
            .meta({
                description:
                    'Updated expense amount in cents. For example, 3499 represents 34.99 units of the applicable currency.',
                examples: [3499],
            }),

        date: z
            .string()
            .optional()
            .meta({
                description: 'Updated date of the expense in YYYY-MM-DD format.',
                examples: ['2026-08-29'],
            }),

        category: ZExpenseCategory.optional(),

        currencyCode: ZCurrencyCode.optional(),
    })
    .meta({
        description: 'Update an existing expense.',
    });
