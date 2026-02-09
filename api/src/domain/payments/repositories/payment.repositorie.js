import { AppDataSource } from '../../../infra/database/data-source.js';
import { Payment } from '../entities/payment.entity.js';

export class PaymentRepository {
  constructor() {
    this.repository = AppDataSource.getRepository(Payment);
  }

  findById = async (id) => {
    return await this.repository.findOne({
      where: { id },
      relations: ['client'],
    });
  };

  findByExternalId = async (externalId) => {
    return await this.repository.findOne({
      where: { externalId },
      relations: ['client'],
    });
  };

  findByClientId = async (clientId) => {
    return await this.repository.find({
      where: { clientId },
      order: { createdAt: 'DESC' },
    });
  };

  create = async (data) => {
    const payment = this.repository.create(data);
    return await this.repository.save(payment);
  };

  update = async (id, data) => {
    await this.repository.update(id, data);
    return await this.findById(id);
  };

  markAsPaid = async (id) => {
    return await this.update(id, {
      status: 'paid',
      paidAt: new Date(),
    });
  };
}