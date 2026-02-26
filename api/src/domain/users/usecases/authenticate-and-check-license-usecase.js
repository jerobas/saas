class AuthenticateAndCheckLicenseUseCase {
  constructor(userRepository, licenseService, paymentService) {
    this.userRepository = userRepository;
    this.licenseService = licenseService;
    this.paymentService = paymentService;
  }

  async execute({ email, password }) {
    // Verificar se o usuário existe
    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      throw new Error('Usuário não encontrado');
    }

    // Verificar se a senha está correta
    const isPasswordValid = await this.userRepository.validatePassword(user, password);
    if (!isPasswordValid) {
      throw new Error('Senha inválida');
    }

    // Verificar status da licença
    const licenseStatus = await this.licenseService.getLicenseStatus(user.id);
    if (licenseStatus === 'active') {
      return { message: 'Acesso liberado', licenseStatus };
    }

    // Gerar novo Pix para ativação
    const payment = await this.paymentService.generatePix(user.id);
    return {
      message: 'Licença inativa. Um novo Pix foi gerado.',
      licenseStatus,
      payment,
    };
  }
}

module.exports = AuthenticateAndCheckLicenseUseCase;