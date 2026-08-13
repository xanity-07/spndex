import { ZHealthResponse } from '@spndex/zod';
import type { ZodOpenApiPathsObject } from 'zod-openapi';
import { getSecurityMetadata } from '../utils.js';

export const healthPaths: ZodOpenApiPathsObject = {
    '/status': {
        get: {
            tags: ['Health'],
            summary: 'Get health',
            description: 'Get health status',
            ...getSecurityMetadata({ security: false }),
            responses: {
                '200': {
                    description: 'Health status',
                    content: {
                        'application/json': {
                            schema: ZHealthResponse,
                        },
                    },
                },
            },
        },
    },
};
