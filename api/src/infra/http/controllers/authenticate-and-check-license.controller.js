import { AuthenticateAndCheckLicenseUseCase } from '../../../domain/users/usecases/authenticate-and-check-license-usecase.js';

export class AuthenticateAndCheckLicenseController {
  constructor(userRepository, licenseService, paymentService) {
    this.authenticateAndCheckLicenseUseCase = new AuthenticateAndCheckLicenseUseCase(
      userRepository,
      licenseService,
      paymentService
    );
  }

  async handle(req, res, next) {
    try {
      const { email, password } = req.body;

      const result = await this.authenticateAndCheckLicenseUseCase.execute({
        email,
        password,
      });

      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }
}