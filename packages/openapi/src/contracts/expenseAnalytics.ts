import { ZAppError, ZCategoryTotals, ZDashboardStats, ZMonthlyTotals } from '@spndex/zod';

import type { ZodOpenApiPathsObject } from 'zod-openapi';

import { getSecurityMetadata } from '../utils.js';

export const expenseAnalyticsPaths: ZodOpenApiPathsObject = {
    '/expenses/analytics/category-totals': {
        get: {
            tags: ['Expense Analytics'],

            summary: 'Get expense totals by category',

            description: 'Retrieve expense totals grouped by category for the authenticated user.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '200': {
                    description: 'Expense totals grouped by category.',

                    content: {
                        'application/json': {
                            schema: ZCategoryTotals.array(),
                        },
                    },
                },

                '400': {
                    description: 'Invalid request.',

                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },

                '401': {
                    description: 'Authentication required.',

                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
            },
        },
    },

    '/expenses/analytics/monthly-expenses': {
        get: {
            tags: ['Expense Analytics'],

            summary: 'Get monthly expense totals',

            description: 'Retrieve expense totals grouped by month for the authenticated user.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '200': {
                    description: 'Monthly expense totals.',

                    content: {
                        'application/json': {
                            schema: ZMonthlyTotals.array(),
                        },
                    },
                },

                '400': {
                    description: 'Invalid request.',

                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },

                '401': {
                    description: 'Authentication required.',

                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
            },
        },
    },

    '/expenses/analytics/dashboard': {
        get: {
            tags: ['Expense Analytics'],

            summary: 'Get expense dashboard',

            description: 'Retrieve summary expense statistics for the authenticated user.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '200': {
                    description: 'Expense dashboard statistics.',

                    content: {
                        'application/json': {
                            schema: ZDashboardStats,
                        },
                    },
                },

                '400': {
                    description: 'Invalid request.',

                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },

                '401': {
                    description: 'Authentication required.',

                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
            },
        },
    },
};
