import fs from "node:fs";
import crypto from "node:crypto";

const privateKeyPem = fs.readFileSync("./src/license/private.pem", "utf8");
const key = crypto.createHash("sha256").update(privateKeyPem, "utf8").digest();

export class PasswordService {
  /**
   * @param {string} password
   * @returns {string}
   */
  hash(password) {
    if (!password || typeof password !== "string") {
      throw new Error("password é obrigatório e deve ser uma string");
    }

    return crypto.createHmac("sha256", key).update(password, "utf8").digest("base64");
  }

  /**
   * @param {string} password
   * @param {string} storedHash
   * @returns {boolean}
   */
  compare(password, storedHash) {
    const computed = this.hash(password ?? "");
    const a = Buffer.from(computed ?? "", "utf8");
    const b = Buffer.from(storedHash ?? "", "utf8");

    if (a.length !== b.length) {
      return false;
    }

    return crypto.timingSafeEqual(a, b);
  }
}
