import { useState } from "react";
import { motion } from "framer-motion";
import { EnvelopeSimple, Lock } from "phosphor-react";
import { useNavigate } from "react-router-dom";

function LoginPage() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validateForm = () => {
    const newErrors = {};
    if (!formData.email.trim() || !/^[^\s@]+@[\w-]+\.[a-z]{2,}$/i.test(formData.email)) {
      newErrors.email = "Email inválido";
    }
    if (!formData.password.trim()) {
      newErrors.password = "A senha é obrigatória";
    }
    return newErrors;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
    setErrors({ ...errors, [name]: "" });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const validationErrors = validateForm();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setIsSubmitting(true);
    try {
      // Simulate login logic
      console.log("Logging in with", formData);
      // Navigate to another page on success
      navigate("/dashboard");
    } catch (error) {
      console.error("Erro ao fazer login:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

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
            <Lock size={32} weight="bold" className="text-white" />
          </motion.div>
          <h1 className="text-3xl font-bold text-slate-900 mb-2">Login</h1>
          <p className="text-slate-600">Entre com suas credenciais</p>
        </div>

        <motion.form
          initial={{ opacity: 0, x: -16, scale: 0.98 }}
          animate={{ opacity: 1, x: 0, scale: 1 }}
          exit={{ opacity: 0, x: 16, scale: 0.98 }}
          transition={{ duration: 0.25 }}
          onSubmit={handleSubmit}
          className="bg-white rounded-3xl shadow-xl p-8 border border-slate-100 space-y-6"
        >
          <div>
            <label
              htmlFor="email"
              className="block text-sm font-medium text-slate-700 mb-2"
            >
              Email
            </label>
            <div className="relative">
              <EnvelopeSimple
                size={20}
                className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                placeholder="seu@email.com"
                required
                className={`w-full pl-12 pr-4 py-3 border ${errors.email ? "border-red-500" : "border-slate-200"} rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all`}
              />
            </div>
            {errors.email && (
              <p className="text-red-500 text-sm mt-1">{errors.email}</p>
            )}
          </div>

          <div>
            <label
              htmlFor="password"
              className="block text-sm font-medium text-slate-700 mb-2"
            >
              Senha
            </label>
            <div className="relative">
              <Lock
                size={20}
                className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                placeholder="Sua senha"
                required
                className={`w-full pl-12 pr-4 py-3 border ${errors.password ? "border-red-500" : "border-slate-200"} rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all`}
              />
            </div>
            {errors.password && (
              <p className="text-red-500 text-sm mt-1">{errors.password}</p>
            )}
          </div>

          <div className="text-center mt-6">
            <motion.button
              onClick={() => navigate("/cadastro")}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="text-sm text-pink-600 font-medium underline hover:text-pink-700 focus:outline-none transition-all"
            >
              Não tem uma conta? <span className="font-bold">Cadastre-se agora</span>
            </motion.button>
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className={`w-full bg-pink-600 text-white py-3 rounded-xl font-semibold hover:bg-pink-700 active:scale-95 transition-all flex items-center justify-center gap-2 ${isSubmitting ? "opacity-50 cursor-not-allowed" : ""}`}
          >
            {isSubmitting ? "Entrando..." : "Entrar"}
          </button>
        </motion.form>
      </motion.div>
    </div>
  );
}

export default LoginPage;