import fs from "node:fs";
import zlib from "node:zlib";
import crypto from "crypto";

const privateKey = fs.readFileSync("./src/license/private.pem");

export class LicenseService {
  /**
   * Gera uma licença assinada digitalmente em formato JWT
   * @param {Object} params Parâmetros
   * @param {string} params.userId ID do usuário (obrigatório)
   * @param {string} params.email Email do usuário (obrigatório)
   * @param {number} [params.days=365] Duração da licença em dias
   * @returns {string} Licença assinada
   * @throws {Error} Se userId ou email forem vazios
   */
  generateLicense({ userId, email, days = 365 }) {
    if (!userId || typeof userId !== "string") {
      throw new Error("userId é obrigatório e deve ser uma string");
    }

    if (!email || typeof email !== "string") {
      throw new Error("email é obrigatório e deve ser uma string");
    }

    if (!email.includes("@")) {
      throw new Error("email inválido");
    }

    const issuedAt = new Date();
    const expiresAt = new Date(Date.now() + days * 86400000);

    const payload = {
      userId,
      email,
      issuedAt: issuedAt.toISOString(),
      expiresAt: expiresAt.toISOString(),
    };

    const packed = zlib.gzipSync(Buffer.from(JSON.stringify(payload)));

    const payload64 = packed.toString("base64");

    const signature = crypto.sign(null, Buffer.from(payload64), privateKey);

    return `${payload64}.${signature.toString("base64")}`;
  }
}
