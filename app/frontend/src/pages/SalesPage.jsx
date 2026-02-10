/* eslint-disable no-unused-vars */
import { useState } from 'react';
import { motion } from 'framer-motion';
import { Plus, Trash, MagnifyingGlass, ShoppingCart } from 'phosphor-react';

const SalesPage = () => {
  // Dados mockados de produtos
  const mockProducts = [
    {
      id: '1',
      name: 'Bolo de Chocolate Belga',
      description: 'Bolo de chocolate belga com cobertura de ganache',
      basePrice: 45.00,
      markup: 35,
      isActive: true,
    },
    {
      id: '2',
      name: 'Bolo Red Velvet',
      description: 'Clássico bolo red velvet com cream cheese',
      basePrice: 50.00,
      markup: 30,
      isActive: true,
    },
    {
      id: '3',
      name: 'Pavê de Morango',
      description: 'Pavê refrescante com morangos frescos',
      basePrice: 38.00,
      markup: 40,
      isActive: true,
    },
    {
      id: '4',
      name: 'Torta de Chocolate com Framboesa',
      description: 'Torta sofisticada com framboesa fresca',
      basePrice: 65.00,
      markup: 25,
      isActive: true,
    },
    {
      id: '5',
      name: 'Broinhas de Chuva',
      description: 'Dúzia de broinhas de chuva crocantes (12 un)',
      basePrice: 18.00,
      markup: 50,
      isActive: true,
    },
    {
      id: '6',
      name: 'Brigadeiro Tradicional',
      description: 'Pote com 500g de brigadeiro caseiro',
      basePrice: 22.00,
      markup: 55,
      isActive: true,
    },
    {
      id: '7',
      name: 'Cupcakes de Vanilla',
      description: 'Caixa com 6 cupcakes de vanilla com cobertura',
      basePrice: 35.00,
      markup: 45,
      isActive: true,
    },
    {
      id: '8',
      name: 'Docinhos para Festas',
      description: 'Seleção com 30 docinhos variados',
      basePrice: 55.00,
      markup: 35,
      isActive: true,
    },
    {
      id: '9',
      name: 'Bolo de Cenoura com Cobertura',
      description: 'Clássico bolo de cenoura com cobertura de chocolate',
      basePrice: 40.00,
      markup: 38,
      isActive: true,
    },
    {
      id: '10',
      name: 'Pudim de Leite Condensado',
      description: 'Pudim caseiro com calda de caramelo',
      basePrice: 25.00,
      markup: 48,
      isActive: true,
    },
  ];

  const products = mockProducts; // Usando apenas os dados mockados

  const [searchTerm, setSearchTerm] = useState('');
  const [cartItems, setCartItems] = useState([]);
  const [saleQuantity, setSaleQuantity] = useState('');
  const [selectedProductId, setSelectedProductId] = useState(null);

  /**
   * Filtra produtos pela busca
   */
  const filteredProducts = products.filter(product =>
    product.name.toLowerCase().includes(searchTerm.toLowerCase()) &&
    product.isActive
  );

  /**
   * Formata valor em moeda (BRL)
   */
  const formatCurrency = (value) => {
    if (!value) return '';
    const numericValue = value.toString().replace(/\D/g, '');
    if (!numericValue) return '';
    const numberValue = parseInt(numericValue, 10) / 100;
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(numberValue);
  };

  /**
   * Calcula o preço final do produto
   */
  const calculateFinalPrice = (basePrice, markup) => {
    const base = parseFloat(basePrice) || 0;
    const markupValue = parseFloat(markup) || 0;
    return base + (base * (markupValue / 100));
  };

  /**
   * Adiciona produto ao carrinho
   */
  const addToCart = (product) => {
    if (!saleQuantity || parseFloat(saleQuantity) <= 0) {
      alert('Digite uma quantidade válida');
      return;
    }

    const finalPrice = calculateFinalPrice(product.basePrice, product.markup);
    const quantity = parseFloat(saleQuantity);

    // Verificar se já existe no carrinho
    const existingItem = cartItems.find(item => item.id === product.id);

    if (existingItem) {
      // Atualizar quantidade
      setCartItems(cartItems.map(item =>
        item.id === product.id
          ? { ...item, quantity: item.quantity + quantity }
          : item
      ));
    } else {
      // Adicionar novo item
      setCartItems([...cartItems, {
        id: product.id,
        name: product.name,
        basePrice: product.basePrice,
        markup: product.markup,
        finalPrice: finalPrice,
        quantity: quantity
      }]);
    }

    // Limpar campos
    setSaleQuantity('');
    setSelectedProductId(null);
  };

  /**
   * Remove item do carrinho
   */
  const removeFromCart = (productId) => {
    setCartItems(cartItems.filter(item => item.id !== productId));
  };

  /**
   * Atualiza quantidade do item no carrinho
   */
  const updateCartQuantity = (productId, newQuantity) => {
    if (newQuantity <= 0) {
      removeFromCart(productId);
      return;
    }
    setCartItems(cartItems.map(item =>
      item.id === productId
        ? { ...item, quantity: parseFloat(newQuantity) }
        : item
    ));
  };

  /**
   * Calcula total da venda
   */
  const calculoTotal = cartItems.reduce((total, item) => {
    return total + (item.finalPrice * item.quantity);
  }, 0);

  /**
   * Finaliza a venda (mockado)
   */
  const finalizeSale = () => {
    if (cartItems.length === 0) {
      alert('Adicione produtos ao carrinho');
      return;
    }

     
    const venda_id = Math.random().toString(36).substr(2, 9).toUpperCase();

    // Log detalhado da venda
    const saleData = {
      id: `VENDA-${venda_id}`,
      items: cartItems,
      total: calculoTotal,
      timestamp: new Date().toLocaleString('pt-BR'),
      itemsCount: cartItems.reduce((acc, item) => acc + item.quantity, 0)
    };

    console.log('🛒 Venda Finalizada:', saleData);

    // Mensagem detalhada
    const itemsList = cartItems.map(item => `• ${item.name} (${item.quantity} un)`).join('\n');
    const message = `✅ Venda Realizada com Sucesso!\n\n${itemsList}\n\nTotal: R$ ${calculoTotal.toFixed(2)}\nID: ${saleData.id}`;
    
    alert(message);
    setCartItems([]);
  };

  return (
    <>
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <h1 className="text-3xl font-bold text-slate-900">Vendas Rápidas</h1>
          <p className="text-slate-600 mt-2">Debite vendas e controle seu estoque em tempo real</p>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        {/* Stats Cards */}
        {cartItems.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8"
          >
            <div className="bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl p-4 border border-blue-200">
              <p className="text-xs text-blue-600 font-medium">Itens no Carrinho</p>
              <p className="text-2xl font-bold text-blue-700 mt-1">
                {cartItems.reduce((acc, item) => acc + item.quantity, 0).toFixed(1)}
              </p>
            </div>
            <div className="bg-gradient-to-br from-purple-50 to-purple-100 rounded-xl p-4 border border-purple-200">
              <p className="text-xs text-purple-600 font-medium">Produtos</p>
              <p className="text-2xl font-bold text-purple-700 mt-1">{cartItems.length}</p>
            </div>
            <div className="bg-gradient-to-br from-green-50 to-green-100 rounded-xl p-4 border border-green-200">
              <p className="text-xs text-green-600 font-medium">Subtotal</p>
              <p className="text-2xl font-bold text-green-700 mt-1">R$ {calculoTotal.toFixed(2)}</p>
            </div>
            <div className="bg-gradient-to-br from-orange-50 to-orange-100 rounded-xl p-4 border border-orange-200">
              <p className="text-xs text-orange-600 font-medium">Ticket Médio</p>
              <p className="text-2xl font-bold text-orange-700 mt-1">
                R$ {(calculoTotal / cartItems.length).toFixed(2)}
              </p>
            </div>
          </motion.div>
        )}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Seção de Busca e Produtos */}
          <div className="lg:col-span-2">
            {/* Busca */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-white rounded-2xl border border-slate-100 shadow-sm p-6 mb-6"
            >
              <label className="block text-sm font-medium text-slate-700 mb-3">
                Buscar Produto
              </label>
              <div className="flex gap-2">
                <div className="flex-1 relative">
                  <MagnifyingGlass
                    size={20}
                    className="absolute left-3 top-3 text-slate-400"
                  />
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Digite o nome do produto..."
                    className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                  />
                </div>
              </div>
              {searchTerm && (
                <p className="text-sm text-slate-600 mt-3">
                  {filteredProducts.length} produto(s) encontrado(s)
                </p>
              )}
            </motion.div>

            {/* Produtos Disponíveis */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden"
            >
              <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
                <h2 className="text-lg font-semibold text-slate-900">Produtos Disponíveis</h2>
              </div>

              <div className="divide-y divide-slate-200">
                {filteredProducts.length > 0 ? (
                  filteredProducts.map((product) => {
                    const finalPrice = calculateFinalPrice(product.basePrice, product.markup);
                    const isSelected = selectedProductId === product.id;

                    return (
                      <motion.div
                        key={product.id}
                        layout
                        className={`p-6 cursor-pointer transition-all ${
                          isSelected ? 'bg-pink-50 border-l-4 border-pink-600' : 'hover:bg-slate-50'
                        }`}
                        onClick={() => setSelectedProductId(isSelected ? null : product.id)}
                      >
                        <div className="flex justify-between items-start mb-2">
                          <div className="flex-1">
                            <h3 className="text-lg font-semibold text-slate-900">
                              {product.name}
                            </h3>
                            {product.description && (
                              <p className="text-sm text-slate-600 mt-1">{product.description}</p>
                            )}
                          </div>
                          <div className="text-right">
                            <p className="text-lg font-bold text-green-600">
                              R$ {finalPrice.toFixed(2)}
                            </p>
                            <p className="text-xs text-slate-500">
                              Base: R$ {parseFloat(product.basePrice).toFixed(2)}
                            </p>
                          </div>
                        </div>

                        {/* Painel de Quantidade (quando selecionado) */}
                        {isSelected && (
                          <motion.div
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: 'auto' }}
                            exit={{ opacity: 0, height: 0 }}
                            className="mt-4 pt-4 border-t border-pink-200"
                          >
                            <div className="flex gap-3">
                              <div className="flex-1">
                                <label className="block text-xs font-medium text-slate-700 mb-2">
                                  Quantidade
                                </label>
                                <input
                                  type="number"
                                  step="0.1"
                                  value={saleQuantity}
                                  onChange={(e) => setSaleQuantity(e.target.value)}
                                  placeholder="0"
                                  className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none text-sm"
                                  autoFocus
                                />
                              </div>
                              <div className="flex gap-2 items-end">
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    addToCart(product);
                                  }}
                                  disabled={!saleQuantity || parseFloat(saleQuantity) <= 0}
                                  className="flex items-center gap-2 px-4 py-2 bg-pink-600 text-white rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed font-medium"
                                >
                                  <Plus size={18} />
                                  Adicionar
                                </button>
                              </div>
                            </div>
                          </motion.div>
                        )}
                      </motion.div>
                    );
                  })
                ) : (
                  <div className="p-12 text-center">
                    <p className="text-slate-500">
                      {searchTerm ? 'Nenhum produto encontrado' : 'Nenhum produto disponível'}
                    </p>
                  </div>
                )}
              </div>
            </motion.div>
          </div>

          {/* Carrinho de Vendas */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden sticky top-20"
          >
            <div className="bg-gradient-to-r from-pink-600 to-pink-700 px-6 py-4 flex items-center gap-3">
              <ShoppingCart size={24} className="text-white" />
              <div>
                <h2 className="text-lg font-bold text-white">Carrinho</h2>
                <p className="text-pink-100 text-sm">{cartItems.length} item(ns)</p>
              </div>
            </div>

            {/* Itens do Carrinho */}
            <div className="divide-y divide-slate-200 max-h-96 overflow-y-auto">
              {cartItems.length > 0 ? (
                cartItems.map((item) => (
                  <motion.div
                    key={item.id}
                    layout
                    className="p-4 hover:bg-slate-50"
                  >
                    <div className="flex justify-between items-start mb-2">
                      <div className="flex-1">
                        <h4 className="font-semibold text-slate-900 text-sm">
                          {item.name}
                        </h4>
                        <p className="text-xs text-slate-600 mt-1">
                          R$ {item.finalPrice.toFixed(2)} x {item.quantity}
                        </p>
                      </div>
                      <button
                        onClick={() => removeFromCart(item.id)}
                        className="text-red-600 hover:text-red-700 transition-colors"
                      >
                        <Trash size={16} />
                      </button>
                    </div>

                    {/* Controle de Quantidade */}
                    <div className="flex gap-2 items-center">
                      <button
                        onClick={() => updateCartQuantity(item.id, item.quantity - 0.1)}
                        className="w-6 h-6 flex items-center justify-center border border-slate-300 rounded text-slate-600 hover:bg-slate-100 text-xs font-bold"
                      >
                        −
                      </button>
                      <input
                        type="number"
                        value={item.quantity}
                        onChange={(e) => updateCartQuantity(item.id, e.target.value)}
                        className="flex-1 text-center px-2 py-1 border border-slate-300 rounded text-sm"
                      />
                      <button
                        onClick={() => updateCartQuantity(item.id, item.quantity + 0.1)}
                        className="w-6 h-6 flex items-center justify-center border border-slate-300 rounded text-slate-600 hover:bg-slate-100 text-xs font-bold"
                      >
                        +
                      </button>
                    </div>

                    {/* Subtotal */}
                    <p className="text-right text-sm font-semibold text-slate-900 mt-2">
                      R$ {(item.finalPrice * item.quantity).toFixed(2)}
                    </p>
                  </motion.div>
                ))
              ) : (
                <div className="p-8 text-center">
                  <ShoppingCart size={32} className="text-slate-300 mx-auto mb-2" />
                  <p className="text-slate-500 text-sm">Carrinho vazio</p>
                </div>
              )}
            </div>

            {/* Resumo e Total */}
            {cartItems.length > 0 && (
              <div className="border-t border-slate-200 p-6 space-y-4">
                <div className="pt-4 border-t border-slate-200">
                  <div className="flex justify-between mb-2">
                    <span className="text-slate-600">Subtotal</span>
                    <span className="font-semibold">R$ {calculoTotal.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between text-lg font-bold text-pink-600">
                    <span>Total</span>
                    <span>R$ {calculoTotal.toFixed(2)}</span>
                  </div>
                </div>

                <button
                  onClick={finalizeSale}
                  className="w-full px-4 py-3 bg-gradient-to-r from-green-600 to-green-700 text-white rounded-lg hover:from-green-700 hover:to-green-800 transition-all font-bold flex items-center justify-center gap-2"
                >
                  <ShoppingCart size={20} />
                  Finalizar Venda
                </button>

                <button
                  onClick={() => setCartItems([])}
                  className="w-full px-4 py-2 border border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 transition-all font-medium"
                >
                  Limpar Carrinho
                </button>
              </div>
            )}
          </motion.div>
        </div>
      </main>
    </>
  );
};

export default SalesPage;
