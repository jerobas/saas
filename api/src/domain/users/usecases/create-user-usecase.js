import { getRabbitMQ } from "../../../infra/queue/rabbitmq.js";

export class CreateUserUseCase {
  constructor(
    userRepository,
    paymentRepository,
    abacatePayService,
    passwordService,
  ) {
    this.userRepository = userRepository;
    this.paymentRepository = paymentRepository;
    this.abacatePayService = abacatePayService;
    this.passwordService = passwordService;
  }

  async execute(input) {
    const { email, name, taxId, cellphone, password } = input;

    if (!password) {
      throw new Error("Senha é obrigatória");
    }

    const existingUser = await this.userRepository.findByEmail(email);
    if (existingUser) {
      throw new Error(`Email ${email} já cadastrado`);
    }

    const passwordHash = await this.passwordService.hash(password);

    const user = await this.userRepository.create({
      email,
      name,
      taxId,
      cellphone,
      passwordHash,
    });

    await getRabbitMQ().send({
      type: "CREATE_USER_STRATEGY",
      payload: {
        userId: user.id,
        email,
        name,
        taxId,
        cellphone,
      },
    });

    return {
      userId: user.id,
    };
  }
}
