export class CleanupPendingUsersUseCase {
  constructor(userRepository) {
    this.userRepository = userRepository;
  }

  async execute(timeoutInHours = 24) {
    const expirationDate = new Date();
    expirationDate.setHours(expirationDate.getHours() - timeoutInHours);

    const pendingUsers = await this.userRepository.find({
      where: {
        isPending: true,
        createdAt: { $lt: expirationDate },
      },
    });

    for (const user of pendingUsers) {
      await this.userRepository.remove(user);
    }
  }
}
