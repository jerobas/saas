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
  const { userId, email, name, taxId, cellphone, paymentMethod } = payload;

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

  await rabbit.send({
    type:
      paymentMethod === "CARD"
        ? "CREATE_CARD_BILLING_STRATEGY"
        : "CREATE_PIX_STRATEGY",
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

const processCardBillingStrategy = async (payload) => {
  const { userId, email, name, taxId, cellphone, customerId } = payload;

  const billing = await abacatePayService.createCardBilling({
    customerId,
    customer: { name, email, taxId, cellphone },
    externalId: `billing_${userId}`,
  });

  const payment = await paymentRepository.create({
    userId,
    abacatePayCustomerId: customerId,
    abacatePayBillingId: billing.paymentId,
    amount: billing.amount,
    status: "PENDING",
    paymentMethod: "CARD",
    paymentUrl: billing.paymentUrl,
    billingFrequency: billing.frequency,
    billingMethods: billing.methods,
    billingProducts: billing.products,
  });

  // TO-DO: Nao sei se realmente precisa do SSE nesse caso, ja que o pagamento via cartao tem um fluxo diferente e o cliente sera redirecionado para a pagina de pagamento. Mas vou deixar aqui caso queira usar para atualizar o status do pagamento em tempo real na interface do usuario
  notifySSE({
    clientId: userId,
    status: "CARD_BILLING_CREATED",
    billing: {
      billingId: billing.paymentId,
      paymentUrl: billing.paymentUrl,
      amount: billing.amount / 100,
      createdAt: payment.createdAt,
    },
  });
};

(async () => {
  await rabbit.consume(async (payload) => {
    if (payload.type === "CREATE_USER_STRATEGY") {
      await processUserCreationStrategy(payload.payload);
    } else if (payload.type === "CREATE_PIX_STRATEGY") {
      await processPixCreationStrategy(payload.payload);
    } else if (payload.type === "CREATE_CARD_BILLING_STRATEGY") {
      await processCardBillingStrategy(payload.payload);
    }
  });
})();
