import type { Application } from 'express';
import express from 'express';
import expenseRoutes from './routes/expenseRoutes.ts';

const app: Application = express();
const PORT = 4001;

app.use(express.json());

// Home route
app.get('/', (req, res) => {
    res.send('hello from typescript and express');
});

// Expense routes
app.use('/api/expenses/', expenseRoutes);

app.listen(PORT, () => {
    console.log(`Server is running on localhost:${PORT}`);
});
