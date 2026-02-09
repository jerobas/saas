export class RenewSubscriptionUseCase {
  constructor(userRepository, paymentRepository, abacatePayService) {
    this.userRepository = userRepository;
    this.paymentRepository = paymentRepository;
    this.abacatePayService = abacatePayService;
  }

  async execute(input) {
    const { userId, email } = input;

    let user;
    if (userId) {
      user = await this.userRepository.findById(userId);
    } else if (email) {
      user = await this.userRepository.findByEmail(email);
    }

    if (!user) {
      throw new Error('Usuário não encontrado');
    }

    if (!user.abacatePayCustomerId) {
      throw new Error('Usuário não possui cliente AbacatePay registrado');
    }

    const amount = parseInt(process.env.LICENSE_PRICE || '50000');
    const pixResponse = await this.abacatePayService.createPixQrCode({
      customerId: user.abacatePayCustomerId,
      amount,
      description: 'Renovação de Licença',
      expiresInMinutes: 30,
    });

    if (pixResponse.error && pixResponse.error !== '<unknown>') {
      throw new Error(`Erro ao criar PIX: ${pixResponse.error}`);
    }

    const pixData = pixResponse.data;
    if (!pixData?.id) {
      throw new Error('Falha ao obter dados do PIX');
    }

    const payment = await this.paymentRepository.create({
      userId: user.id,
      abacatePayCustomerId: user.abacatePayCustomerId,
      abacatePayPixId: pixData.id,
      amount,
      status: 'PENDING',
      brCode: pixData.brCode,
      brCodeBase64: pixData.brCodeBase64,
      expiresAt: new Date(pixData.expiresAt),
    });

    return {
      userId: user.id,
      paymentId: payment.id,
      brCode: pixData.brCode,
      brCodeBase64: pixData.brCodeBase64,
      expiresAt: pixData.expiresAt,
      amount: amount / 100,
    };
  }
}
