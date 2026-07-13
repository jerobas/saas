import axios from "axios";
import { AppError } from "../http/errors/app-error.js";

export class AbacatePayService {
  constructor() {
    this.baseURL = process.env.ABACATEPAY_BASE_URL;
    this.apiKey = process.env.ABACATEPAY_API_KEY;
    this.webhookSecret = process.env.ABACATEPAY_WEBHOOK_SECRET;

    if (!this.apiKey) {
      throw new Error("ABACATEPAY_API_KEY não configurada");
    }

    this.client = axios.create({
      baseURL: this.baseURL,
      headers: {
        Authorization: `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
      },
    });
  }

  /**
   * Cria um cliente no AbacatePay
   */
  createCustomer = async ({ name, email, taxId, cellphone }) => {
    try {
      const response = await this.client.post("/customer/create", {
        name,
        email,
        taxId: taxId || "",
        cellphone: cellphone || "",
      });

      if (
        !response.data ||
        response.data.success === false ||
        !response.data.data?.id
      ) {
        console.error(
          "Erro ao criar cliente no AbacatePay:",
          response.data.error,
        );
        throw new AppError(
          `Erro ao criar cliente no AbacatePay: ${response.data.error}`,
          403,
        );
      }
      return {
        id: response.data.data?.id,
        name: response.data.data?.metadata?.name,
        email: response.data.data?.metadata?.email,
      };
    } catch (error) {
      console.error(
        "AbacatePay - Erro ao criar cliente:",
        error.response?.data || error.message,
      );
      throw new AppError(
        "Erro ao criar cliente no AbacatePay",
        error.response?.status || 500,
      );
    }
  };

  /**
   * Lista clientes no AbacatePay
   */
  listCustomers = async () => {
    try {
      const response = await this.client.get("/customers/list");
      return response.data || [];
    } catch (error) {
      console.error(
        "AbacatePay - Erro ao listar clientes:",
        error.response?.data || error.message,
      );
      throw new AppError("Erro ao listar clientes no AbacatePay", 500);
    }
  };

  /**
   * Encontra cliente por email
   */
  findCustomerByEmail = async (email) => {
    try {
      const customers = await this.listCustomers();
      return customers.find((c) => c.email === email);
    } catch (error) {
      console.error(
        "AbacatePay - Erro ao buscar cliente por email:",
        error.message,
      );
      return null;
    }
  };

  /**
   * Cria um QR Code PIX para pagamento
   */
  createPixQrCode = async ({
    amount,
    description,
    expiresInMinutes = 30,
    metadata = {},
    customer,
  }) => {
    try {
      const expiresIn = expiresInMinutes * 60;

      const response = await this.client.post("/pixQrCode/create", {
        amount,
        expiresIn,
        description,
        customer,
        metadata,
      });

      if (
        !response.data ||
        response.data.success === false ||
        !response.data.data?.id
      ) {
        console.error(
          "AbacatePay - Resposta da API ao criar QR Code PIX:",
          response.data,
        );
        throw new AppError("Falha ao criar QR Code PIX", 500);
      }

      return {
        id: response.data.data.id,
        pixCode: response.data.data.brCode || "",
        pixQrCode: response.data.data.brCodeBase64 || "",
        amount: response.data.data.amount,
        expiresAt:
          response.data.data.expiresAt ||
          new Date(Date.now() + expiresInMinutes * 60 * 1000),
      };
    } catch (error) {
      console.error(
        "AbacatePay - Erro ao criar QR Code PIX:",
        error.response?.data || error.message,
      );
      throw new AppError(
        "Erro ao criar QR Code PIX",
        error.response?.status || 500,
      );
    }
  };

  /**
   * Cria uma cobranca por cartao no AbacatePay
   */
  createCardBilling = async ({
    allowCoupons = false,
    coupons = [],
    customerId,
    customer,
    externalId,
    metadata = {},
  }) => {
    try {
      const frequency = "MULTIPLE_PAYMENTS";
      const methods = ["CARD"];
      const returnUrl = process.env.CARD_BILLING_RETURN_URL;
      const completionUrl = process.env.CARD_BILLING_COMPLETION_URL;
      const productPrice = Number(
        process.env.ABACATEPAY_AMOUNT || process.env.LICENSE_PRICE || 50000,
      );
      const productQuantity = Number(
        process.env.ABACATEPAY_PRODUCT_QUANTITY || 1,
      );
      const productExternalId =
        process.env.ABACATEPAY_PRODUCT_ID || externalId || "license_product";
      const productName =
        process.env.ABACATEPAY_PRODUCT_NAME || "Licença de Software";
      const productDescription =
        process.env.ABACATEPAY_PRODUCT_DESCRIPTION || "Licença de Software";

      const normalizedProducts = [
        {
          externalId: productExternalId,
          name: productName,
          description: productDescription,
          quantity: productQuantity,
          price: productPrice,
        },
      ];

      const response = await this.client.post("/billing/create", {
        frequency,
        methods,
        products: normalizedProducts,
        returnUrl,
        completionUrl,
        allowCoupons,
        coupons,
        customerId,
        customer,
        externalId,
        metadata,
      });

      if (
        !response.data ||
        response.data.success === false ||
        !response.data.data
      ) {
        console.error(
          "AbacatePay - Resposta da API ao criar billing de cartao:",
          response.data,
        );
        throw new AppError("Falha ao criar cobranca de cartao", 500);
      }

      if (response.data.error) {
        throw new AppError(response.data.error, 400);
      }

      const billingData = response.data.data;
      const paymentId =
        billingData.id || billingData.paymentId || billingData.billingId;
      const paymentUrl =
        billingData.url ||
        billingData.paymentUrl ||
        billingData.checkoutUrl ||
        billingData.redirectUrl;

      if (!paymentId || !paymentUrl) {
        console.error(
          "AbacatePay - Dados insuficientes no billing de cartao:",
          billingData,
        );
        throw new AppError(
          "Resposta invalida ao criar cobranca de cartao",
          500,
        );
      }

      return {
        paymentId,
        paymentUrl,
        amount: billingData.amount || productPrice,
        frequency: billingData.frequency || frequency,
        methods: billingData.methods || methods,
        products: normalizedProducts,
        status: billingData.status || "PENDING",
      };
    } catch (error) {
      if (error instanceof AppError) {
        throw error;
      }

      const apiErrorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message;

      console.error(
        "AbacatePay - Erro ao criar billing de cartao:",
        error.response?.data || apiErrorMessage,
      );
      throw new AppError(
        apiErrorMessage || "Erro ao criar cobranca de cartao no AbacatePay",
        error.response?.status || error.statusCode || 500,
      );
    }
  };

  /**
   * Simula um pagamento PIX (apenas em DEVELOPMENT)
   * Útil para testes
   */
  simulatePixPayment = async (pixId, metadata = {}) => {
    if (
      process.env.NODE_ENV !== "development" &&
      process.env.NODE_ENV !== "testing"
    ) {
      console.error(
        "❌ Simulação de pagamento só permitida em desenvolvimento",
      );
      throw new AppError(
        "Payment simulation is only available in development environment",
        403,
      );
    }

    try {
      const response = await this.client.post(
        `/pixQrCode/simulate-payment?id=${pixId}`,
        { metadata },
      );

      console.log(`✅ Pagamento PIX simulado (DEV only): ${pixId}`);
      return {
        id: response.data.id,
        status: "paid",
        paidAt: new Date(),
      };
    } catch (error) {
      console.error(
        "AbacatePay - Erro ao simular pagamento:",
        error.response?.data || error.message,
      );
      throw new AppError(
        "Erro ao simular pagamento PIX",
        error.response?.status || 500,
      );
    }
  };
}
