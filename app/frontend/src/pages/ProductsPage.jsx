/* eslint-disable no-unused-vars */
import { useState } from 'react';
import { motion } from 'framer-motion';
import { Plus, Trash } from 'phosphor-react';

const ProductsPage = () => {
  const products = [
    { id: 1, name: 'Bolo de Chocolate', description: 'Delicioso bolo de chocolate', basePrice: '20.00', markup: '50', isActive: true },
    { id: 2, name: 'Torta de Limão', description: 'Torta com recheio de limão', basePrice: '15.00', markup: '40', isActive: true },
    { id: 3, name: 'Pão de Mel', description: 'Pão de mel caseiro', basePrice: '5.00', markup: '30', isActive: false },
  ];

  const addProduct = (product) => {
    console.log('Adding product:', product);
  };

  const deleteProduct = (id) => {
    console.log('Deleting product with id:', id);
  };

  const loading = false;
  const error = null;

  const [openDialog, setOpenDialog] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState(null);
  const [newProduct, setNewProduct] = useState({
    name: '',
    description: '',
    basePrice: '',
    markup: '',
    isActive: true,
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

  /**
   * Calcula o preço final com markup
   * @param {number} basePrice - Preço base
   * @param {number} markup - Margem em porcentagem
   * @returns {number} Preço final
   */
  const calculateFinalPrice = (basePrice, markup) => {
    const base = parseFloat(basePrice) || 0;
    const markupValue = parseFloat(markup) || 0;
    return base + (base * (markupValue / 100));
  };

  const handleAddProduct = async () => {
    if (isSubmitting) return;
    
    if (newProduct.name && newProduct.basePrice) {
      setIsSubmitting(true);
      
      const cleanPrice = unformatCurrency(newProduct.basePrice);
      const priceValue = parseFloat(cleanPrice) / 100 || parseFloat(newProduct.basePrice);
      
      const success = await addProduct({
        name: newProduct.name,
        description: newProduct.description || null,
        basePrice: priceValue,
        markup: parseFloat(newProduct.markup) || 0,
        isActive: newProduct.isActive
      });

      if (success) {
        setNewProduct({ name: '', description: '', basePrice: '', markup: '', isActive: true });
        setOpenDialog(false);
      }
      
      setIsSubmitting(false);
    }
  };

  const handleDeleteProduct = async (id) => {
    if (loading) return;
    setDeleteConfirmId(id);
  };

  const confirmDelete = async () => {
    if (deleteConfirmId && !loading) {
      await deleteProduct(deleteConfirmId);
      setDeleteConfirmId(null);
    }
  };

  const totalProductValue = products.reduce((acc, product) => {
    return acc + parseFloat(product.basePrice) + (parseFloat(product.basePrice) * (parseFloat(product.markup) / 100));
  }, 0);

  const activeProducts = products.filter(p => p.isActive).length;

  return (
    <>
      {/* Loading Overlay */}
      {(loading || isSubmitting) && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="fixed inset-0 flex items-center justify-center z-60 backdrop-blur-md"
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
                {isSubmitting ? 'Salvando produto...' : 'Carregando produtos...'}
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
              Deletar Produto?
            </h2>

            <p className="text-slate-600 text-center mb-6">
              Tem certeza que deseja remover este produto? Esta ação não pode ser desfeita.
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
              <h1 className="text-3xl font-bold text-slate-900">Produtos</h1>
              <p className="text-slate-600 mt-2">Gerencie seus produtos e preços</p>
            </div>
            <button
              onClick={() => setOpenDialog(true)}
              disabled={loading || isSubmitting}
              className="flex items-center gap-2 bg-pink-600 text-white px-6 py-3 rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Novo Produto
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
            <h3 className="text-slate-600 text-sm font-medium">Total de Produtos</h3>
            <p className="text-3xl font-bold text-slate-900 mt-2">{products.length}</p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Produtos Ativos</h3>
            <p className="text-3xl font-bold text-green-600 mt-2">{activeProducts}</p>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <h3 className="text-slate-600 text-sm font-medium">Valor Total (Com Markup)</h3>
            <p className="text-3xl font-bold text-blue-600 mt-2">
              R$ {totalProductValue.toFixed(2)}
            </p>
          </div>
        </motion.div>

        {/* Products Table */}
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
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Preço Base</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Markup (%)</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Preço Final</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Status</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Ação</th>
                </tr>
              </thead>
              <tbody>
                {products.map((product) => {
                  const finalPrice = calculateFinalPrice(product.basePrice, product.markup);
                  return (
                    <tr key={product.id} className="border-b border-slate-100 hover:bg-slate-50">
                      <td className="px-6 py-4 text-sm text-slate-900 font-medium">{product.name}</td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        R$ {parseFloat(product.basePrice).toFixed(2)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-600">
                        {parseFloat(product.markup).toFixed(1)}%
                      </td>
                      <td className="px-6 py-4 text-sm font-semibold text-green-600">
                        R$ {finalPrice.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                          product.isActive 
                            ? 'bg-green-100 text-green-700' 
                            : 'bg-slate-100 text-slate-700'
                        }`}>
                          {product.isActive ? 'Ativo' : 'Inativo'}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <button
                          onClick={() => handleDeleteProduct(product.id)}
                          className="text-red-600 hover:text-red-700 disabled:text-slate-300 disabled:cursor-not-allowed transition-colors"
                          disabled={loading || isSubmitting}
                          title={loading || isSubmitting ? 'Operação em progresso' : 'Deletar produto'}
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


        {/* Add Product Dialog */}
        {openDialog && (
          <div className="fixed inset-0 flex items-center justify-center z-50 backdrop-blur-md">
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="bg-white rounded-2xl p-8 max-w-md w-full shadow-xl"
            >
              <h2 className="text-2xl font-bold text-slate-900 mb-6">Adicionar Produto</h2>
              
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
                  <label className="block text-sm font-medium text-slate-700 mb-2">Nome do Produto</label>
                  <input
                    type="text"
                    value={newProduct.name}
                    onChange={(e) => setNewProduct({ ...newProduct, name: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="Ex: Bolo de Chocolate Belga"
                    disabled={isSubmitting}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Descrição</label>
                  <textarea
                    value={newProduct.description}
                    onChange={(e) => setNewProduct({ ...newProduct, description: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none resize-none"
                    placeholder="Descrição do produto (opcional)"
                    rows="2"
                    disabled={isSubmitting}
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">Preço Base (R$)</label>
                    <input
                      type="text"
                      value={formatCurrency(newProduct.basePrice)}
                      onChange={(e) => {
                        const inputValue = e.target.value;
                        const numericOnly = inputValue.replace(/\D/g, '');
                        setNewProduct({ ...newProduct, basePrice: numericOnly });
                      }}
                      className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                      placeholder="R$ 0,00"
                      disabled={isSubmitting}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">Markup (%)</label>
                    <input
                      type="number"
                      step="0.1"
                      value={newProduct.markup}
                      onChange={(e) => setNewProduct({ ...newProduct, markup: e.target.value })}
                      className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                      placeholder="0"
                      disabled={isSubmitting}
                    />
                  </div>
                </div>

                {newProduct.basePrice && newProduct.markup && (
                  <div className="bg-blue-50 rounded-lg p-3">
                    <p className="text-xs text-blue-600 mb-1">Preço Final Estimado:</p>
                    <p className="text-xl font-bold text-blue-700">
                      R$ {calculateFinalPrice(
                        parseInt(unformatCurrency(newProduct.basePrice)) / 100 || 0,
                        newProduct.markup
                      ).toFixed(2)}
                    </p>
                  </div>
                )}

                <div className="flex items-center gap-3 bg-slate-50 p-3 rounded-lg">
                  <input
                    type="checkbox"
                    checked={newProduct.isActive}
                    onChange={(e) => setNewProduct({ ...newProduct, isActive: e.target.checked })}
                    className="w-5 h-5 rounded border-slate-300 text-pink-600 focus:ring-pink-500"
                    disabled={isSubmitting}
                  />
                  <label className="text-sm font-medium text-slate-700">Produto Ativo</label>
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
                  onClick={handleAddProduct}
                  disabled={isSubmitting || loading || !newProduct.name.trim() || !newProduct.basePrice}
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

export default ProductsPage;
