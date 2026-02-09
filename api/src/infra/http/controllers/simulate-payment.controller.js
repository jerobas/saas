import { SimulatePaymentUseCase } from '../../../domain/payments/usecases/simulate-payment-usecase.js';

export class SimulatePaymentController {
  constructor(
    paymentRepository,
    userRepository,
    abacatePayService,
    licenseService,
  ) {
    this.simulatePaymentUseCase = new SimulatePaymentUseCase(
      paymentRepository,
      userRepository,
      abacatePayService,
      licenseService,
    );
  }

  async handle(req, res) {
    try {
      // Verificar ambiente
      if (process.env.NODE_ENV !== 'development') {
        return res.status(403).json({
          error: 'Este endpoint está disponível apenas em desenvolvimento',
        });
      }

      const { paymentId } = req.body;

      if (!paymentId) {
        return res.status(400).json({
          error: 'paymentId é obrigatório',
        });
      }

      const result = await this.simulatePaymentUseCase.execute({ paymentId });

      return res.status(200).json({
        success: true,
        data: result,
      });
    } catch (error) {
      console.error('[SimulatePaymentController]', error);
      return res.status(400).json({
        error: error.message,
      });
    }
  }
}
