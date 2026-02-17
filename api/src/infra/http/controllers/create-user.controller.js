import { CreateUserUseCase } from '../../../domain/users/usecases/create-user-usecase.js';

export class CreateUserController {
  constructor(userRepository, paymentRepository, abacatePayService, passwordService) {
    this.createUserUseCase = new CreateUserUseCase(
      userRepository,
      paymentRepository,
      abacatePayService,
      passwordService
    );
  }

  async handle(req, res) {
    try {
      const { email, name, taxId, cellphone, password } = req.body;

      // Validar campos obrigatórios
      if (!email || !name || !taxId || !cellphone || !password) {
        return res.status(400).json({
          error: 'Email, name, taxId, cellphone e password são obrigatórios',
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
        password,
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
