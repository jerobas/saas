import { useState } from "react";
import { Export, Import } from "../../wailsjs/go/service/DatabaseService";

const DatabasePage = () => {
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(false);

  const handleExport = async () => {
    try {
      setLoading(true);
      setStatus("Exportando...");
      await Export();
      setStatus("Exportacao concluida.");
    } catch (error) {
      console.error(error);
      const message = `${error?.message || ""}`.toLowerCase();
      setStatus(message.includes("cancel") ? "Operacao cancelada." : "Erro ao exportar.");
    } finally {
      setLoading(false);
    }
  };

  const handleImport = async () => {
    try {
      setLoading(true);
      setStatus("Importando...");
      await Import();
      setStatus("Importacao concluida.");
    } catch (error) {
      console.error(error);
      const message = `${error?.message || ""}`.toLowerCase();
      setStatus(message.includes("cancel") ? "Operacao cancelada." : "Erro ao importar.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-900">Backup da base</h1>
        <p className="text-slate-500">
          Exporte ou importe a base usando os dialogs nativos do sistema.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-slate-900">Exportar</h2>
          <p className="mt-1 text-sm text-slate-500">
            Salve uma copia da base atual em um arquivo.
          </p>

          <button
            onClick={handleExport}
            disabled={loading}
            className="mt-4 rounded-lg bg-pink-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-pink-700 disabled:opacity-60"
          >
            Exportar
          </button>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-slate-900">Importar</h2>
          <p className="mt-1 text-sm text-slate-500">
            Restaure a base a partir de um arquivo exportado.
          </p>

          <button
            onClick={handleImport}
            disabled={loading}
            className="mt-4 rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:opacity-60"
          >
            Importar
          </button>
        </div>
      </div>

      {status && (
        <div className="mt-6 rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700">
          {status}
        </div>
      )}
    </div>
  );
};

export default DatabasePage;
