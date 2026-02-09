export class CreateUserUseCase {
  constructor(userRepository, paymentRepository, abacatePayService) {
    this.userRepository = userRepository;
    this.paymentRepository = paymentRepository;
    this.abacatePayService = abacatePayService;
  }

  async execute(input) {
    const { email, name, taxId, cellphone } = input;

    const existingUser = await this.userRepository.findByEmail(email);
    if (existingUser) {
      throw new Error(`Email ${email} já cadastrado`);
    }

    const user = await this.userRepository.create({
      email,
      name,
      taxId,
      cellphone,
    });

    const customerResponse = await this.abacatePayService.createCustomer({
      name,
      email,
      taxId,
      cellphone,
    });

    if (customerResponse.error && customerResponse.error !== '<unknown>') {
      throw new Error(`Erro ao criar cliente: ${customerResponse.error}`);
    }

    const abacatePayCustomerId = customerResponse.data?.id;
    if (!abacatePayCustomerId) {
      throw new Error('Falha ao obter ID do cliente AbacatePay');
    }

    user.abacatePayCustomerId = abacatePayCustomerId;
    await this.userRepository.update(user.id, {
      abacatePayCustomerId,
    });

    const amount = parseInt(process.env.LICENSE_PRICE || '50000');
    const pixResponse = await this.abacatePayService.createPixQrCode({
      customerId: abacatePayCustomerId,
      amount,
      description: 'Licença de Software',
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
      abacatePayCustomerId,
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
