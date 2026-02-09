import { ProcessPaymentWebhookUseCase } from '../../../domain/payments/usecases/process-payment-webhook-usecase.js';

export class WebhookPaymentController {
  constructor(
    paymentRepository,
    userRepository,
    abacatePayService,
    licenseService,
  ) {
    this.processPaymentWebhookUseCase = new ProcessPaymentWebhookUseCase(
      paymentRepository,
      userRepository,
      abacatePayService,
      licenseService,
    );
  }

  async handle(req, res) {
    try {
      const signature = req.headers['x-abacatepay-signature'];
      const body = req.body;

      // Validar assinatura do webhook
      const isValid = this.abacatePayService.validateWebhookSignature(
        body,
        signature
      );

      if (!isValid) {
        return res.status(401).json({
          error: 'Assinatura de webhook inválida',
        });
      }

      const result = await this.processPaymentWebhookUseCase.execute({
        event: body.event,
        data: body.data,
      });

      return res.status(200).json({
        success: true,
        data: result,
      });
    } catch (error) {
      console.error('[WebhookPaymentController]', error);
      return res.status(500).json({
        error: 'Erro ao processar webhook',
      });
    }
  }
}
