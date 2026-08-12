import { Router } from 'express';
import {
    CreateExpense,
    DeleteExpense,
    GetAllExpenses,
    GetExpenseById,
    UpdateExpense,
} from '../controllers/expenseController.ts';

const router: Router = Router();

router.get('/', GetAllExpenses);
router.get('/:id', GetExpenseById);
router.post('/', CreateExpense);
router.patch('/:id', UpdateExpense);
router.delete('/:id', DeleteExpense);

export default router;
