// eslint-disable-next-line no-unused-vars
import { motion } from 'framer-motion';
import { useState } from 'react';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { ArrowUpRight, Package, ShoppingCart, CurrencyDollar } from 'phosphor-react';

const DashboardPage = () => {
  const [activeTab, setActiveTab] = useState('overview');
  // Dados fake para vendas mensais
  const salesdData = [
    { month: 'Jan', sales: 4000, revenue: 2400 },
    { month: 'Fev', sales: 3000, revenue: 1398 },
    { month: 'Mar', sales: 2000, revenue: 9800 },
    { month: 'Abr', sales: 2780, revenue: 3908 },
    { month: 'Mai', sales: 1890, revenue: 4800 },
    { month: 'Jun', sales: 2390, revenue: 3800 },
  ];

  // Dados fake para produtos mais vendidos
  const topProductsData = [
    { name: 'Bolo de Chocolate', sales: 450 },
    { name: 'Brigadeiro', sales: 380 },
    { name: 'Mousse', sales: 320 },
    { name: 'Torta de Morango', sales: 290 },
    { name: 'Docinhos', sales: 250 },
  ];

  // Dados fake para distribuição de vendas por categoria
  const categoriesData = [
    { name: 'Bolos', value: 35 },
    { name: 'Doces', value: 25 },
    { name: 'Tortas', value: 20 },
    { name: 'Outros', value: 20 },
  ];

  const COLORS = ['#ec4899', '#f472b6', '#fbcfe8', '#fce7f3'];

  // Métricas do topo
  const metrics = [
    {
      title: 'Receita Total',
      value: 'R$ 24.500',
      icon: <CurrencyDollar size={32} />,
      color: 'bg-green-100',
      textColor: 'text-green-600',
    },
    {
      title: 'Vendas',
      value: '1.850',
      icon: <ShoppingCart size={32} />,
      color: 'bg-blue-100',
      textColor: 'text-blue-600',
    },
    {
      title: 'Produtos',
      value: '45',
      icon: <Package size={32} />,
      color: 'bg-purple-100',
      textColor: 'text-purple-600',
    },
    {
      title: 'Crescimento',
      value: '+12.5%',
      icon: <ArrowUpRight size={32} />,
      color: 'bg-orange-100',
      textColor: 'text-orange-600',
    },
  ];

  return (
    <>
      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <h1 className="text-3xl font-bold text-slate-900">Painel</h1>
          <p className="text-slate-600 mt-2">Bem-vindo de volta! Aqui está um resumo do seu negócio.</p>
        </div>
      </header>

      {/* Tab Navigation */}
      <div className="bg-white border-b border-slate-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex gap-8">
            <button
              onClick={() => setActiveTab('overview')}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === 'overview'
                  ? 'border-pink-600 text-pink-600'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              Visão Geral
            </button>
            <button
              onClick={() => setActiveTab('revenue')}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === 'revenue'
                  ? 'border-pink-600 text-pink-600'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              Receita
            </button>
            <button
              onClick={() => setActiveTab('sales')}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === 'sales'
                  ? 'border-pink-600 text-pink-600'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              Vendas
            </button>
            <button
              onClick={() => setActiveTab('products')}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === 'products'
                  ? 'border-pink-600 text-pink-600'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              Produtos
            </button>
          </div>
        </div>
      </div>

      {/* Tab Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
          {/* Overview Tab */}
          {activeTab === 'overview' && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="space-y-8"
            >
              {/* Métricas do Topo */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
              >
                {metrics.map((metric, i) => (
                  <motion.div
                    key={i}
                    whileHover={{ scale: 1.05 }}
                    className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm"
                  >
                    <div className="flex items-center justify-between mb-4">
                      <div className={`${metric.color} rounded-xl p-3`}>
                        <div className={metric.textColor}>{metric.icon}</div>
                      </div>
                    </div>
                    <h3 className="text-slate-600 text-sm font-medium">{metric.title}</h3>
                    <p className="text-2xl font-bold text-slate-900 mt-2">{metric.value}</p>
                  </motion.div>
                ))}
              </motion.div>

              {/* Gráficos */}
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Gráfico de Vendas */}
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.1 }}
                  className="lg:col-span-2 bg-white rounded-2xl p-6 border border-slate-100 shadow-sm"
                >
                  <h2 className="text-xl font-bold text-slate-900 mb-6">Vendas e Receita</h2>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={salesdData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                      <XAxis dataKey="month" stroke="#94a3b8" />
                      <YAxis stroke="#94a3b8" />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: '#1e293b',
                          border: 'none',
                          borderRadius: '8px',
                          color: '#fff',
                        }}
                      />
                      <Legend />
                      <Line
                        type="monotone"
                        dataKey="sales"
                        stroke="#ec4899"
                        strokeWidth={2}
                        dot={{ fill: '#ec4899' }}
                        activeDot={{ r: 6 }}
                      />
                      <Line
                        type="monotone"
                        dataKey="revenue"
                        stroke="#8b5cf6"
                        strokeWidth={2}
                        dot={{ fill: '#8b5cf6' }}
                        activeDot={{ r: 6 }}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </motion.div>

                {/* Gráfico de Categorias */}
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.2 }}
                  className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm"
                >
                  <h2 className="text-xl font-bold text-slate-900 mb-6">Categorias</h2>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={categoriesData}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ name, value }) => `${name} ${value}%`}
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="value"
                      >
                        {categoriesData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value) => `${value}%`} />
                    </PieChart>
                  </ResponsiveContainer>
                </motion.div>
              </div>

              {/* Produtos Mais Vendidos */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
                className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm"
              >
                <h2 className="text-xl font-bold text-slate-900 mb-6">Produtos Mais Vendidos</h2>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={topProductsData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis dataKey="name" stroke="#94a3b8" />
                    <YAxis stroke="#94a3b8" />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1e293b',
                        border: 'none',
                        borderRadius: '8px',
                        color: '#fff',
                      }}
                    />
                    <Bar dataKey="sales" fill="#ec4899" radius={[8, 8, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </motion.div>
            </motion.div>
          )}

          {/* Revenue Tab */}
          {activeTab === 'revenue' && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="space-y-8"
            >
              <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
                <h2 className="text-2xl font-bold text-slate-900 mb-6">Receita Mensal</h2>
                <ResponsiveContainer width="100%" height={400}>
                  <LineChart data={salesdData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis dataKey="month" stroke="#94a3b8" />
                    <YAxis stroke="#94a3b8" />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1e293b',
                        border: 'none',
                        borderRadius: '8px',
                        color: '#fff',
                      }}
                    />
                    <Legend />
                    <Line
                      type="monotone"
                      dataKey="revenue"
                      stroke="#10b981"
                      strokeWidth={3}
                      dot={{ fill: '#10b981', r: 6 }}
                      activeDot={{ r: 8 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </motion.div>
          )}

          {/* Sales Tab */}
          {activeTab === 'sales' && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="space-y-8"
            >
              <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
                <h2 className="text-2xl font-bold text-slate-900 mb-6">Vendas Mensais</h2>
                <ResponsiveContainer width="100%" height={400}>
                  <BarChart data={salesdData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis dataKey="month" stroke="#94a3b8" />
                    <YAxis stroke="#94a3b8" />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1e293b',
                        border: 'none',
                        borderRadius: '8px',
                        color: '#fff',
                      }}
                    />
                    <Bar dataKey="sales" fill="#3b82f6" radius={[8, 8, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </motion.div>
          )}

          {/* Products Tab */}
          {activeTab === 'products' && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="space-y-8"
            >
              <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
                <h2 className="text-2xl font-bold text-slate-900 mb-6">Top 5 Produtos Mais Vendidos</h2>
                <ResponsiveContainer width="100%" height={400}>
                  <BarChart data={topProductsData} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis type="number" stroke="#94a3b8" />
                    <YAxis dataKey="name" type="category" stroke="#94a3b8" width={150} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1e293b',
                        border: 'none',
                        borderRadius: '8px',
                        color: '#fff',
                      }}
                    />
                    <Bar dataKey="sales" fill="#a855f7" radius={[0, 8, 8, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </motion.div>
          )}
        </main>
      </>
    );
};

export default DashboardPage;
