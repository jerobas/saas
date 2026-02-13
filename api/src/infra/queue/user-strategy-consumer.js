import { getRabbitMQ } from "./rabbitmq.js";
import { AppDataSource } from "../database/data-source.js";
import { AbacatePayService } from "../services/abacatepay.service.js";
import { UserRepository } from "../../domain/users/repositories/user.repository.js";
import { PaymentRepository } from "../../domain/payments/repositories/payment.repository.js";
import { notifySSE } from "../http/controllers/sse.controller.js";

const userRepository = new UserRepository(AppDataSource);
const paymentRepository = new PaymentRepository(AppDataSource);
const abacatePayService = new AbacatePayService();
const rabbit = getRabbitMQ();

const processPixCreationStrategy = async (payload) => {
  const { userId, email, name, taxId, cellphone, customerId } = payload;

  const amount = Number(process.env.LICENSE_PRICE || 50000);

  const pix = await abacatePayService.createPixQrCode({
    amount,
    description: "Licença de Software",
    expiresInMinutes: 30,
    customer: { name, email, taxId, cellphone },
  });

  if (pix?.error) throw new Error(pix.error);

  const payment = await paymentRepository.create({
    userId,
    abacatePayCustomerId: customerId,
    abacatePayPixId: pix.id,
    amount,
    status: "PENDING",
    pixCode: pix.pixCode,
    pixQrCode: pix.pixQrCode,
    expiresAt: new Date(pix.expiresAt),
  });

  notifySSE({
    clientId: userId,
    status: "PIX_CREATED",
    payment: {
      paymentId: payment.id,
      pixCode: pix.pixCode,
      pixQrCode: pix.pixQrCode,
      expiresAt: pix.expiresAt,
      amount: amount / 100,
    },
  });
};

const processUserCreationStrategy = async (payload) => {
  const { userId, email, name, taxId, cellphone } = payload;

  console.log(`Processing user ${userId}`);

  const customer = await abacatePayService.createCustomer({
    name,
    email,
    taxId,
    cellphone,
  });

  if (customer?.error) throw new Error(customer.error);

  await userRepository.update(userId, {
    abacatePayCustomerId: customer.id,
  });

  // Enviar para a fila de criação de Pix
  await rabbit.send({
    type: "CREATE_PIX_STRATEGY",
    payload: {
      userId,
      email,
      name,
      taxId,
      cellphone,
      customerId: customer.id,
    },
  });
};

(async () => {
  await rabbit.consume(async (payload) => {
    if (payload.type === "CREATE_USER_STRATEGY") {
      await processUserCreationStrategy(payload.payload);
    } else if (payload.type === "CREATE_PIX_STRATEGY") {
      await processPixCreationStrategy(payload.payload);
    }
  });
})();
