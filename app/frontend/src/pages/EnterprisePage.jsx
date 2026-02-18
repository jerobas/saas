import React, { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { CurrencyDollar, Percent, Target, CreditCard } from "phosphor-react";
// import { ProfileService } from '../services/profileService';

const EnterprisePage = () => {
  const [profile, setProfile] = useState({
    id: 1,
    hourly_cost: 0,
    default_profit_margin: 30,
    expected_monthly_profit: null,
    fixed_monthly_expenses: null,
  });

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState({ type: "", text: "" });

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    try {
      setLoading(true);
      // const data = await ProfileService.getProfile();
      setProfile(data);
      setMessage({ type: "", text: "" });
    } catch (error) {
      console.error("Erro ao carregar perfil:", error);
      setMessage({ type: "error", text: "Erro ao carregar perfil" });
      setProfile({
        id: 1,
        hourly_cost: 0,
        default_profit_margin: 30,
        expected_monthly_profit: null,
        fixed_monthly_expenses: null,
      });
    } finally {
      setLoading(false);
    }
  };

  const formatBRL = (value) => {
    if (value === "" || value === null) return "";
    const numValue = parseFloat(value.toString().replace(/\D/g, "")) / 100;
    if (isNaN(numValue)) return "";
    return numValue.toLocaleString("pt-BR", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  };

  const parseBRL = (value) => {
    if (value === "" || value === null) return null;
    const numValue = parseFloat(value.toString().replace(/\D/g, "")) / 100;
    return isNaN(numValue) || numValue < 0 ? 0 : numValue;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    const monetaryFields = [
      "hourly_cost",
      "expected_monthly_profit",
      "fixed_monthly_expenses",
    ];

    let finalValue;
    if (monetaryFields.includes(name)) {
      finalValue = parseBRL(value);
    } else {
      const numValue = value === "" ? null : parseFloat(value);
      finalValue = numValue !== null && numValue < 0 ? 0 : numValue;
    }

    setProfile((prev) => ({
      ...prev,
      [name]: finalValue,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (profile.hourly_cost < 0 || profile.default_profit_margin < 0) {
      setMessage({ type: "error", text: "Os valores não podem ser negativos" });
      return;
    }

    try {
      setSaving(true);
      // const updated = await ProfileService.updateProfile({
      //   hourly_cost: profile.hourly_cost,
      //   default_profit_margin: profile.default_profit_margin,
      //   expected_monthly_profit: profile.expected_monthly_profit,
      //   fixed_monthly_expenses: profile.fixed_monthly_expenses,
      // });

      setProfile(updated);
      setMessage({ type: "success", text: "Perfil atualizado com sucesso!" });

      setTimeout(() => {
        setMessage({ type: "", text: "" });
      }, 3000);
    } catch (error) {
      console.error("Erro ao salvar perfil:", error);
      setMessage({ type: "error", text: "Erro ao salvar perfil" });
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-pink-50 via-white to-purple-50 flex items-center justify-center">
        <motion.div
          animate={{ opacity: [0.5, 1, 0.5] }}
          transition={{ duration: 2, repeat: Infinity }}
          className="text-slate-600 font-semibold text-lg"
        >
          Carregando...
        </motion.div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-pink-50 via-white to-purple-50 p-4 py-12">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="max-w-2xl mx-auto"
      >
        <div className="text-center mb-8">
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
            className="inline-flex items-center justify-center w-16 h-16 bg-pink-600 rounded-2xl mb-4"
          >
            <CurrencyDollar size={32} weight="bold" className="text-white" />
          </motion.div>
          <h1 className="text-3xl font-bold text-slate-900 mb-2">
            Configurações da Empresa
          </h1>
          <p className="text-slate-600">
            Gerencie suas configurações de custos e projeções financeiras
          </p>
        </div>

        {message.text && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className={`mb-6 p-4 rounded-xl font-medium ${
              message.type === "success"
                ? "bg-green-50 border border-green-200 text-green-700"
                : "bg-red-50 border border-red-200 text-red-700"
            }`}
          >
            {message.text}
          </motion.div>
        )}

        <motion.form
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.1 }}
          onSubmit={handleSubmit}
          className="space-y-6"
        >
          {/* Seção Custos e Margens */}
          <div className="bg-white rounded-3xl shadow-xl p-8 border border-slate-100">
            <div className="flex items-center gap-3 mb-6">
              <div className="p-3 bg-pink-100 rounded-lg">
                <Percent size={24} className="text-pink-600" weight="bold" />
              </div>
              <h2 className="text-xl font-bold text-slate-900">
                Custos e Margens
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label
                  htmlFor="hourly_cost"
                  className="block text-sm font-semibold text-slate-700 mb-2"
                >
                  Custo Horário (R$)
                </label>
                <p className="text-xs text-slate-500 mb-3">
                  Qual é seu custo horário de trabalho?
                </p>
                <div className="relative">
                  <CurrencyDollar
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="text"
                    id="hourly_cost"
                    name="hourly_cost"
                    value={formatBRL(profile.hourly_cost)}
                    onChange={handleChange}
                    min="0"
                    placeholder="Ex: 50,00"
                    disabled={message.type === 'error'}
                    className="w-full pl-12 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all disabled:bg-slate-100 disabled:cursor-not-allowed"
                  />
                </div>
              </div>

              <div>
                <label
                  htmlFor="default_profit_margin"
                  className="block text-sm font-semibold text-slate-700 mb-2"
                >
                  Margem de Lucro Padrão (%)
                </label>
                <p className="text-xs text-slate-500 mb-3">
                  Qual é a margem de lucro padrão para as receitas?
                </p>
                <div className="relative">
                  <Percent
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="number"
                    id="default_profit_margin"
                    name="default_profit_margin"
                    value={profile.default_profit_margin || ""}
                    onChange={handleChange}
                    step="0.01"
                    min="0"
                    placeholder="Ex: 30"
                    disabled={message.type === 'error'}
                    className="w-full pl-12 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all disabled:bg-slate-100 disabled:cursor-not-allowed"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Seção Projeções Financeiras */}
          <div className="bg-white rounded-3xl shadow-xl p-8 border border-slate-100">
            <div className="flex items-center gap-3 mb-6">
              <div className="p-3 bg-purple-100 rounded-lg">
                <Target size={24} className="text-purple-600" weight="bold" />
              </div>
              <h2 className="text-xl font-bold text-slate-900">
                Projeções Financeiras
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label
                  htmlFor="expected_monthly_profit"
                  className="block text-sm font-semibold text-slate-700 mb-2"
                >
                  Lucro Mensal Esperado (R$)
                </label>
                <p className="text-xs text-slate-500 mb-3">
                  Qual é o lucro mensal que você espera ganhar?
                </p>
                <div className="relative">
                  <Target
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="text"
                    id="expected_monthly_profit"
                    name="expected_monthly_profit"
                    value={formatBRL(profile.expected_monthly_profit)}
                    onChange={handleChange}
                    min="0"
                    placeholder="Ex: 5.000,00"
                    disabled={message.type === 'error'}
                    className="w-full pl-12 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all disabled:bg-slate-100 disabled:cursor-not-allowed"
                  />
                </div>
              </div>

              <div>
                <label
                  htmlFor="fixed_monthly_expenses"
                  className="block text-sm font-semibold text-slate-700 mb-2"
                >
                  Despesas Mensais Fixas (R$)
                </label>
                <p className="text-xs text-slate-500 mb-3">
                  Qual é o total de despesas fixas mensais?
                </p>
                <div className="relative">
                  <CreditCard
                    size={20}
                    className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                  />
                  <input
                    type="text"
                    id="fixed_monthly_expenses"
                    name="fixed_monthly_expenses"
                    value={formatBRL(profile.fixed_monthly_expenses)}
                    onChange={handleChange}
                    min="0"
                    placeholder="Ex: 2.000,00"
                    disabled={message.type === 'error'}
                    className="w-full pl-12 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition-all disabled:bg-slate-100 disabled:cursor-not-allowed"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Botões de Ação */}
          <div className="flex gap-4">
            <button
              type="submit"
              disabled={saving || message.type === 'error'}
              className="flex-1 bg-pink-600 text-white py-3 rounded-xl font-semibold hover:bg-pink-700 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? "Salvando..." : "Salvar Configurações"}
            </button>
            <button
              type="button"
              onClick={loadProfile}
              disabled={saving}
              className="flex-1 bg-slate-200 text-slate-700 py-3 rounded-xl font-semibold hover:bg-slate-300 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Descartar Alterações
            </button>
          </div>

          {/* Última Atualização */}
          {profile.updated_at && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.3 }}
              className="text-center text-sm text-slate-500 pt-4 border-t border-slate-200"
            >
              Última atualização:{" "}
              {new Date(profile.updated_at).toLocaleString("pt-BR")}
            </motion.div>
          )}
        </motion.form>
      </motion.div>
    </div>
  );
};

export default EnterprisePage;
