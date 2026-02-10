export class SimulatePaymentUseCase {
  constructor(
    paymentRepository,
    userRepository,
    abacatePayService,
    licenseService
  ) {
    this.paymentRepository = paymentRepository;
    this.userRepository = userRepository;
    this.abacatePayService = abacatePayService;
    this.licenseService = licenseService;
  }

  async execute(input) {
    const { paymentId } = input;

    // Encontrar pagamento
    const payment = await this.paymentRepository.findById(paymentId);
    if (!payment) {
      throw new Error('Pagamento não encontrado');
    }

    if (payment.status === 'PAID') {
      throw new Error('Este pagamento já foi confirmado');
    }

    // Simular pagamento no AbacatePay
    try {
      await this.abacatePayService.simulatePixPayment(payment.abacatePayPixId);
    } catch (error) {
      throw new Error(`Erro ao simular pagamento: ${error.message}`);
    }

    // Atualizar status do pagamento
    await this.paymentRepository.update(paymentId, {
      status: 'PAID',
      updatedAt: new Date(),
    });

    // Atualizar pagamento com dados atualizados
    const updatedPayment = await this.paymentRepository.findById(paymentId);

    // Encontrar usuário
    const user = await this.userRepository.findById(updatedPayment.userId);
    if (!user) {
      throw new Error('Usuário associado ao pagamento não encontrado');
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
      licenseToken
    });

    console.log(`🔓 Licença ativada para usuário ${user.id}, expira em ${licenseExpiresAt.toISOString()}`);

    return {
      success: true,
      paymentId,
      status: 'PAID',
      message: 'Pagamento simulado com sucesso',
      tokenSent: true,
    };
  }
}
