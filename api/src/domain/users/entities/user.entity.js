import { EntitySchema } from "typeorm";

export class User {
  constructor(data) {
    Object.assign(this, data);
  }
}

export const UserSchema = new EntitySchema({
  name: "User",
  target: User,
  tableName: "users",
  columns: {
    id: {
      type: "uuid",
      primary: true,
      generated: "uuid",
    },
    email: {
      type: "varchar",
      unique: true,
      nullable: false,
    },
    name: {
      type: "varchar",
      nullable: false,
    },
    taxId: {
      type: "varchar",
      nullable: true,
    },
    cellphone: {
      type: "varchar",
      nullable: true,
    },
    abacatePayCustomerId: {
      type: "varchar",
      nullable: true,
      comment: "ID do customer no AbacatePay",
    },
    licenseActive: {
      type: "boolean",
      default: false,
      comment: "Se a licença está ativa",
    },
    licenseExpiresAt: {
      type: "timestamp",
      nullable: true,
      comment: "Data de expiração da licença (1 ano após pagamento confirmado)",
    },
    licenseToken: {
      type: "text",
      nullable: true,
      comment: "Token de licença assinado digitalmente",
    },
    createdAt: {
      type: "timestamp",
      createDate: true,
    },
    updatedAt: {
      type: "timestamp",
      updateDate: true,
    },
    passwordHash: {
      type: "varchar",
      nullable: false,
      comment: "Hash da senha do usuário",
    },
  },
  relations: {
    payments: {
      type: "one-to-many",
      target: "Payment",
      inverseSide: "user",
      cascade: true,
    },
  },
});
