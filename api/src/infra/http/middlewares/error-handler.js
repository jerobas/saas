import { AppError } from '../errors/app-error.js';

export const errorHandler = (err, req, res, next) => {
  console.error(err);

  if (err instanceof AppError) {
    return res.status(err.statusCode).json({
      error: err.message,
      statusCode: err.statusCode,
    });
  }

  return res.status(500).json({
    error: 'Internal server error',
    statusCode: 500,
  });
};
