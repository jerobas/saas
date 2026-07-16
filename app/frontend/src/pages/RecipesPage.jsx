import { motion } from "motion/react";
import { FileText } from "@phosphor-icons/react";

const RecipesPage = () => (
  <>
    <header className="border-b border-slate-200 bg-white">
      <div className="mx-auto max-w-7xl px-6 py-8">
        <h1 className="text-3xl font-bold text-slate-900">Receitas</h1>
        <p className="mt-2 text-slate-600">
          O fluxo V2 de receitas ainda não foi conectado à interface.
        </p>
      </div>
    </header>

    <main className="mx-auto max-w-7xl px-6 py-8">
      <motion.section
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="rounded-2xl border border-slate-100 bg-white p-8 text-center shadow-sm"
      >
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-pink-50 text-pink-600">
          <FileText size={28} />
        </div>
        <h2 className="text-xl font-semibold text-slate-900">Receitas entram depois de compras</h2>
        <p className="mx-auto mt-3 max-w-2xl text-slate-600">
          A modelagem V2 de receitas e revisões já existe no banco e nos stores, mas ainda falta a
          camada de aplicação/Wails/frontend. Em vez de chamar a implementação antiga, esta página
          fica parada até construirmos esse slice bottom-up.
        </p>
      </motion.section>
    </main>
  </>
);

export default RecipesPage;
