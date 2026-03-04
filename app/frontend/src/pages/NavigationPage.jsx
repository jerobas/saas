import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import { User, SignIn } from "phosphor-react";
import { useEffect, useState } from "react";
import { Bell } from "phosphor-react";

function NavigationPage() {
  const navigate = useNavigate();
  const [showReminder, setShowReminder] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      setShowReminder(true);
    }, 30000); // 30 segundos de inatividade

    return () => clearTimeout(timer);
  }, []);

  return (
    <div className="relative min-h-screen bg-gradient-to-br from-pink-50 via-white to-purple-50 flex items-center justify-center p-4 overflow-hidden">

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="relative w-full max-w-md text-center bg-white rounded-3xl shadow-xl p-8 border border-slate-100"
      >
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
          className="inline-flex items-center justify-center w-16 h-16 bg-pink-600 rounded-2xl mb-4 mx-auto"
        >
          <User size={32} weight="bold" className="text-white" />
        </motion.div>

        <h1 className="text-3xl font-bold text-slate-900 mb-4">Bem-vindo</h1>

        <p className="text-slate-600 mb-8">Escolha uma opção para continuar</p>

        <div className="space-y-4">
          <button
            onClick={() => navigate("/login")}
            className="w-full bg-pink-600 text-white py-3 rounded-xl font-semibold hover:bg-pink-700 active:scale-95 transition-all flex items-center justify-center gap-2"
          >
            <SignIn size={20} /> Fazer Login
          </button>

          <button
            onClick={() => navigate("/cadastro")}
            className="w-full bg-indigo-600 text-white py-3 rounded-xl font-semibold hover:bg-indigo-700 active:scale-95 transition-all flex items-center justify-center gap-2"
          >
            <User size={20} /> Cadastrar-se
          </button>
        </div>
      </motion.div>

      {showReminder && (
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5, type: "spring", stiffness: 100 }}
          className="absolute bottom-8 right-8 bg-yellow-100 text-yellow-800 p-4 rounded-xl shadow-lg flex items-center gap-2"
        >
          <Bell size={24} weight="fill" className="text-yellow-500" />
          <span>Ei! Não se esqueça de escolher uma opção 😊</span>
        </motion.div>
      )}
    </div>
  );
}

export default NavigationPage;