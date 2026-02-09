export class ProcessPaymentWebhookUseCase {
  constructor(
    paymentRepository,
    userRepository,
    abacatePayService,
    licenseService,
  ) {
    this.paymentRepository = paymentRepository;
    this.userRepository = userRepository;
    this.abacatePayService = abacatePayService;
    this.licenseService = licenseService;
  }

  async execute(input) {
    const { event, data } = input;

    // Encontrar pagamento pelo abacatePayPixId
    const payment = await this.paymentRepository.findByAbacatePayPixId(data.id);

    if (!payment) {
      console.warn(`[ProcessPaymentWebhook] Pagamento não encontrado: ${data.id}`);
      return { message: 'Pagamento não encontrado' };
    }

    // Processar evento de pagamento confirmado
    if (event === 'payment.confirmed') {
      // Atualizar status
      await this.paymentRepository.update(payment.id, {
        status: 'PAID',
        updatedAt: new Date(),
      });

      console.log(`✅ Pagamento confirmado: ${payment.id}`);

      // Encontrar usuário
      const user = await this.userRepository.findById(payment.userId);
      if (!user) {
        console.error(`❌ Usuário não encontrado: ${payment.userId}`);
        return { error: 'Usuário não encontrado' };
      }

      // Gerar token de licença
      const licenseToken = this.licenseService.generateLicense({
        userId: user.id,
        email: user.email,
        days: 365,
      });

      // Ativar licença do usuário
      const licenseExpiresAt = new Date();
      licenseExpiresAt.setFullYear(licenseExpiresAt.getFullYear() + 1);
      
      await this.userRepository.update(user.id, {
        licenseActive: true,
        licenseExpiresAt,
        licenseToken,
      });

      return {
        event: 'payment.confirmed',
        paymentId: payment.id,
        userId: user.id,
        status: 'PAID',
        message: 'Pagamento confirmado e licença ativada',
      };
    }

    // Processar evento de pagamento cancelado
    if (event === 'payment.cancelled') {
      await this.paymentRepository.update(payment.id, {
        status: 'CANCELLED',
        updatedAt: new Date(),
      });

      console.log(`❌ Pagamento cancelado: ${payment.id}`);

      return {
        event: 'payment.cancelled',
        paymentId: payment.id,
        status: 'CANCELLED',
      };
    }

    // Processar evento de pagamento expirado
    if (event === 'payment.expired') {
      await this.paymentRepository.update(payment.id, {
        status: 'EXPIRED',
        updatedAt: new Date(),
      });

      console.log(`⏰ Pagamento expirado: ${payment.id}`);

      return {
        event: 'payment.expired',
        paymentId: payment.id,
        status: 'EXPIRED',
      };
    }

    return {
      message: 'Evento processado',
      event,
    };
  }
}
