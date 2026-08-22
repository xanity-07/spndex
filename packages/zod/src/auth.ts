import * as z from 'zod';

export const ZLoginResponse = z.object({
    token: z.string(),
});

export const ZLoginRequest = z
    .object({
        email: z.email(),
        password: z.string(),
    })
    .meta({
        example: {
            email: 'john@example.com',
            password: 'Password@123',
        },
    });

export const ZRegisterRequest = z
    .object({
        firstName: z.string(),
        lastName: z.string(),
        email: z.email(),
        password: z.string().min(8),
    })
    .meta({
        example: {
            firstName: 'John',
            lastName: 'Doe',
            email: 'john@example.com',
            password: 'Password123!',
        },
    });

export const ZLogoutResponse = z.object({
    message: z.string(),
});
