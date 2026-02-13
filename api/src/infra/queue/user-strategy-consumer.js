import RabbitMQService from './rabbitmq';
import { notifySSE } from '../http/controllers/sse.controller';
import { AbacatePayService } from '../services/abacatepay.service';

const abacatePayService = new AbacatePayService();

async function processUserCreationStrategy(message) {
  const { userId, email, name, taxId, cellphone } = message;

  try {
    console.log(`Iniciando processamento para o usuário: ${userId}`);

    const customerResponse = await abacatePayService.createCustomer({
      name,
      email,
      taxId,
      cellphone,
    });

    if (!customerResponse || customerResponse.error) {
      throw new Error(`Erro ao criar cliente no AbacatePay: ${customerResponse?.error}`);
    }

    const abacatePayCustomerId = customerResponse.id;

    await userRepository.update(userId, { abacatePayCustomerId });

    const amount = parseInt(process.env.LICENSE_PRICE || '50000');
    const pixResponse = await abacatePayService.createPixQrCode({
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

    if (!pixResponse || pixResponse.error) {
      throw new Error(`Erro ao criar PIX: ${pixResponse?.error}`);
    }

    const payment = await paymentRepository.create({
      userId,
      abacatePayCustomerId,
      abacatePayPixId: pixResponse.id,
      amount,
      status: 'PENDING',
      pixCode: pixResponse.pixCode,
      pixQrCode: pixResponse.pixQrCode,
      expiresAt: new Date(pixResponse.expiresAt),
    });

    notifySSE({
      clientId: userId,
      status: 'PIX_CREATED',
      payment: {
        paymentId: payment.id,
        pixCode: pixResponse.pixCode,
        pixQrCode: pixResponse.pixQrCode,
        expiresAt: pixResponse.expiresAt,
        amount: amount / 100,
      },
    });

    console.log(`Processamento concluído para o usuário: ${userId}`);
  } catch (error) {
    console.error(`Erro ao processar estratégia para o usuário ${userId}:`, error);
    throw error;
  }
}

(async () => {
  await RabbitMQService.connect();
  await RabbitMQService.consumeMessages(async (message) => {
    try {
      await processUserCreationStrategy(message);
    } catch (error) {
      console.error('Erro no processamento, mensagem será reencaminhada para retry:', error);
      RabbitMQService.channel.nack(message, false, false);
    }
  });
})();