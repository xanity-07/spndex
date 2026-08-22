import {
    schemaWithPagination,
    ZAppError,
    ZCreateExpense,
    ZExpense,
    ZUpdateExpense,
} from '@spndex/zod';

import type { ZodOpenApiPathsObject } from 'zod-openapi';

import { getSecurityMetadata } from '../utils.js';

export const expensePaths: ZodOpenApiPathsObject = {
    '/expenses': {
        get: {
            tags: ['Expenses'],

            summary: 'List expenses',

            description:
                'Retrieve a paginated list of expenses belonging to the authenticated user.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '200': {
                    description: 'Expenses retrieved successfully.',

                    content: {
                        'application/json': {
                            schema: schemaWithPagination(ZExpense),
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

        post: {
            tags: ['Expenses'],

            summary: 'Create an expense',

            description: 'Create a new expense for the authenticated user.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '201': {
                    description: 'Expense created successfully.',

                    content: {
                        'application/json': {
                            schema: ZExpense,
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

            requestBody: {
                required: true,

                content: {
                    'application/json': {
                        schema: ZCreateExpense,
                    },
                },
            },
        },
    },

    '/expenses/{id}': {
        get: {
            tags: ['Expenses'],

            summary: 'Get an expense',

            description: 'Retrieve a single expense belonging to the authenticated user by its ID.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '200': {
                    description: 'Expense retrieved successfully.',

                    content: {
                        'application/json': {
                            schema: ZExpense,
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

        patch: {
            tags: ['Expenses'],

            summary: 'Update an expense',

            description: 'Update an existing expense belonging to the authenticated user.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '200': {
                    description: 'Expense updated successfully.',

                    content: {
                        'application/json': {
                            schema: ZExpense,
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

            requestBody: {
                required: true,

                content: {
                    'application/json': {
                        schema: ZUpdateExpense,
                    },
                },
            },
        },

        delete: {
            tags: ['Expenses'],

            summary: 'Delete an expense',

            description: 'Delete an existing expense belonging to the authenticated user.',

            ...getSecurityMetadata({
                securityType: 'bearer',
            }),

            responses: {
                '204': {
                    description: 'Expense deleted successfully.',
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
