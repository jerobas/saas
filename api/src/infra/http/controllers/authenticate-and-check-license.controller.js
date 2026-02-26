import AuthenticateAndCheckLicenseUseCase from '../../../domain/users/usecases/authenticate-and-check-license-usecase.js';
import { UserRepository } from '../../../domain/users/repositories/user.repository.js';
import { LicenseService } from '../../infra/services/license.service.js';
import { PaymentService } from '../../infra/services/abacatepay.service.js';

const userRepository = new UserRepository();
const licenseService = new LicenseService();
const paymentService = new PaymentService();

const authenticateAndCheckLicenseUseCase = new AuthenticateAndCheckLicenseUseCase(
  userRepository,
  licenseService,
  paymentService
);

export class AuthenticateAndCheckLicenseController {
  async handle(req, res, next) {
    try {
      const { email, password } = req.body;

      const result = await authenticateAndCheckLicenseUseCase.execute({
        email,
        password,
      });

      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }
}