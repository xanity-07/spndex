import * as z from 'zod';

export const ZUserResponse = z.object({
    id: z.uuid(),
    first_name: z.string(),
    last_name: z.string(),
    email: z.email(),
    created_at: z.iso.datetime(),
    updated_at: z.iso.datetime(),
});
