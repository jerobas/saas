import { motion } from "motion/react";
import { ChartBar, Database, Package, ShoppingCart } from "@phosphor-icons/react";

const nextSlices = [
  {
    icon: <ShoppingCart size={24} />,
    title: "Compras",
    description: "Próximo fluxo operacional: postar compras e criar lotes de entrada.",
  },
  {
    icon: <Package size={24} />,
    title: "Estoque",
    description: "Saldos e lotes já leem o backend V2; falta a tela de lançamento.",
  },
  {
    icon: <ChartBar size={24} />,
    title: "Relatórios",
    description: "Dashboard real depende das consultas de vendas, produção e inventário.",
  },
];

const DashboardPage = () => (
  <>
    <header className="border-b border-slate-200 bg-white">
      <div className="mx-auto max-w-7xl px-6 py-8">
        <h1 className="text-3xl font-bold text-slate-900">Painel</h1>
        <p className="mt-2 text-slate-600">
          O dashboard fake foi removido. Esta área volta quando existirem consultas V2 reais.
        </p>
      </div>
    </header>

    <main className="mx-auto max-w-7xl px-6 py-8">
      <motion.section
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="mb-8 rounded-2xl border border-slate-100 bg-white p-8 shadow-sm"
      >
        <div className="mb-4 inline-flex rounded-full bg-pink-50 p-3 text-pink-600">
          <Database size={28} />
        </div>
        <h2 className="text-2xl font-semibold text-slate-900">Fonte de verdade: V2 local</h2>
        <p className="mt-3 max-w-3xl text-slate-600">
          Settings, unidades, catálogo, embalagens, contrapartes, saldos e lotes já passam pelo
          SQLite local e handlers Wails V2. O painel precisa esperar as próximas consultas reais em
          vez de exibir números inventados.
        </p>
      </motion.section>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
        {nextSlices.map((slice) => (
          <motion.article
            key={slice.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm"
          >
            <div className="mb-4 inline-flex rounded-full bg-slate-100 p-3 text-slate-700">
              {slice.icon}
            </div>
            <h3 className="text-lg font-semibold text-slate-900">{slice.title}</h3>
            <p className="mt-2 text-sm text-slate-600">{slice.description}</p>
          </motion.article>
        ))}
      </div>
    </main>
  </>
);

export default DashboardPage;
