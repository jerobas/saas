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

    let customerResponse;
    try {
      customerResponse = await this.abacatePayService.createCustomer({
        name,
        email,
        taxId,
        cellphone,
      });

      if (customerResponse.error && customerResponse.error !== '<unknown>') {
        throw new Error(`Erro ao criar cliente: ${customerResponse.error}`);
      }
    } catch (error) {
      await this.userRepository.delete(user.id);
      throw new Error(`Erro ao criar cliente AbacatePay: ${error.message}`);
    }

    const abacatePayCustomerId = customerResponse.id;
    if (!abacatePayCustomerId) {
      await this.userRepository.delete(user.id);
      throw new Error('Falha ao obter ID do cliente AbacatePay');
    }

    user.abacatePayCustomerId = abacatePayCustomerId;
    await this.userRepository.update(user.id, {
      abacatePayCustomerId,
    });

    const amount = parseInt(process.env.LICENSE_PRICE || '50000');
    const pixResponse = await this.abacatePayService.createPixQrCode({
      amount,
      description: 'Licença de Software',
      expiresInMinutes: 30,
      customer: {
        name,
        email,
        taxId,
        cellphone,
      },
    });

    if (pixResponse.error && pixResponse.error !== '<unknown>') {
      throw new Error(`Erro ao criar PIX: ${pixResponse.error}`);
    }

    if (!pixResponse?.id) {
      throw new Error('Falha ao obter dados do PIX');
    }

    const payment = await this.paymentRepository.create({
      userId: user.id,
      abacatePayCustomerId,
      abacatePayPixId: pixResponse.id,
      amount,
      status: 'PENDING',
      pixCode: pixResponse.pixCode,
      pixQrCode: pixResponse.pixQrCode,
      expiresAt: new Date(pixResponse.expiresAt),
    });

    return {
      userId: user.id,
      paymentId: payment.id,
      pixCode: pixResponse.pixCode,
      pixQrCode: pixResponse.pixQrCode,
      expiresAt: pixResponse.expiresAt,
      amount: amount / 100,
    };
  }
}
