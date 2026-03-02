import { Router } from "express";
import { AppDataSource } from "../../database/data-source.js";

// Repositories
import { UserRepository } from "../../../domain/users/repositories/user.repository.js";
import { PaymentRepository } from "../../../domain/payments/repositories/payment.repository.js";

// Services
import { AbacatePayService } from "../../services/abacatepay.service.js";
import { LicenseService } from "../../services/license.service.js";
import { PasswordService } from "../../services/password.service.js";

// Controllers
import { CreateUserController } from "../controllers/create-user.controller.js";
import { RenewSubscriptionController } from "../controllers/renew-subscription.controller.js";
import { SimulatePaymentController } from "../controllers/simulate-payment.controller.js";
import { WebhookPaymentController } from "../controllers/webhook-payment.controller.js";
import { GetLicenseStatusController } from "../controllers/get-license-status.controller.js";
import { sendSSE } from "../controllers/sse.controller.js";
import { AuthenticateAndCheckLicenseController } from "../controllers/authenticate-and-check-license.controller.js";

const router = Router();

// Factory para criar instâncias com dependências injetadas
const createControllers = () => {
  const userRepository = new UserRepository(AppDataSource);
  const paymentRepository = new PaymentRepository(AppDataSource);
  const abacatePayService = new AbacatePayService();
  const licenseService = new LicenseService();
  const passwordService = new PasswordService();

  return {
    createUserController: new CreateUserController(
      userRepository,
      paymentRepository,
      abacatePayService,
      passwordService,
    ),
    renewSubscriptionController: new RenewSubscriptionController(
      userRepository,
      paymentRepository,
      abacatePayService,
    ),
    simulatePaymentController: new SimulatePaymentController(
      paymentRepository,
      userRepository,
      abacatePayService,
      licenseService,
    ),
    webhookPaymentController: new WebhookPaymentController(
      paymentRepository,
      userRepository,
      abacatePayService,
      licenseService,
    ),
    getLicenseStatusController: new GetLicenseStatusController(userRepository),
    authenticateAndCheckLicenseController: new AuthenticateAndCheckLicenseController(
      userRepository,
      licenseService,
      abacatePayService,
    ),
  };
};

// Lazy load controllers para evitar inicialização prematura
let controllers = null;

const getControllers = () => {
  if (!controllers) {
    controllers = createControllers();
  }
  return controllers;
};

// POST /api/create - Criar usuário + gerar PIX
router.post("/create", async (req, res, next) => {
  try {
    const ctrl = getControllers();
    await ctrl.createUserController.handle(req, res);
  } catch (error) {
    next(error);
  }
});

// POST /api/renew - Renovar assinatura
router.post("/renew", async (req, res, next) => {
  try {
    const ctrl = getControllers();
    await ctrl.renewSubscriptionController.handle(req, res);
  } catch (error) {
    next(error);
  }
});

// POST /api/dev/simulate - Simular pagamento (apenas em desenvolvimento)
router.post("/dev/simulate", async (req, res, next) => {
  try {
    const ctrl = getControllers();
    await ctrl.simulatePaymentController.handle(req, res);
  } catch (error) {
    next(error);
  }
});

// POST /api/webhooks/payment - Webhook de pagamento do AbacatePay
router.post("/webhooks/payment", async (req, res, next) => {
  try {
    const ctrl = getControllers();
    await ctrl.webhookPaymentController.handle(req, res);
  } catch (error) {
    next(error);
  }
});

// GET /api/license - Verificar status da licença do usuário
router.get("/license", async (req, res, next) => {
  try {
    const ctrl = getControllers();
    await ctrl.getLicenseStatusController.handle(req, res);
  } catch (error) {
    next(error);
  }
});

// GET /api/sse/:clientId - SSE para atualizações em tempo real
router.get("/sse/:clientId", sendSSE);

// POST /api/auth/check-license - Autenticar usuário e verificar a licença
router.post("/auth/check-license", async (req, res, next) => {
  try {
    const ctrl = getControllers();
    await ctrl.authenticateAndCheckLicenseController.handle(req, res);
  } catch (error) {
    next(error);
  }
});

// GET /api/health - Verificar status da API
router.get("/health", (req, res) => {
  res.status(200).json({ status: "OK", timestamp: new Date().toISOString() });
});

export const setupRoutes = (app) => {
  app.use("/api", router);
};
