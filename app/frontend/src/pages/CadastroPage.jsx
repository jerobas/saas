import { useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
  User,
  EnvelopeSimple,
  IdentificationCard,
  Phone,
  Lock,
} from "phosphor-react";
import { createUser } from "../services/apiService";
import { SaveUserData } from "../../wailsjs/go/main/UserService";
import { useContext } from "react";
import { AppContext } from "../context/AppContext";
import { useNavigate } from "react-router-dom";

function CadastroPage() {
  const navigate = useNavigate();
  const { setPixData } = useContext(AppContext);

  const [step, setStep] = useState(1);
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    taxId: "",
    cellphone: "",
    password: "",
  });
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState({});

  const validateForm = () => {
    const newErrors = {};
    if (!formData.password.trim() || formData.password.length < 8)
      newErrors.password = "A senha deve ter no mínimo 8 caracteres";
    if (confirmPassword !== formData.password)
      newErrors.confirmPassword = "As senhas não conferem";
    if (!formData.name.trim()) newErrors.name = "O nome é obrigatório";
    if (
      !formData.email.trim() ||
      !/^[^\s@]+@[\w-]+\.[a-z]{2,}$/i.test(formData.email)
    )
      newErrors.email = "Email inválido";
    if (
      !formData.taxId.trim() ||
      !/^\d{3}\.\d{3}\.\d{3}-\d{2}$/.test(formData.taxId)
    )
      newErrors.taxId = "CPF inválido";
    if (
      !formData.cellphone.trim() ||
      !/^\(\d{2}\) \d{5}-\d{4}$/.test(formData.cellphone)
    )
      newErrors.cellphone = "Número de celular inválido";
    return newErrors;
  };

  const validatePasswordStep = () => {
    const newErrors = {};
    if (!formData.password.trim() || formData.password.length < 8)
      newErrors.password = "A senha deve ter no mínimo 8 caracteres";
    if (confirmPassword !== formData.password)
      newErrors.confirmPassword = "As senhas não conferem";
    return newErrors;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    let maskedValue = value;

    if (name === "taxId") {
      maskedValue = value
        .replace(/\D/g, "")
        .replace(/(\d{3})(\d)/, "$1.$2")
        .replace(/(\d{3})(\d)/, "$1.$2")
        .replace(/(\d{3})(\d{1,2})$/, "$1-$2");
    } else if (name === "cellphone") {
      maskedValue = value
        .replace(/\D/g, "")
        .replace(/(\d{2})(\d)/, "($1) $2")
        .replace(/(\d{5})(\d)/, "$1-$2")
        .replace(/(-\d{4})\d+?$/, "$1");
    }

    if (name === "confirmPassword") {
      setConfirmPassword(maskedValue);
      setErrors({ ...errors, confirmPassword: "" });
      return;
    }

    setFormData({ ...formData, [name]: maskedValue });
    setErrors({ ...errors, [name]: "" });
  };

  const handlePasswordSubmit = (e) => {
    e.preventDefault();
    const validationErrors = validatePasswordStep();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setStep(2);
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
      const { data } = await createUser(formData);

      await SaveUserData(data.userId, formData.email);

      setIsSubmitting(true);

      const sse = new EventSource(
        `http://localhost:3000/api/sse/${data.userId}`,
      );

      sse.onmessage = (event) => {
        setPixData({ data: JSON.parse(event.data), email: formData.email });
        navigate("/pix-payment");
        setIsSubmitting(false);
      };

      sse.onerror = (error) => {
        console.error("SSE error:", error);
        sse.close();
        setIsSubmitting(false);
      };
    } catch (error) {
      console.error("Erro ao enviar dados:", error);
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
            <User size={32} weight="bold" className="text-white" />
          </motion.div>
          <h1 className="text-3xl font-bold text-slate-900 mb-2">Cadastro</h1>
          <p className="text-slate-600">
            Preencha os campos abaixo para se cadastrar
          </p>
        </div>

        <AnimatePresence mode="wait">
          {step === 1 ? (
            <motion.form
              key="step-1"
              initial={{ opacity: 0, x: -16, scale: 0.98 }}
              animate={{ opacity: 1, x: 0, scale: 1 }}
              exit={{ opacity: 0, x: 16, scale: 0.98 }}
              transition={{ duration: 0.25 }}
              onSubmit={handlePasswordSubmit}
              className="bg-white rounded-3xl shadow-xl p-8 border border-slate-100 space-y-6"
            >
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
                    placeholder="Crie uma senha"
                    required
                    className={`w-full pl-12 pr-4 py-3 border ${errors.password ? "border-red-500" : "border-slate-200"} rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all`}
                  />
                </div>
                {errors.password && (
                  <p className="text-red-500 text-sm mt-1">{errors.password}</p>
                )}
              </div>

              <div>
                <label
                  htmlFor="confirmPassword"
                  className="block text-sm font-medium text-slate-700 mb-2"
                >
                  Confirmar senha
                </label>
                <div className="relative">
                  <Lock
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="password"
                    id="confirmPassword"
                    name="confirmPassword"
                    value={confirmPassword}
                    onChange={handleChange}
                    placeholder="Repita a senha"
                    required
                    className={`w-full pl-12 pr-4 py-3 border ${errors.confirmPassword ? "border-red-500" : "border-slate-200"} rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all`}
                  />
                </div>
                {errors.confirmPassword && (
                  <p className="text-red-500 text-sm mt-1">
                    {errors.confirmPassword}
                  </p>
                )}
              </div>

              <button
                type="submit"
                className="w-full bg-pink-600 text-white py-3 rounded-xl font-semibold hover:bg-pink-700 active:scale-95 transition-all flex items-center justify-center gap-2"
              >
                Continuar
              </button>
            </motion.form>
          ) : (
            <motion.form
              key="step-2"
              initial={{ opacity: 0, x: 16, scale: 0.98 }}
              animate={{ opacity: 1, x: 0, scale: 1 }}
              exit={{ opacity: 0, x: -16, scale: 0.98 }}
              transition={{ duration: 0.25 }}
              onSubmit={handleSubmit}
              className="bg-white rounded-3xl shadow-xl p-8 border border-slate-100 space-y-6"
            >
              <div>
                <label
                  htmlFor="name"
                  className="block text-sm font-medium text-slate-700 mb-2"
                >
                  Nome
                </label>
                <div className="relative">
                  <User
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="text"
                    id="name"
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    placeholder="Seu nome completo"
                    required
                    className={`w-full pl-12 pr-4 py-3 border ${errors.name ? "border-red-500" : "border-slate-200"} rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all`}
                  />
                </div>
                {errors.name && (
                  <p className="text-red-500 text-sm mt-1">{errors.name}</p>
                )}
              </div>

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
                  htmlFor="taxId"
                  className="block text-sm font-medium text-slate-700 mb-2"
                >
                  CPF ou CNPJ
                </label>
                <div className="relative">
                  <IdentificationCard
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="text"
                    id="taxId"
                    name="taxId"
                    value={formData.taxId}
                    onChange={handleChange}
                    placeholder="000.000.000-00"
                    required
                    className={`w-full pl-12 pr-4 py-3 border ${errors.taxId ? "border-red-500" : "border-slate-200"} rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all`}
                  />
                </div>
                {errors.taxId && (
                  <p className="text-red-500 text-sm mt-1">{errors.taxId}</p>
                )}
              </div>

              <div>
                <label
                  htmlFor="cellphone"
                  className="block text-sm font-medium text-slate-700 mb-2"
                >
                  Celular
                </label>
                <div className="relative">
                  <Phone
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="text"
                    id="cellphone"
                    name="cellphone"
                    value={formData.cellphone}
                    onChange={handleChange}
                    placeholder="(00) 00000-0000"
                    required
                    className={`w-full pl-12 pr-4 py-3 border ${errors.cellphone ? "border-red-500" : "border-slate-200"} rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all`}
                  />
                </div>
                {errors.cellphone && (
                  <p className="text-red-500 text-sm mt-1">
                    {errors.cellphone}
                  </p>
                )}
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className={`w-full bg-pink-600 text-white py-3 rounded-xl font-semibold hover:bg-pink-700 active:scale-95 transition-all flex items-center justify-center gap-2 ${isSubmitting ? "opacity-50 cursor-not-allowed" : ""}`}
              >
                {isSubmitting ? "Enviando..." : "Enviar"}
              </button>
            </motion.form>
          )}
        </AnimatePresence>

        <p className="text-center text-sm text-slate-500 mt-6">
          Ao se cadastrar, você concorda com nossos Termos de Serviço e Política
          de Privacidade
        </p>
      </motion.div>
    </div>
  );
}

export default CadastroPage;
