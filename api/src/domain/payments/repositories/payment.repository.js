export class PaymentRepository {
  constructor(dataSource) {
    this.dataSource = dataSource;
    this.payload = dataSource.getRepository('Payment');
  }

  async findById(id) {
    return this.payload.findOne({
      where: { id },
      relations: ['user'],
    });
  }

  async findByUserId(userId) {
    return this.payload.find({
      where: { userId },
      order: { createdAt: 'DESC' },
      relations: ['user'],
    });
  }

  async findByAbacatePayPixId(abacatePayPixId) {
    return this.payload.findOne({
      where: { abacatePayPixId },
      relations: ['user'],
    });
  }

  async findByAbacatePayCustomerId(abacatePayCustomerId) {
    return this.payload.find({
      where: { abacatePayCustomerId },
      order: { createdAt: 'DESC' },
      relations: ['user'],
    });
  }

  async findAll() {
    return this.payload.find({
      relations: ['user'],
      order: { createdAt: 'DESC' },
    });
  }

  async create(paymentData) {
    const payment = this.payload.create(paymentData);
    return this.payload.save(payment);
  }

  async update(id, paymentData) {
    await this.payload.update(id, paymentData);
    return this.findById(id);
  }

  async delete(id) {
    await this.payload.delete(id);
  }
}
