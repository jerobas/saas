import { useContext, useEffect, useState } from "react";
import { motion } from "framer-motion";
import { QrCode, CopySimple } from "phosphor-react";
import { checkLicenseStatus } from "../services/apiService";
import { AppContext } from "../context/AppContext";
import { ActivateLicense } from "../../wailsjs/go/main/UserService";

function PixPaymentPage() {
  const { pixData } = useContext(AppContext);
  const { amount, expiresAt, pixCode, pixQrCode, userId } = pixData.data || {};
  const [licenseActive, setLicenseActive] = useState(false);

  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const { data } = await checkLicenseStatus(userId, pixData.email);

        if (data.licenseActive) {
          ActivateLicense(data.licenseToken);
          setLicenseActive(true);
          clearInterval(interval);
        }
      } catch (error) {
        console.log("erro ao verificar licença:", error);
      }
    }, 5000);

    return () => clearInterval(interval);
  }, [pixData, pixData.email]);

  const handleCopy = () => {
    navigator.clipboard.writeText(pixCode);
    alert("Código PIX copiado para a área de transferência!");
  };

  if (licenseActive) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <h1 className="text-3xl font-bold text-green-600">
          Licença ativada com sucesso!
        </h1>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-linear-to-br from-pink-50 via-white to-purple-50 flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md"
      >
        <div className="text-center mb-8">
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
            className="inline-flex items-center justify-center w-16 h-16 bg-pink-600 rounded-2xl mb-4"
          >
            <QrCode size={32} color="white" />
          </motion.div>
          <h1 className="text-3xl font-bold text-slate-900 mb-2">
            Pagamento PIX
          </h1>
          <p className="text-slate-600">
            Escaneie o QR Code abaixo para realizar o pagamento
          </p>
        </div>

        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3 }}
          className="bg-white rounded-3xl shadow-xl p-8 border border-slate-100 space-y-6"
        >
          <div className="text-center">
            <img src={pixQrCode} alt="QR Code PIX" className="mx-auto mb-4" />
            <p className="text-lg font-semibold text-slate-700">
              Valor: R$ {amount.toFixed(2).replace(".", ",")}
            </p>
            <p className="text-sm text-slate-500">
              Expira em: {new Date(expiresAt).toLocaleString("pt-BR")}
            </p>
          </div>

          <div className="bg-slate-50 rounded-xl p-4 text-sm text-slate-600">
            <p className="mb-2">
              Copie o código abaixo para realizar o pagamento no seu aplicativo
              bancário:
            </p>
            <div className="relative">
              <textarea
                readOnly
                value={pixCode}
                className="w-full p-2 border border-slate-200 rounded-md text-sm font-mono bg-gray-100 focus:outline-none"
                rows="4"
              />
              <button
                onClick={handleCopy}
                className="absolute top-2 right-2 bg-pink-600 text-white p-2 rounded-md hover:bg-pink-700 transition-all"
              >
                <CopySimple size={16} />
              </button>
            </div>
          </div>
        </motion.div>

        <p className="text-center text-sm text-slate-500 mt-6">
          Após o pagamento, sua licença será ativada automaticamente.
        </p>
      </motion.div>
    </div>
  );
}

export default PixPaymentPage;
