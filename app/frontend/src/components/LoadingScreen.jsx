import { motion } from "framer-motion";
import { SpinnerGap } from "phosphor-react";

function LoadingScreen() {
  return (
    <div className="min-h-screen bg-linear-to-br from-pink-50 via-white to-purple-50 flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md text-center"
      >
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
          className="inline-flex items-center justify-center w-16 h-16 bg-pink-600 rounded-2xl mb-4"
        >
          <SpinnerGap size={32} color="white" className="animate-spin" />
        </motion.div>
        <h1 className="text-3xl font-bold text-slate-900 mb-2">
          Carregando...
        </h1>
        <p className="text-slate-600">
          Aguarde enquanto verificamos suas informações
        </p>
      </motion.div>
    </div>
  );
}

export default LoadingScreen;
