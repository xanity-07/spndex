import type { ZodOpenApiOperationObject } from 'zod-openapi';

type SecurityType = 'bearer' | 'service';

const securitySchemes: Record<SecurityType, Record<string, string[]>[]> = {
    bearer: [{ bearerAuth: [] }],
    service: [{ 'x-service-token': [] }],
};

export const getSecurityMetadata = ({
    security = true,
    securityType = 'bearer',
}: {
    security?: boolean;
    securityType?: SecurityType;
} = {}): Partial<ZodOpenApiOperationObject> => {
    return {
        ...(security && { security: securitySchemes[securityType] }),
    };
};
