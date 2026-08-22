import {
    ZAppError,
    ZLoginRequest,
    ZLoginResponse,
    ZRegisterRequest,
    ZUserResponse,
} from '@spndex/zod';
import type { ZodOpenApiPathsObject } from 'zod-openapi';
import { getSecurityMetadata } from '../utils.js';

export const authPaths: ZodOpenApiPathsObject = {
    'auth/register': {
        post: {
            tags: ['Auth'],
            summary: 'Register user',
            description: 'Register a platform user',

            responses: {
                '201': {
                    description: 'Register user',
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
            requestBody: {
                required: true,
                content: {
                    'application/json': {
                        schema: ZRegisterRequest,
                    },
                },
            },
        },
    },
    'auth/login': {
        post: {
            tags: ['Auth'],
            summary: 'Login',
            description: 'Login to user account',

            responses: {
                '200': {
                    description: 'Successfully authenticated',
                    content: {
                        'application/json': {
                            schema: ZLoginResponse,
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
                    description: 'Invalid credentials',
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
                        schema: ZLoginRequest,
                    },
                },
            },
        },
    },
    'auth/logout': {
        post: {
            tags: ['Auth'],
            summary: 'Logout',
            description: 'Logout of user account',
            ...getSecurityMetadata({ securityType: 'bearer' }),

            responses: {
                '204': {
                    description: 'Successfully logged out',
                },

                '401': {
                    description: 'Failed to logout',
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
