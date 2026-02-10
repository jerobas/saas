import { EntitySchema } from 'typeorm';

export class Payment {
  constructor(data) {
    Object.assign(this, data);
  }
}

export const PaymentSchema = new EntitySchema({
  name: 'Payment',
  target: Payment,
  tableName: 'payments',
  columns: {
    id: {
      type: 'uuid',
      primary: true,
      generated: 'uuid',
    },
    userId: {
      type: 'uuid',
      nullable: false,
    },
    abacatePayCustomerId: {
      type: 'varchar',
      nullable: false,
      index: true,
      comment: 'ID do cliente no AbacatePay',
    },
    abacatePayPixId: {
      type: 'varchar',
      nullable: false,
      index: true,
      comment: 'ID da transação PIX no AbacatePay',
    },
    amount: {
      type: 'bigint',
      nullable: false,
      comment: 'Valor em centavos',
    },
    status: {
      type: 'varchar',
      default: 'PENDING',
      comment: 'PENDING, PAID, CANCELLED, EXPIRED',
    },
    pixCode: {
      type: 'text',
      nullable: true,
      comment: 'Código QR do PIX',
    },
    pixQrCode: {
      type: 'text',
      nullable: true,
      comment: 'Imagem do QR code em base64',
    },
    expiresAt: {
      type: 'timestamp',
      nullable: false,
      comment: 'Data de expiração do PIX',
    },
    error: {
      type: 'varchar',
      nullable: true,
      comment: 'Erro retornado do AbacatePay se houver',
    },
    createdAt: {
      type: 'timestamp',
      createDate: true,
    },
    updatedAt: {
      type: 'timestamp',
      updateDate: true,
    },
  },
  relations: {
    user: {
      type: 'many-to-one',
      target: 'User',
      joinColumn: {
        name: 'userId',
      },
    },
  },
});