export class GetLicenseStatusUseCase {
  constructor(userRepository) {
    this.userRepository = userRepository;
  }

  async execute(input) {
    const { userId, email } = input;

    let user;
    if (userId) {
      user = await this.userRepository.findById(userId);
    } else if (email) {
      user = await this.userRepository.findByEmail(email);
    }

    if (!user) {
      throw new Error('Usuário não encontrado');
    }

    return {
      userId: user.id,
      email: user.email,
      licenseActive: user.licenseActive || false,
      licenseExpiresAt: user.licenseExpiresAt,
      licenseToken: user.licenseToken || null,
    };
  }
}
