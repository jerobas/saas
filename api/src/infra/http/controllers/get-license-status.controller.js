import { GetLicenseStatusUseCase } from '../../../domain/users/usecases/get-license-status-usecase.js';

export class GetLicenseStatusController {
  constructor(userRepository) {
    this.getLicenseStatusUseCase = new GetLicenseStatusUseCase(userRepository);
  }

  async handle(req, res) {
    try {
      const { userId, email } = req.query;

      if (!userId && !email) {
        return res.status(400).json({
          error: 'userId ou email é obrigatório (como query param)',
        });
      }

      const result = await this.getLicenseStatusUseCase.execute({
        userId,
        email,
      });

      return res.status(200).json({
        success: true,
        data: result,
      });
    } catch (error) {
      console.error('[GetLicenseStatusController]', error);
      return res.status(404).json({
        error: error.message,
      });
    }
  }
}
