import type { Application } from 'express';
import express from 'express';
import { ExpenseCategory, Expenses } from './types/index.js';

const app: Application = express();
const PORT = 4001;

app.use(express.json());

let fakeExpenses: Expenses[] = [
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

app.get('/', (req, res) => {
    res.send('hello from typescript and express');
});

app.listen(PORT, () => {
    console.log(`Server is running on localhost:${PORT}`);
});
