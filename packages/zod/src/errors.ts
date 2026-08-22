import * as z from 'zod';

export const ZActionType = z.enum(['redirect']);

export const ZAction = z.object({
    type: ZActionType,
    message: z.string(),
    value: z.string(),
});

export const ZFieldError = z.object({
    field: z.string(),
    error: z.string(),
});

export const ZHttpStatus = z.number().int().min(100).max(599);

export const ZAppError = z.object({
    action: ZAction.optional(),
    code: z.string().min(1),
    message: z.string().min(1),
    errors: z.array(ZFieldError).optional(),
    status: ZHttpStatus,
    override: z.boolean().optional(),
});

export type ActionType = z.infer<typeof ZActionType>;
export type Action = z.infer<typeof ZAction>;
export type FieldError = z.infer<typeof ZFieldError>;
export type HttpStatus = z.infer<typeof ZHttpStatus>;
export type AppError = z.infer<typeof ZAppError>;
