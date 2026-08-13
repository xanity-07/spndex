import { ZAppError } from '@spndex/zod';
import { createDocument, type ZodOpenApiObject } from 'zod-openapi';
import { healthPaths } from './contracts/health.js';
import { userPaths } from './contracts/users.js';

const openApiConfig: ZodOpenApiObject = {
    openapi: '3.1.0',
    info: {
        version: '1.0.0',
        title: 'Spndex REST API - Documentation',
        description: 'Spndex REST API - Documentation',
    },
    servers: [
        {
            url: 'http://localhost:8080/',
            description: 'Local Server',
        },
        {
            url: 'http://localhost:8080/api/v1',
            description: 'Local Server',
        },
    ],
    paths: {
        ...healthPaths,
        ...userPaths,
    },
    components: {
        schemas: {
            AppError: ZAppError,
        },
        securitySchemes: {
            bearerAuth: {
                type: 'http',
                scheme: 'bearer',
                bearerFormat: 'JWT',
            },
            'x-service-token': {
                type: 'apiKey',
                name: 'x-service-token',
                in: 'header',
            },
        },
    },
};

export const OpenAPI: ReturnType<typeof createDocument> = createDocument(openApiConfig);
