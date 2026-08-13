import { schemaWithPagination, ZAppError, ZUserResponse } from '@spndex/zod';
import type { ZodOpenApiPathsObject } from 'zod-openapi';
import { getSecurityMetadata } from '../utils.js';

export const userPaths: ZodOpenApiPathsObject = {
    '/users': {
        get: {
            tags: ['Users'],
            summary: 'Get all users',
            description: 'Get a list of all users',
            ...getSecurityMetadata({ securityType: 'bearer' }),
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

        post: {
            tags: ['Users'],
            summary: 'Create user',
            description: 'Create a platform user',
            ...getSecurityMetadata({ securityType: 'bearer' }),
            responses: {
                '201': {
                    description: 'Create user',
                    content: {
                        'application/json': {
                            schema: ZUserResponse,
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
                '409': {
                    description: 'User already exists',
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
            ...getSecurityMetadata({ securityType: 'bearer' }),
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
    },
};
