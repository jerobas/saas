import { RenewSubscriptionUseCase } from '../../../domain/users/usecases/renew-subscription-usecase.js';

export class RenewSubscriptionController {
  constructor(userRepository, paymentRepository, abacatePayService) {
    this.renewSubscriptionUseCase = new RenewSubscriptionUseCase(
      userRepository,
      paymentRepository,
      abacatePayService
    );
  }

  async handle(req, res) {
    try {
      const { userId, email } = req.body;

      // Validar que pelo menos um identificador foi fornecido
      if (!userId && !email) {
        return res.status(400).json({
          error: 'userId ou email é obrigatório',
        });
      }

      const result = await this.renewSubscriptionUseCase.execute({
        userId,
        email,
      });

      return res.status(200).json({
        success: true,
        data: result,
      });
    } catch (error) {
      console.error('[RenewSubscriptionController]', error);
      return res.status(400).json({
        error: error.message,
      });
    }
  }
}
