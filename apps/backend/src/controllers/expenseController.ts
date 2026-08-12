import type { Request, Response } from 'express';
import type { APIResponse, Expense } from '../types/index.ts';
import { ExpenseCategory } from '../types/index.ts';

export const fakeExpenses: Expense[] = [
    {
        id: '1',
        userId: 'user123',
        amount: 45.99,
        category: ExpenseCategory.FOOD,
        description: '',
        date: new Date('2026-08-01'),
        createdAt: new Date(),
        updatedAt: new Date(),
    },
    {
        id: '2',
        userId: 'user123',
        amount: 31.99,
        category: ExpenseCategory.SHOPPING,
        description: '',
        date: new Date('2026-08-04'),
        createdAt: new Date(),
        updatedAt: new Date(),
    },
];

export const GetAllExpenses = (req: Request, res: Response) => {
    const response: APIResponse<Expense[]> = {
        success: true,
        data: fakeExpenses,
        message: 'Expenses retreived successfully',
    };

    res.status(200).json(response);
};

export const GetExpenseById = (req: Request, res: Response) => {
    const { id } = req.params;

    const expense = fakeExpenses.find((expense) => id === expense.id);

    if (!expense) {
        const response: APIResponse<null> = {
            success: false,
            error: 'Expense not found',
        };

        return res.status(404).json(response);
    }
    const response: APIResponse<Expense> = {
        success: true,
        data: expense,
    };
    res.status(200).json(response);
};

export const CreateExpense = (req: Request, res: Response) => {
    const { amount, category, description, date } = req.body;

    if (typeof amount !== 'number' || amount === 0 || category === '') {
        const response: APIResponse<null> = {
            success: false,
            error: 'invalid payload',
        };
        return res.status(400).json(response);
    }

    const newExpense: Expense = {
        id: crypto.randomUUID(),
        userId: 'user123',
        amount,
        category,
        description,
        date: date ? new Date(date) : new Date(),
        createdAt: new Date(),
        updatedAt: new Date(),
    };

    fakeExpenses.push(newExpense);

    const response: APIResponse<Expense> = {
        success: true,
        message: 'expense created',
        data: newExpense,
    };

    return res.status(200).json(response);
};

export const UpdateExpense = (req: Request, res: Response) => {
    const { id } = req.params;
    const { category, amount, description, data } = req.body;

    const expenseID = fakeExpenses.findIndex((expense) => expense.id === id);

    if (expenseID === -1) {
        const response: APIResponse<null> = {
            success: false,
            error: 'Expense not found',
        };
        return res.status(404).json(response);
    }

    fakeExpenses[expenseID] = {
        ...fakeExpenses[expenseID],
        amount: amount || fakeExpenses[expenseID].amount,
        category: category || fakeExpenses[expenseID].category,
        description: description || fakeExpenses[expenseID].description,
        date: data ? new Date() : fakeExpenses[expenseID].date,
        updatedAt: new Date(),
    };

    const response: APIResponse<Expense> = {
        success: true,
        message: 'Expense updated successfully',
        data: fakeExpenses[expenseID],
    };

    return res.status(200).json(response);
};

export const DeleteExpense = (req: Request, res: Response) => {
    const { id } = req.params;

    const expenseId = fakeExpenses.findIndex((expense) => expense.id == id);

    if (expenseId === -1) {
        const response: APIResponse<null> = {
            success: false,
            error: 'expense not found',
        };
        return res.status(404).json(response);
    }

    fakeExpenses.splice(expenseId, 1);
    const response: APIResponse<null> = {
        success: true,
        message: 'expense deleted successfully',
    };
    return res.status(404).json(response);
};
