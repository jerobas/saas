import { getRabbitMQ } from "../../../infra/queue/rabbitmq.js";

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
