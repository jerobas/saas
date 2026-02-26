const AuthenticateAndCheckLicenseUseCase = require('../../domain/users/usecases/authenticate-and-check-license-usecase');
const UserRepository = require('../../domain/users/repositories/user.repository');
const LicenseService = require('../../infra/services/license.service');
const PaymentService = require('../../infra/services/abacatepay.service');

const userRepository = new UserRepository();
const licenseService = new LicenseService();
const paymentService = new PaymentService();

const authenticateAndCheckLicenseUseCase = new AuthenticateAndCheckLicenseUseCase(
  userRepository,
  licenseService,
  paymentService
);

async function authenticateAndCheckLicenseController(req, res, next) {
  try {
    const { email, password } = req.body;

    const result = await authenticateAndCheckLicenseUseCase.execute({ email, password });

    res.status(200).json(result);
  } catch (error) {
    next(error);
  }
}

module.exports = authenticateAndCheckLicenseController;