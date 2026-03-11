import { EntitySchema } from "typeorm";

export class Payment {
  constructor(data) {
    Object.assign(this, data);
  }
}

export const PaymentSchema = new EntitySchema({
  name: "Payment",
  target: Payment,
  tableName: "payments",
  columns: {
    id: {
      type: "uuid",
      primary: true,
      generated: "uuid",
    },
    userId: {
      type: "uuid",
      nullable: false,
    },
    abacatePayCustomerId: {
      type: "varchar",
      nullable: false,
      index: true,
      comment: "ID do cliente no AbacatePay",
    },
    abacatePayPixId: {
      type: "varchar",
      nullable: true,
      index: true,
      comment: "ID da transação PIX no AbacatePay",
    },
    abacatePayBillingId: {
      type: "varchar",
      nullable: true,
      index: true,
      comment: "ID do billing no AbacatePay para pagamento com cartao",
    },
    amount: {
      type: "bigint",
      nullable: false,
      comment: "Valor em centavos",
    },
    paymentMethod: {
      type: "varchar",
      default: "PIX",
      comment: "PIX ou CARD",
    },
    status: {
      type: "varchar",
      default: "PENDING",
      comment: "PENDING, PAID, CANCELLED, EXPIRED",
    },
    pixCode: {
      type: "text",
      nullable: true,
      comment: "Código QR do PIX",
    },
    pixQrCode: {
      type: "text",
      nullable: true,
      comment: "Imagem do QR code em base64",
    },
    paymentUrl: {
      type: "text",
      nullable: true,
      comment: "URL de checkout para pagamentos com cartao",
    },
    billingFrequency: {
      type: "varchar",
      nullable: true,
      comment: "Frequencia usada no billing do cartao (ex: MULTIPLE_PAYMENTS)",
    },
    billingMethods: {
      type: "simple-json",
      nullable: true,
      comment: 'Metodos de pagamento usados no billing (ex: ["CARD"])',
    },
    billingProducts: {
      type: "simple-json",
      nullable: true,
      comment: "Produtos enviados para criacao do billing no AbacatePay",
    },
    expiresAt: {
      type: "timestamp",
      nullable: true,
      comment: "Data de expiração do pagamento (geralmente PIX)",
    },
    error: {
      type: "varchar",
      nullable: true,
      comment: "Erro retornado do AbacatePay se houver",
    },
    createdAt: {
      type: "timestamp",
      createDate: true,
    },
    updatedAt: {
      type: "timestamp",
      updateDate: true,
    },
  },
  relations: {
    user: {
      type: "many-to-one",
      target: "User",
      joinColumn: {
        name: "userId",
      },
    },
  },
});
