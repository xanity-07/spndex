import * as z from 'zod';

export const ZUserResponse = z.object({
    id: z.uuid(),
    role: z.enum(['user', 'admin']),
    firstName: z.string(),
    lastName: z.string(),
    email: z.email(),
    createdAt: z.iso.datetime(),
    updatedAt: z.iso.datetime(),
    deletedAt: z.string(),
});

export const ZUpdateUserRequest = z
    .object({
        firstName: z.string().optional(),
        lastName: z.string().optional(),
        email: z.email().optional(),
        password: z.string().min(8).optional(),
    })
    .meta({
        example: {
            firstName: 'John',
            lastName: 'Doe',
            email: 'john@example.com',
            password: 'Password123!',
        },
    });
