import type { ZodOpenApiOperationObject } from 'zod-openapi';

type SecurityType = 'bearer' | 'service';

const securitySchemes = {
    bearer: [{ bearerAuth: [] }],
    service: [{ 'x-service-token': [] }],
} satisfies Record<SecurityType, Record<string, string[]>[]>;

export const getSecurityMetadata = ({
    security = true,
    securityType = 'bearer',
}: {
    security?: boolean;
    securityType?: SecurityType;
} = {}): Partial<ZodOpenApiOperationObject> => ({
    ...(security && {
        security: securitySchemes[securityType],
    }),
});
