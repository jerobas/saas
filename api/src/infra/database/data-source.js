import { DataSource } from "typeorm";
import { UserSchema } from "../../domain/users/entities/user.entity.js";
import { PaymentSchema } from "../../domain/payments/entities/payment.entity.js";

const dataSourceConfig = {
  type: "postgres",
  url: process.env.DATABASE_URL,
  ssl: {
    rejectUnauthorized: false,
  },
};

export const AppDataSource = new DataSource({
  ...dataSourceConfig,
  synchronize: process.env.NODE_ENV !== "production",
  entities: [UserSchema, PaymentSchema],
  subscribers: [],
});
