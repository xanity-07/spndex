import * as z from 'zod';

export type PaginatesResponse<T> = {
    data: T[];
    total: number;
    page: number;
    limit: number;
    totalPages: number;
};

export const schemaWithPagination = <T extends z.ZodType>(schema: T) =>
    z.object({
        data: z.array(schema),
        total: z.number(),
        page: z.number(),
        limit: z.number(),
        totalPages: z.number(),
    });
