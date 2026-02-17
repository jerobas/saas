import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Plus, Trash, Package } from 'phosphor-react';
import { GetAllItems } from '../../wailsjs/go/service/ItemService';
import { CreateBatch, GetBatchesByItem, DeleteBatch } from '../../wailsjs/go/service/BatchService';

const BatchesPage = () => {
  const [batches, setBatches] = useState([]);
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState(null);
  const [selectedItemFilter, setSelectedItemFilter] = useState('all');
  const [newBatch, setNewBatch] = useState({
    itemId: '',
    quantity: '',
    totalPrice: '',
  });

  useEffect(() => {
    loadData();
  }, []);

  useEffect(() => {
    if (selectedItemFilter !== 'all') {
      loadBatchesByItem(selectedItemFilter);
    } else {
      loadAllBatches();
    }
  }, [selectedItemFilter]);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const itemsData = await GetAllItems();
      setItems(itemsData || []);
      await loadAllBatches();
    } catch (err) {
      console.error('Erro ao carregar dados:', err);
      setError('Erro ao carregar dados. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  const loadAllBatches = async () => {
    try {
      const allBatches = [];
      for (const item of items) {
        const itemBatches = await GetBatchesByItem(item.id);
        if (itemBatches && itemBatches.length > 0) {
          allBatches.push(...itemBatches);
        }
      }
      setBatches(allBatches);
    } catch (err) {
      console.error('Erro ao carregar lotes:', err);
    }
  };

  const loadBatchesByItem = async (itemId) => {
    try {
      setLoading(true);
      const itemBatches = await GetBatchesByItem(itemId);
      setBatches(itemBatches || []);
    } catch (err) {
      console.error('Erro ao carregar lotes:', err);
      setError('Erro ao carregar lotes. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (value) => {
    if (!value) return '';
    const numericValue = value.replace(/\D/g, '');
    if (!numericValue) return '';
    const numberValue = parseInt(numericValue, 10) / 100;
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(numberValue);
  };

  const unformatCurrency = (value) => {
    const cleaned = value.replace(/\D/g, '');
    return parseFloat(cleaned) / 100;
  };

  const handleAddBatch = async () => {
    if (isSubmitting) return;

    if (newBatch.itemId && newBatch.quantity && newBatch.totalPrice) {
      setIsSubmitting(true);
      setError(null);

      try {
        const quantity = parseFloat(newBatch.quantity);
        const totalPrice = unformatCurrency(newBatch.totalPrice);

        await CreateBatch(newBatch.itemId, quantity, totalPrice);

        setNewBatch({ itemId: '', quantity: '', totalPrice: '' });
        setOpenDialog(false);
        
        if (selectedItemFilter !== 'all') {
          await loadBatchesByItem(selectedItemFilter);
        } else {
          await loadData();
        }
      } catch (err) {
        console.error('Erro ao adicionar lote:', err);
        setError('Erro ao adicionar lote. Tente novamente.');
      } finally {
        setIsSubmitting(false);
      }
    }
  };

  const handleDeleteBatch = async (id) => {
    if (loading) return;
    setDeleteConfirmId(id);
  };

  const confirmDelete = async () => {
    if (deleteConfirmId && !loading) {
      try {
        setError(null);
        await DeleteBatch(deleteConfirmId);
        setDeleteConfirmId(null);
        
        if (selectedItemFilter !== 'all') {
          await loadBatchesByItem(selectedItemFilter);
        } else {
          await loadData();
        }
      } catch (err) {
        console.error('Erro ao deletar lote:', err);
        setError('Erro ao deletar lote. Tente novamente.');
      }
    }
  };

  const getItemName = (itemId) => {
    const item = items.find((i) => i.id === itemId);
    return item ? item.name : 'Desconhecido';
  };

  const getItemUnit = (itemId) => {
    const item = items.find((i) => i.id === itemId);
    return item ? item.unit : '';
  };

  const totalBatchesValue = batches.reduce((acc, batch) => {
    return acc + (batch.unit_price * batch.quantity_remaining);
  }, 0);

  const totalQuantity = batches.reduce((acc, batch) => {
    return acc + batch.quantity_remaining;
  }, 0);

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
                {isSubmitting ? 'Salvando lote...' : 'Carregando lotes...'}
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
              Deletar Lote?
            </h2>

            <p className="text-slate-600 text-center mb-6">
              Tem certeza que deseja remover este lote? Esta ação não pode ser desfeita.
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
              <h1 className="text-3xl font-bold text-slate-900">Lotes de Estoque</h1>
              <p className="text-slate-600 mt-2">Gerencie os lotes de compra dos ingredientes</p>
            </div>
            <button
              onClick={() => setOpenDialog(true)}
              disabled={loading || isSubmitting || items.length === 0}
              className="flex items-center gap-2 bg-pink-600 text-white px-6 py-3 rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Novo Lote
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

        {/* Filter */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm mb-6"
        >
          <label className="block text-sm font-medium text-slate-700 mb-2">
            Filtrar por Ingrediente
          </label>
          <select
            value={selectedItemFilter}
            onChange={(e) => setSelectedItemFilter(e.target.value)}
            className="w-full md:w-96 px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
            disabled={loading}
          >
            <option value="all">Todos os ingredientes</option>
            {items.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </motion.div>

        {/* Summary Cards */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8"
        >
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <div className="flex items-center gap-3 mb-2">
              <Package size={24} className="text-pink-600" />
              <h3 className="text-slate-600 text-sm font-medium">Total de Lotes</h3>
            </div>
            <p className="text-3xl font-bold text-slate-900">{batches.length}</p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Valor Total</h3>
            <p className="text-3xl font-bold text-green-600 mt-2">
              {new Intl.NumberFormat('pt-BR', {
                style: 'currency',
                currency: 'BRL',
              }).format(totalBatchesValue)}
            </p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Quantidade Total Restante</h3>
            <p className="text-3xl font-bold text-blue-600 mt-2">
              {totalQuantity.toFixed(3)}
            </p>
          </div>
        </motion.div>

        {/* Batches Table */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden"
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-50 border-b border-slate-200">
                <tr>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Ingrediente</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Qtd. Total</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Qtd. Restante</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Preço Total</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Preço Unitário</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Data Compra</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Ação</th>
                </tr>
              </thead>
              <tbody>
                {batches.length === 0 ? (
                  <tr>
                    <td colSpan="7" className="px-6 py-12 text-center text-slate-500">
                      {items.length === 0
                        ? 'Cadastre ingredientes primeiro para criar lotes.'
                        : 'Nenhum lote cadastrado. Clique em "Novo Lote" para começar.'}
                    </td>
                  </tr>
                ) : (
                  batches.map((batch) => (
                    <tr key={batch.id} className="border-b border-slate-100 hover:bg-slate-50">
                      <td className="px-6 py-4 text-sm text-slate-900 font-medium">
                        {getItemName(batch.item_id)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {parseFloat(batch.quantity_total).toFixed(3)} {getItemUnit(batch.item_id)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {parseFloat(batch.quantity_remaining).toFixed(3)} {getItemUnit(batch.item_id)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-900 font-semibold">
                        {new Intl.NumberFormat('pt-BR', {
                          style: 'currency',
                          currency: 'BRL',
                        }).format(batch.purchase_price_total)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {new Intl.NumberFormat('pt-BR', {
                          style: 'currency',
                          currency: 'BRL',
                        }).format(batch.unit_price)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {new Date(batch.purchased_at).toLocaleDateString('pt-BR', {
                          day: '2-digit',
                          month: '2-digit',
                          year: 'numeric',
                        })}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <button
                          onClick={() => handleDeleteBatch(batch.id)}
                          className="text-red-600 hover:text-red-700 disabled:text-slate-300 disabled:cursor-not-allowed transition-colors"
                          disabled={loading || isSubmitting}
                          title="Deletar lote"
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

        {/* Add Batch Dialog */}
        {openDialog && (
          <div className="fixed inset-0 flex items-center justify-center z-50 backdrop-blur-md">
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="bg-white rounded-2xl p-8 max-w-md w-full shadow-xl"
            >
              <h2 className="text-2xl font-bold text-slate-900 mb-6">Adicionar Lote</h2>

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
                  <label className="block text-sm font-medium text-slate-700 mb-2">
                    Ingrediente
                  </label>
                  <select
                    value={newBatch.itemId}
                    onChange={(e) => setNewBatch({ ...newBatch, itemId: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    disabled={isSubmitting}
                  >
                    <option value="">Selecione um ingrediente</option>
                    {items.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name} ({item.unit})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">
                    Quantidade
                  </label>
                  <input
                    type="number"
                    step="0.001"
                    value={newBatch.quantity}
                    onChange={(e) => setNewBatch({ ...newBatch, quantity: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="0.000"
                    disabled={isSubmitting}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">
                    Preço Total (R$)
                  </label>
                  <input
                    type="text"
                    value={formatCurrency(newBatch.totalPrice)}
                    onChange={(e) => {
                      const inputValue = e.target.value;
                      const numericOnly = inputValue.replace(/\D/g, '');
                      setNewBatch({ ...newBatch, totalPrice: numericOnly });
                    }}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="R$ 0,00"
                    disabled={isSubmitting}
                  />
                </div>

                {newBatch.quantity && newBatch.totalPrice && (
                  <div className="p-4 bg-slate-50 rounded-lg">
                    <p className="text-sm text-slate-600">
                      Preço unitário:{' '}
                      <span className="font-semibold text-slate-900">
                        {new Intl.NumberFormat('pt-BR', {
                          style: 'currency',
                          currency: 'BRL',
                        }).format(unformatCurrency(newBatch.totalPrice) / parseFloat(newBatch.quantity))}
                      </span>
                    </p>
                  </div>
                )}
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
                  onClick={handleAddBatch}
                  disabled={
                    isSubmitting ||
                    loading ||
                    !newBatch.itemId ||
                    !newBatch.quantity ||
                    !newBatch.totalPrice
                  }
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

export default BatchesPage;