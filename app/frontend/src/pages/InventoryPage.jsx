/* eslint-disable no-unused-vars */
import { useState } from 'react';
import { motion } from 'framer-motion';
import { Plus, Trash } from 'phosphor-react';

const InventoryPage = () => {
  const ingredients = [
    { id: 1, name: 'Farinha', purchaseQuantity: 10, unit: 'kg', purchasePrice: '5.00', minStock: 2 },
    { id: 2, name: 'Açúcar', purchaseQuantity: 5, unit: 'kg', purchasePrice: '3.50', minStock: 1 },
    { id: 3, name: 'Ovos', purchaseQuantity: 30, unit: 'un', purchasePrice: '0.50', minStock: 10 },
  ];

  const addIngredient = (ingredient) => {
    console.log('Adding ingredient:', ingredient);
  };

  const deleteIngredient = (id) => {
    console.log('Deleting ingredient with id:', id);
  };

  const loading = false;
  const error = null;

  const [openDialog, setOpenDialog] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState(null);
  const [newItem, setNewItem] = useState({
    name: '',
    purchaseQuantity: '',
    unit: 'kg',
    purchasePrice: '',
    minStock: '0',
  });

  /**
   * Formata valor em moeda (BRL)
   * @param {string} value - Valor não formatado
   * @returns {string} Valor formatado ex: R$ 1.234,56
   */
  const formatCurrency = (value) => {
    if (!value) return '';
    
    // Remove caracteres não numéricos
    const numericValue = value.replace(/\D/g, '');
    
    if (!numericValue) return '';
    
    // Converte para número e formata
    const numberValue = parseInt(numericValue, 10) / 100;
    
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(numberValue);
  };

  /**
   * Remove formatação de moeda
   * @param {string} value - Valor formatado
   * @returns {string} Valor limpo
   */
  const unformatCurrency = (value) => {
    return value.replace(/\D/g, '');
  };

  const handleAddItem = async () => {
    if (isSubmitting) return; // Previne múltiplos cliques
    
    if (newItem.name && newItem.purchaseQuantity && newItem.purchasePrice) {
      setIsSubmitting(true);
      
      // Limpar formatação da moeda para enviar
      const cleanPrice = unformatCurrency(newItem.purchasePrice);
      const priceValue = parseFloat(cleanPrice) / 100 || parseFloat(newItem.purchasePrice);
      
      const success = await addIngredient({
        name: newItem.name,
        purchaseQuantity: parseFloat(newItem.purchaseQuantity),
        unit: newItem.unit,
        purchasePrice: priceValue,
        minStock: parseFloat(newItem.minStock) || 0,
        currentStock: parseFloat(newItem.purchaseQuantity)
      });

      if (success) {
        setNewItem({ name: '', purchaseQuantity: '', unit: 'kg', purchasePrice: '', minStock: '0' });
        setOpenDialog(false);
      }
      
      setIsSubmitting(false);
    }
  };

  const handleDeleteItem = async (id) => {
    if (loading) return; // Previne múltiplos cliques
    setDeleteConfirmId(id);
  };

  const confirmDelete = async () => {
    if (deleteConfirmId && !loading) {
      await deleteIngredient(deleteConfirmId);
      setDeleteConfirmId(null);
    }
  };

  const totalInventoryValue = ingredients.reduce((acc, item) => {
    return acc + parseFloat(item.purchasePrice) * item.purchaseQuantity;
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
        {/* Summary Cards */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8"
        >
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Total de Ingredientes</h3>
            <p className="text-3xl font-bold text-slate-900 mt-2">{ingredients.length}</p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Valor Total em Estoque</h3>
            <p className="text-3xl font-bold text-green-600 mt-2">
              R$ {totalInventoryValue.toFixed(2)}
            </p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Custo Médio</h3>
            <p className="text-3xl font-bold text-blue-600 mt-2">
              R$ {ingredients.length > 0 ? (totalInventoryValue / ingredients.length).toFixed(2) : '0.00'}
            </p>
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
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Preço Compra</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Custo Total</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Estoque Mínimo</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Ação</th>
                </tr>
              </thead>
              <tbody>
                {ingredients.map((item) => {
                  const totalCost = parseFloat(item.purchasePrice) * parseFloat(item.currentStock);
                  return (
                    <tr key={item.id} className="border-b border-slate-100 hover:bg-slate-50">
                      <td className="px-6 py-4 text-sm text-slate-900">{item.name}</td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {parseFloat(item.currentStock).toFixed(3)} {item.unit}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        R$ {parseFloat(item.purchasePrice).toFixed(2)}
                      </td>
                      <td className="px-6 py-4 text-sm font-semibold text-slate-900">
                        R$ {totalCost.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {parseFloat(item.minStock).toFixed(3)} {item.unit}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <button
                          onClick={() => handleDeleteItem(item.id)}
                          className="text-red-600 hover:text-red-700 disabled:text-slate-300 disabled:cursor-not-allowed transition-colors"
                          disabled={loading || isSubmitting}
                          title={loading || isSubmitting ? 'Operação em progresso' : 'Deletar ingrediente'}
                        >
                          <Trash size={18} />
                        </button>
                      </td>
                    </tr>
                  );
                })}
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

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">Quantidade Compra</label>
                    <input
                      type="number"
                      step="0.1"
                      value={newItem.purchaseQuantity}
                      onChange={(e) => setNewItem({ ...newItem, purchaseQuantity: e.target.value })}
                      className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                      placeholder="0"
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
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Preço de Compra (R$)</label>
                  <input
                    type="text"
                    value={formatCurrency(newItem.purchasePrice)}
                    onChange={(e) => {
                      const inputValue = e.target.value;
                      // Apenas aceita números
                      const numericOnly = inputValue.replace(/\D/g, '');
                      setNewItem({ ...newItem, purchasePrice: numericOnly });
                    }}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="R$ 0,00"
                    disabled={isSubmitting}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Estoque Mínimo</label>
                  <input
                    type="number"
                    step="0.1"
                    value={newItem.minStock}
                    onChange={(e) => setNewItem({ ...newItem, minStock: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="0"
                    disabled={isSubmitting}
                  />
                </div>
              </div>

              <div className="flex gap-4">
                <button
                  onClick={() => setOpenDialog(false)}
                  className="flex-1 px-4 py-2 border border-slate-300 rounded-lg text-slate-700 hover:bg-slate-50 transition-all disabled:bg-slate-50 disabled:cursor-not-allowed"
                  disabled={isSubmitting}
                >
                  Cancelar
                </button>
                <button
                  onClick={handleAddItem}
                  disabled={isSubmitting || loading || !newItem.name.trim() || !newItem.purchaseQuantity || !newItem.purchasePrice}
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
