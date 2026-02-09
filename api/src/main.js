import 'reflect-metadata';
import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import http from 'http';
import { AppDataSource } from './infra/database/data-source.js';
import { setupRoutes } from './infra/http/routes/index.js';
import { setupSwagger } from './infra/http/swagger/setup.js';
import { errorHandler } from './infra/http/middlewares/error-handler.js';

const app = express();
const server = http.createServer(app);
const PORT = process.env.PORT || 3000;

// Middlewares
app.use(helmet());
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Database connection
const initializeDatabase = async () => {
  try {
    await AppDataSource.initialize();
    console.log('✅ Database connected successfully');
  } catch (error) {
    console.error('❌ Database connection failed:', error);
    process.exit(1);
  }
};

// Setup routes and swagger
setupSwagger(app);
setupRoutes(app);

// Error handler (must be last)
app.use(errorHandler);

// Start server
const startServer = async () => {
  await initializeDatabase();
  
  server.listen(PORT, () => {
    console.log(`🚀 Server running on port ${PORT}`);
    console.log(`📚 Swagger docs: http://localhost:${PORT}/api-docs`);
  });
};

startServer();
