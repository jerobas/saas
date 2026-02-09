import { CreateUserUseCase } from '../../../domain/users/usecases/create-user-usecase.js';

export class CreateUserController {
  constructor(userRepository, paymentRepository, abacatePayService) {
    this.createUserUseCase = new CreateUserUseCase(
      userRepository,
      paymentRepository,
      abacatePayService
    );
  }

  async handle(req, res) {
    try {
      const { email, name, taxId, cellphone } = req.body;

      // Validar campos obrigatórios
      if (!email || !name || !taxId || !cellphone) {
        return res.status(400).json({
          error: 'Email, name, taxId e cellphone são obrigatórios',
        });
      }

      // Validar formato de email
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        return res.status(400).json({
          error: 'Email inválido',
        });
      }

      const result = await this.createUserUseCase.execute({
        email,
        name,
        taxId,
        cellphone,
      });

      return res.status(201).json({
        success: true,
        data: result,
      });
    } catch (error) {
      console.error('[CreateUserController]', error);
      return res.status(400).json({
        error: error.message,
      });
    }
  }
}
