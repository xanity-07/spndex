import type { Application } from 'express';
import express from 'express';

const app: Application = express();
const PORT = 4001;

app.use(express.json());

app.get('/', (req, res) => {
  res.send('hello from typescript and express');
});

app.listen(PORT, () => {
  console.log(`Server is running on localhost:${PORT}`);
});
