import type { Application, Request, Response } from 'express';
import express from 'express';
import type { Expense } from './types/index.ts';
import { ExpenseCategory } from './types/index.ts';

const app: Application = express();
const PORT = 4001;

app.use(express.json());

let fakeExpenses: Expense[] = [
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

// Home route
app.get('/', (req, res) => {
    res.send('hello from typescript and express');
});

// GET All Expenses
app.get('/api/expenses/', (req: Request, res: Response) => {
    const response = {
        success: true,
        data: fakeExpenses,
        message: 'Expenses retreived successfully',
    };

    res.status(200).json(response);
});

// Get Expense by ID
app.get('/api/expenses/:id', (req: Request, res: Response) => {
    const { id } = req.params;

    const expense = fakeExpenses.find((expense) => id === expense.id);

    if (!expense) {
        const response = {
            success: false,
            error: 'Expense not found',
        };

        return res.status(404).json(response);
    }
    res.json(expense);
});

// Create Expense
app.post('/api/expenses', (req: Request, res: Response) => {
    const { amount, category, description, date } = req.body;

    if (typeof amount !== 'number' || amount === 0 || category === '') {
        const response = {
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

    const response = {
        success: true,
        message: 'expense created',
        data: newExpense,
    };

    return res.status(200).json(response);
});

app.patch('/api/expenses/:id', (req: Request, res: Response) => {
    const { id } = req.params;
    const { category, amount, description, data } = req.body;

    const expenseID = fakeExpenses.findIndex((expense) => expense.id === id);

    if (expenseID === -1) {
        const response = {
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

    const response = {
        success: true,
        message: 'Expense updated successfully',
        data: fakeExpenses[expenseID],
    };

    return res.status(200).json(response);
});

app.delete('/api/expenses/:id', (req: Request, res: Response) => {
    const { id } = req.params;

    const expenseId = fakeExpenses.findIndex((expense) => expense.id == id);

    if (expenseId === -1) {
        const response = {
            success: false,
            error: 'expense not found',
        };
        return res.status(404).json(response);
    }

    fakeExpenses.splice(expenseId, 1);
    const response = {
        success: true,
        message: 'expense deleted successfully',
    };
    return res.status(404).json(response);
});

app.listen(PORT, () => {
    console.log(`Server is running on localhost:${PORT}`);
});
