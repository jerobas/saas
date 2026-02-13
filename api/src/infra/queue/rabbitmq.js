import amqp from 'amqplib';

class RabbitMQService {
  connection = null;
  channel = null;
  queue = 'user_creation_queue';

  async connect() {
    if (this.channel) return;

    this.connection = await amqp.connect(process.env.RABBITMQ_URL);

    this.connection.on('close', () => {
      console.error('RabbitMQ connection closed');
      this.connection = null;
      this.channel = null;
    });

    this.connection.on('error', console.error);

    this.channel = await this.connection.createChannel();

    const deadLetterExchange = 'dead_letter_exchange';
    const deadLetterQueue = 'dead_letter_queue';

    await this.channel.assertExchange(deadLetterExchange, 'direct', { durable: true });
    await this.channel.assertQueue(deadLetterQueue, { durable: true });
    await this.channel.bindQueue(deadLetterQueue, deadLetterExchange, 'retry');

    await this.channel.assertQueue(this.queue, {
      durable: true,
      arguments: {
        'x-dead-letter-exchange': deadLetterExchange,
        'x-dead-letter-routing-key': 'retry',
        'x-message-ttl': 60000
      }
    });
  }

  async send(message) {
    await this.connect();

    this.channel.sendToQueue(
      this.queue,
      Buffer.from(JSON.stringify(message)),
      { persistent: true }
    );
  }

  async consume(handler) {
    await this.connect();

    this.channel.consume(this.queue, async msg => {
      if (!msg) return;

      try {
        const payload = JSON.parse(msg.content.toString());
        await handler(payload);
        this.channel.ack(msg);
      } catch (err) {
        console.error(err);
        this.channel.nack(msg, false, false);
      }
    });
  }

  async close() {
    await this.channel?.close();
    await this.connection?.close();
  }
}

let instance;

export const getRabbitMQ = () => {
  if (!instance) instance = new RabbitMQService();
  return instance;
};
