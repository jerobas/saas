import { CleanupPendingUsersUseCase } from "../../domain/users/usecases/cleanup-pending-users-usecase.js";
import { UserRepository } from "../../domain/users/repositories/user.repository.js";
import { AppDataSource } from "../database/data-source.js";

const cleanupJob = async () => {
  const userRepository = new UserRepository(AppDataSource);
  const cleanupUseCase = new CleanupPendingUsersUseCase(userRepository);

  try {
    console.log("Iniciando limpeza de cadastros pendentes...");
    await cleanupUseCase.execute(24);
    console.log("Limpeza de cadastros pendentes concluída.");
  } catch (error) {
    console.error("Erro ao executar limpeza de cadastros pendentes:", error);
  }
};

setInterval(cleanupJob, 60 * 60 * 1000);
