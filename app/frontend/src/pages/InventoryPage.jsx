/* eslint-disable no-unused-vars */
import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Plus, Trash, Warning } from 'phosphor-react';
import { CreateItem, GetAllItems, DeleteItem } from '../../wailsjs/go/service/ItemService';
import { GetBatchesByItem } from '../../wailsjs/go/service/BatchService';

const InventoryPage = () => {
  const [ingredients, setIngredients] = useState([]);
  const [ingredientsWithStock, setIngredientsWithStock] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState(null);
  const [newItem, setNewItem] = useState({
    name: '',
    unit: 'kg',
    minStock: '0',
  });

  useEffect(() => {
    loadIngredients();
  }, []);

  const loadIngredients = async () => {
    try {
      setLoading(true);
      setError(null);
      const items = await GetAllItems();
      setIngredients(items || []);

      // Buscar estoque de cada item
      const itemsWithStock = await Promise.all(
        (items || []).map(async (item) => {
          const batches = await GetBatchesByItem(item.id);
          const totalStock = batches.reduce((acc, batch) => acc + batch.quantity_remaining, 0);
          const avgPrice = batches.length > 0
            ? batches.reduce((acc, batch) => acc + batch.unit_price, 0) / batches.length
            : 0;
          
          return {
            ...item,
            currentStock: totalStock,
            avgPrice: avgPrice,
            totalValue: totalStock * avgPrice,
            isLowStock: totalStock <= item.min_stock_alert,
          };
        })
      );

      setIngredientsWithStock(itemsWithStock);
    } catch (err) {
      console.error('Erro ao carregar ingredientes:', err);
      setError('Erro ao carregar ingredientes. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  const addIngredient = async (ingredient) => {
    try {
      setError(null);

      const item = await CreateItem(
        ingredient.name,
        ingredient.unit,
        ingredient.minStock
      );

      if (!item) {
        throw new Error('Falha ao criar item');
      }

      await loadIngredients();
      return true;
    } catch (err) {
      console.error('Erro ao adicionar ingrediente:', err);
      setError(err.message || 'Erro ao adicionar ingrediente. Tente novamente.');
      return false;
    }
  };

  const deleteIngredient = async (id) => {
    try {
      setError(null);
      await DeleteItem(id);
      await loadIngredients();
      return true;
    } catch (err) {
      console.error('Erro ao deletar ingrediente:', err);
      setError('Erro ao deletar ingrediente. Tente novamente.');
      return false;
    }
  };

  const handleAddItem = async () => {
    if (isSubmitting) return;

    if (newItem.name) {
      setIsSubmitting(true);

      const success = await addIngredient({
        name: newItem.name,
        unit: newItem.unit,
        minStock: parseFloat(newItem.minStock) || 0,
      });

      if (success) {
        setNewItem({ name: '', unit: 'kg', minStock: '0' });
        setOpenDialog(false);
      }

      setIsSubmitting(false);
    }
  };

  const handleDeleteItem = async (id) => {
    if (loading) return;
    setDeleteConfirmId(id);
  };

  const confirmDelete = async () => {
    if (deleteConfirmId && !loading) {
      await deleteIngredient(deleteConfirmId);
      setDeleteConfirmId(null);
    }
  };

  const totalInventoryValue = ingredientsWithStock.reduce((acc, item) => acc + item.totalValue, 0);
  const lowStockItems = ingredientsWithStock.filter((item) => item.isLowStock).length;

  return (
    <>
      {/* Loading Overlay */}
      {(loading || isSubmitting) && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="fixed inset-0 flex items-center justify-center z-[60] backdrop-blur-md"
        >
          <motion.div
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="bg-white rounded-2xl p-8 flex flex-col items-center gap-4"
          >
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
              className="w-12 h-12 border-4 border-pink-200 border-t-pink-600 rounded-full"
            />
            <div className="text-center">
              <h3 className="text-lg font-semibold text-slate-900">
                {isSubmitting ? 'Salvando ingrediente...' : 'Carregando estoque...'}
              </h3>
              <p className="text-sm text-slate-600 mt-1">Por favor aguarde</p>
            </div>
          </motion.div>
        </motion.div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirmId && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="fixed inset-0 flex items-center justify-center z-50 backdrop-blur-md"
        >
          <motion.div
            initial={{ scale: 0.9, opacity: 0, y: 20 }}
            animate={{ scale: 1, opacity: 1, y: 0 }}
            className="bg-white rounded-2xl p-8 max-w-md w-full shadow-xl"
          >
            <div className="flex items-center justify-center w-12 h-12 bg-red-100 rounded-full mx-auto mb-6">
              <Trash size={24} className="text-red-600" />
            </div>

            <h2 className="text-2xl font-bold text-slate-900 mb-2 text-center">
              Deletar Ingrediente?
            </h2>

            <p className="text-slate-600 text-center mb-6">
              Tem certeza que deseja remover este ingrediente do estoque? Esta ação não pode ser desfeita.
            </p>

            <div className="flex gap-4">
              <button
                onClick={() => setDeleteConfirmId(null)}
                disabled={loading}
                className="flex-1 px-4 py-3 border border-slate-300 rounded-lg text-slate-700 hover:bg-slate-50 transition-all disabled:bg-slate-50 disabled:cursor-not-allowed font-medium"
              >
                Cancelar
              </button>
              <button
                onClick={confirmDelete}
                disabled={loading}
                className="flex-1 px-4 py-3 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed font-medium flex items-center justify-center gap-2"
              >
                {loading ? (
                  <>
                    <motion.div
                      animate={{ rotate: 360 }}
                      transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
                      className="w-4 h-4 border-2 border-white border-t-transparent rounded-full"
                    />
                    Deletando...
                  </>
                ) : (
                  'Sim, Deletar'
                )}
              </button>
            </div>
          </motion.div>
        </motion.div>
      )}

      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-slate-900">Ingredientes</h1>
              <p className="text-slate-600 mt-2">Gerencie seu estoque de ingredientes</p>
            </div>
            <button
              onClick={() => setOpenDialog(true)}
              disabled={loading || isSubmitting}
              className="flex items-center gap-2 bg-pink-600 text-white px-6 py-3 rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Novo Ingrediente
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        {/* Error Message */}
        {error && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm mb-6"
          >
            <span>{error}</span>
            <button
              onClick={() => setError(null)}
              className="ml-auto text-red-600 hover:text-red-800"
            >
              ✕
            </button>
          </motion.div>
        )}

        {/* Summary Cards */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8"
        >
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Total de Ingredientes</h3>
            <p className="text-3xl font-bold text-slate-900 mt-2">{ingredientsWithStock.length}</p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Valor Total em Estoque</h3>
            <p className="text-3xl font-bold text-green-600 mt-2">
              {new Intl.NumberFormat('pt-BR', {
                style: 'currency',
                currency: 'BRL',
              }).format(totalInventoryValue)}
            </p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Custo Médio</h3>
            <p className="text-3xl font-bold text-blue-600 mt-2">
              {new Intl.NumberFormat('pt-BR', {
                style: 'currency',
                currency: 'BRL',
              }).format(
                ingredientsWithStock.length > 0 ? totalInventoryValue / ingredientsWithStock.length : 0
              )}
            </p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <div className="flex items-center gap-2">
              <Warning size={20} className="text-orange-600" />
              <h3 className="text-slate-600 text-sm font-medium">Estoque Baixo</h3>
            </div>
            <p className="text-3xl font-bold text-orange-600 mt-2">{lowStockItems}</p>
          </div>
        </motion.div>

        {/* Items Table */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden"
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-50 border-b border-slate-200">
                <tr>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Nome</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Estoque Atual</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Preço Médio</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Valor Total</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Estoque Mínimo</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Status</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Ação</th>
                </tr>
              </thead>
              <tbody>
                {ingredientsWithStock.length === 0 ? (
                  <tr>
                    <td colSpan="7" className="px-6 py-12 text-center text-slate-500">
                      Nenhum ingrediente cadastrado. Clique em "Novo Ingrediente" para começar.
                    </td>
                  </tr>
                ) : (
                  ingredientsWithStock.map((item) => (
                    <tr
                      key={item.id}
                      className={`border-b border-slate-100 hover:bg-slate-50 ${
                        item.isLowStock ? 'bg-orange-50' : ''
                      }`}
                    >
                      <td className="px-6 py-4 text-sm text-slate-900 font-medium">{item.name}</td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {item.currentStock.toFixed(3)} {item.unit}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {new Intl.NumberFormat('pt-BR', {
                          style: 'currency',
                          currency: 'BRL',
                        }).format(item.avgPrice)}
                      </td>
                      <td className="px-6 py-4 text-sm font-semibold text-slate-900">
                        {new Intl.NumberFormat('pt-BR', {
                          style: 'currency',
                          currency: 'BRL',
                        }).format(item.totalValue)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {parseFloat(item.min_stock_alert).toFixed(3)} {item.unit}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        {item.isLowStock ? (
                          <span className="inline-flex items-center gap-1 px-2 py-1 bg-orange-100 text-orange-700 rounded-full text-xs font-medium">
                            <Warning size={14} />
                            Baixo
                          </span>
                        ) : (
                          <span className="inline-flex items-center px-2 py-1 bg-green-100 text-green-700 rounded-full text-xs font-medium">
                            OK
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <button
                          onClick={() => handleDeleteItem(item.id)}
                          className="text-red-600 hover:text-red-700 disabled:text-slate-300 disabled:cursor-not-allowed transition-colors"
                          disabled={loading || isSubmitting}
                          title="Deletar ingrediente"
                        >
                          <Trash size={18} />
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </motion.div>

        {/* Add Item Dialog */}
        {openDialog && (
          <div className="fixed inset-0 flex items-center justify-center z-50 backdrop-blur-md">
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="bg-white rounded-2xl p-8 max-w-md w-full shadow-xl"
            >
              <h2 className="text-2xl font-bold text-slate-900 mb-6">Adicionar Ingrediente</h2>

              {error && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm mb-6"
                >
                  <span>{error}</span>
                </motion.div>
              )}

              <div className="space-y-4 mb-6">
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Nome</label>
                  <input
                    type="text"
                    value={newItem.name}
                    onChange={(e) => setNewItem({ ...newItem, name: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="Ex: Farinha de Trigo"
                    disabled={isSubmitting}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Unidade</label>
                  <select
                    value={newItem.unit}
                    onChange={(e) => setNewItem({ ...newItem, unit: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    disabled={isSubmitting}
                  >
                    <option value="kg">Quilograma (kg)</option>
                    <option value="g">Grama (g)</option>
                    <option value="l">Litro (l)</option>
                    <option value="ml">Mililitro (ml)</option>
                    <option value="dz">Dúzia (dz)</option>
                    <option value="un">Unidade (un)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">
                    Estoque Mínimo
                  </label>
                  <input
                    type="number"
                    step="0.001"
                    value={newItem.minStock}
                    onChange={(e) => setNewItem({ ...newItem, minStock: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="0"
                    disabled={isSubmitting}
                  />
                </div>

                <div className="p-4 bg-blue-50 rounded-lg">
                  <p className="text-sm text-blue-800">
                    💡 <strong>Dica:</strong> Após criar o ingrediente, vá para a página de{' '}
                    <strong>Lotes</strong> para adicionar o estoque inicial.
                  </p>
                </div>
              </div>

              <div className="flex gap-4">
                <button
                  onClick={() => {
                    setOpenDialog(false);
                    setError(null);
                  }}
                  className="flex-1 px-4 py-2 border border-slate-300 rounded-lg text-slate-700 hover:bg-slate-50 transition-all disabled:bg-slate-50 disabled:cursor-not-allowed"
                  disabled={isSubmitting}
                >
                  Cancelar
                </button>
                <button
                  onClick={handleAddItem}
                  disabled={isSubmitting || loading || !newItem.name.trim()}
                  className="flex-1 px-4 py-2 bg-pink-600 text-white rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {isSubmitting || loading ? (
                    <>
                      <motion.div
                        animate={{ rotate: 360 }}
                        transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
                        className="w-4 h-4 border-2 border-white border-t-transparent rounded-full"
                      />
                      Salvando...
                    </>
                  ) : (
                    'Adicionar'
                  )}
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </main>
    </>
  );
};

export default InventoryPage;