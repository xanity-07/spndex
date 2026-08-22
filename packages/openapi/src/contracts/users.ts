import { schemaWithPagination, ZAppError, ZUpdateUserRequest, ZUserResponse } from '@spndex/zod';
import type { ZodOpenApiPathsObject } from 'zod-openapi';
import { getSecurityMetadata } from '../utils.js';

export const userPaths: ZodOpenApiPathsObject = {
    '/users': {
        get: {
            tags: ['Users'],
            summary: 'Get all users',
            description: 'Get a list of all users',
            responses: {
                '200': {
                    description: 'List of users',
                    content: {
                        'application/json': {
                            schema: schemaWithPagination(ZUserResponse),
                        },
                    },
                },
                '400': {
                    description: 'Invalid request',
                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
                '401': {
                    description: 'Unauthorized',
                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
            },
        },
    },

    '/users/{id}': {
        get: {
            tags: ['Users'],
            summary: 'Get user by ID',
            description: "Get a single user's details",
            responses: {
                '200': {
                    description: 'User details',
                    content: {
                        'application/json': {
                            schema: ZUserResponse,
                        },
                    },
                },
                '400': {
                    description: 'Invalid user ID',
                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
                '404': {
                    description: 'User not found',
                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
            },
        },
        patch: {
            tags: ['Users'],
            summary: 'Update User',
            description: 'Update User',
            ...getSecurityMetadata({ securityType: 'bearer' }),
            responses: {
                '200': {
                    description: 'User updated',
                    content: {
                        'application/json': {
                            schema: ZUserResponse,
                        },
                    },
                },
                '404': {
                    description: 'User not found',
                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
                '401': {
                    description: 'Unauthorized',
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
                        schema: ZUpdateUserRequest,
                    },
                },
            },
        },
        delete: {
            tags: ['Users'],
            summary: 'Delete User',
            description: 'Delete User',
            ...getSecurityMetadata({ securityType: 'bearer' }),
            responses: {
                '204': {
                    description: 'User deleted',
                    content: {
                        'application/json': {
                            schema: ZUserResponse,
                        },
                    },
                },
                '404': {
                    description: 'User not found',
                    content: {
                        'application/json': {
                            schema: ZAppError,
                        },
                    },
                },
                '401': {
                    description: 'Unauthorized',
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
