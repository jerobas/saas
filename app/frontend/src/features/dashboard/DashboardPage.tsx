import { motion } from "motion/react";
import { useEffect, useState } from "react";
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
} from "recharts";
import { ArrowUpRight, Package, ShoppingCart, CurrencyDollar } from "@phosphor-icons/react";
import {
  catalogGateway,
  reportingGateway,
  type ReportingPeriodRequest,
  type SalesReportResponse,
} from "../../gateways/desktopBridge";

type DashboardTab = "overview" | "revenue" | "sales" | "products";

type ReportingEndpointResult =
  { status: "fulfilled"; value: unknown } | { status: "rejected"; reason: string };

interface HiddenReportingDump {
  status: "idle" | "loading" | "loaded";
  period: ReportingPeriodRequest | null;
  endpoints: Record<string, ReportingEndpointResult>;
}

interface DashboardMetricData {
  salesReport: SalesReportResponse | null;
  activeProductCount: number | null;
}

interface DashboardMetricsState extends DashboardMetricData {
  status: "idle" | "loading" | "loaded";
}

type ReportingEndpoint = [name: string, load: () => Promise<unknown>];

const businessDate = (date: Date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};

const defaultReportingPeriod = (): ReportingPeriodRequest => {
  const today = new Date();
  const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
  return {
    fromOccurredOn: businessDate(firstDayOfMonth),
    toOccurredOn: businessDate(today),
    granularity: "MONTH",
  };
};

const loadHiddenReportingDump = async (
  period: ReportingPeriodRequest,
  salesReportPromise: Promise<SalesReportResponse>,
): Promise<Record<string, ReportingEndpointResult>> => {
  const endpoints: ReportingEndpoint[] = [
    ["salesReport", () => salesReportPromise],
    ["inventoryReport", () => reportingGateway.getInventoryReport(period)],
    ["purchaseReport", () => reportingGateway.getPurchaseReport(period)],
    ["productionReport", () => reportingGateway.getProductionReport(period)],
    ["adjustmentReport", () => reportingGateway.getAdjustmentReport(period)],
    ["categoryMixReport", () => reportingGateway.getCategoryMixReport(period)],
  ];

  const entries = await Promise.all(
    endpoints.map(async ([name, load]): Promise<[string, ReportingEndpointResult]> => {
      try {
        return [name, { status: "fulfilled", value: await load() }];
      } catch (error) {
        return [
          name,
          {
            status: "rejected",
            reason: error instanceof Error ? error.message : String(error),
          },
        ];
      }
    }),
  );

  return Object.fromEntries(entries);
};

const loadActiveProductCount = async () => {
  let count = 0;
  let after: { name: string; id: number } | undefined;

  do {
    const page = await catalogGateway.listItems({
      archiveFilter: "ACTIVE",
      requireCapabilities: { purchasable: false, producible: false, sellable: false },
      after,
      pageSize: 100,
    });
    count += page.items.length;
    after = page.next ? { name: page.next.name, id: page.next.id } : undefined;
  } while (after);

  return count;
};

const loadVisibleMetrics = async (
  salesReportPromise: Promise<SalesReportResponse>,
): Promise<DashboardMetricData> => {
  const [salesResult, productsResult] = await Promise.allSettled([
    salesReportPromise,
    loadActiveProductCount(),
  ]);
  return {
    salesReport: salesResult.status === "fulfilled" ? salesResult.value : null,
    activeProductCount: productsResult.status === "fulfilled" ? productsResult.value : null,
  };
};

const formatMoney = (minor: number, currencyCode: string, minorDigits: number) =>
  new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: currencyCode,
    minimumFractionDigits: minorDigits,
    maximumFractionDigits: minorDigits,
  }).format(minor / 10 ** minorDigits);

const formatCount = (value: number) => new Intl.NumberFormat("pt-BR").format(value);

const formatGrowth = (basisPoints: number | null | undefined) => {
  if (basisPoints === null || basisPoints === undefined) return "Sem comparativo";
  const value = new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 2 }).format(
    basisPoints / 100,
  );
  return `${basisPoints > 0 ? "+" : ""}${value}%`;
};

const DashboardPage = () => {
  const [activeTab, setActiveTab] = useState<DashboardTab>("overview");
  const [hiddenReportingDump, setHiddenReportingDump] = useState<HiddenReportingDump>(() => ({
    status: "idle",
    period: null,
    endpoints: {},
  }));
  const [dashboardMetrics, setDashboardMetrics] = useState<DashboardMetricsState>(() => ({
    status: "idle",
    salesReport: null,
    activeProductCount: null,
  }));

  useEffect(() => {
    let cancelled = false;
    const period = defaultReportingPeriod();
    const salesReportPromise = reportingGateway.getSalesReport(period);

    setHiddenReportingDump({
      status: "loading",
      period,
      endpoints: {},
    });
    setDashboardMetrics({ status: "loading", salesReport: null, activeProductCount: null });

    Promise.all([
      loadHiddenReportingDump(period, salesReportPromise),
      loadVisibleMetrics(salesReportPromise),
    ]).then(([endpoints, metrics]) => {
      if (cancelled) {
        return;
      }
      setHiddenReportingDump({
        status: "loaded",
        period,
        endpoints,
      });
      setDashboardMetrics({ status: "loaded", ...metrics });
    });

    return () => {
      cancelled = true;
    };
  }, []);

  // Preview data kept intentionally until the visible widgets use the wired V2 reports.
  const salesdData = [
    { month: "Jan", sales: 4000, revenue: 2400 },
    { month: "Fev", sales: 3000, revenue: 1398 },
    { month: "Mar", sales: 2000, revenue: 9800 },
    { month: "Abr", sales: 2780, revenue: 3908 },
    { month: "Mai", sales: 1890, revenue: 4800 },
    { month: "Jun", sales: 2390, revenue: 3800 },
  ];

  const topProductsData = [
    { name: "Bolo de Chocolate", sales: 450 },
    { name: "Brigadeiro", sales: 380 },
    { name: "Mousse", sales: 320 },
    { name: "Torta de Morango", sales: 290 },
    { name: "Docinhos", sales: 250 },
  ];

  const categoriesData = [
    { name: "Bolos", value: 35 },
    { name: "Doces", value: 25 },
    { name: "Tortas", value: 20 },
    { name: "Outros", value: 20 },
  ];

  const COLORS = ["#ec4899", "#f472b6", "#fbcfe8", "#fce7f3"];

  const salesReport = dashboardMetrics.salesReport;
  const loadingMetric = dashboardMetrics.status !== "loaded" ? "..." : "—";

  const metrics = [
    {
      title: "Receita no período",
      value: salesReport
        ? formatMoney(
            salesReport.commercialTotalMinor,
            salesReport.currencyCode,
            salesReport.currencyMinorDigits,
          )
        : loadingMetric,
      icon: <CurrencyDollar size={32} />,
      color: "bg-green-100",
      textColor: "text-green-600",
    },
    {
      title: "Vendas no período",
      value: salesReport ? formatCount(salesReport.totalSalesCount) : loadingMetric,
      icon: <ShoppingCart size={32} />,
      color: "bg-blue-100",
      textColor: "text-blue-600",
    },
    {
      title: "Produtos ativos",
      value:
        dashboardMetrics.activeProductCount === null
          ? loadingMetric
          : formatCount(dashboardMetrics.activeProductCount),
      icon: <Package size={32} />,
      color: "bg-purple-100",
      textColor: "text-purple-600",
    },
    {
      title: "Crescimento",
      value: salesReport ? formatGrowth(salesReport.growthBasisPoints) : loadingMetric,
      icon: <ArrowUpRight size={32} />,
      color: "bg-orange-100",
      textColor: "text-orange-600",
    },
  ];

  return (
    <>
      <section
        aria-hidden="true"
        data-testid="dashboard-hidden-reporting-wire"
        style={{ display: "none" }}
      >
        <h2>Hidden reporting endpoint dump</h2>
        <pre>{JSON.stringify(hiddenReportingDump, null, 2)}</pre>
      </section>

      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <h1 className="text-3xl font-bold text-slate-900">Painel</h1>
          <p className="text-slate-600 mt-2">
            Visão geral da operação no mês atual, com indicadores calculados pelos relatórios V2.
          </p>
        </div>
      </header>

      <div className="bg-amber-50 border-b border-amber-100">
        <div className="max-w-7xl mx-auto px-6 py-3 text-sm text-amber-900">
          Os cards já usam dados reais. Os gráficos abaixo ainda preservam dados de demonstração até
          a próxima etapa visual.
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="bg-white border-b border-slate-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex gap-8">
            <button
              onClick={() => setActiveTab("overview")}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === "overview"
                  ? "border-pink-600 text-pink-600"
                  : "border-transparent text-slate-600 hover:text-slate-900"
              }`}
            >
              Visão Geral
            </button>
            <button
              onClick={() => setActiveTab("revenue")}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === "revenue"
                  ? "border-pink-600 text-pink-600"
                  : "border-transparent text-slate-600 hover:text-slate-900"
              }`}
            >
              Receita
            </button>
            <button
              onClick={() => setActiveTab("sales")}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === "sales"
                  ? "border-pink-600 text-pink-600"
                  : "border-transparent text-slate-600 hover:text-slate-900"
              }`}
            >
              Vendas
            </button>
            <button
              onClick={() => setActiveTab("products")}
              className={`py-4 px-2 border-b-2 font-semibold transition-all ${
                activeTab === "products"
                  ? "border-pink-600 text-pink-600"
                  : "border-transparent text-slate-600 hover:text-slate-900"
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
        {activeTab === "overview" && (
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
                        backgroundColor: "#1e293b",
                        border: "none",
                        borderRadius: "8px",
                        color: "#fff",
                      }}
                    />
                    <Legend />
                    <Line
                      type="monotone"
                      dataKey="sales"
                      stroke="#ec4899"
                      strokeWidth={2}
                      dot={{ fill: "#ec4899" }}
                      activeDot={{ r: 6 }}
                    />
                    <Line
                      type="monotone"
                      dataKey="revenue"
                      stroke="#8b5cf6"
                      strokeWidth={2}
                      dot={{ fill: "#8b5cf6" }}
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
                      backgroundColor: "#1e293b",
                      border: "none",
                      borderRadius: "8px",
                      color: "#fff",
                    }}
                  />
                  <Bar dataKey="sales" fill="#ec4899" radius={[8, 8, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </motion.div>
          </motion.div>
        )}

        {/* Revenue Tab */}
        {activeTab === "revenue" && (
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
                      backgroundColor: "#1e293b",
                      border: "none",
                      borderRadius: "8px",
                      color: "#fff",
                    }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="revenue"
                    stroke="#10b981"
                    strokeWidth={3}
                    dot={{ fill: "#10b981", r: 6 }}
                    activeDot={{ r: 8 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </motion.div>
        )}

        {/* Sales Tab */}
        {activeTab === "sales" && (
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
                      backgroundColor: "#1e293b",
                      border: "none",
                      borderRadius: "8px",
                      color: "#fff",
                    }}
                  />
                  <Bar dataKey="sales" fill="#3b82f6" radius={[8, 8, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </motion.div>
        )}

        {/* Products Tab */}
        {activeTab === "products" && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="space-y-8"
          >
            <div className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
              <h2 className="text-2xl font-bold text-slate-900 mb-6">
                Top 5 Produtos Mais Vendidos
              </h2>
              <ResponsiveContainer width="100%" height={400}>
                <BarChart data={topProductsData} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis type="number" stroke="#94a3b8" />
                  <YAxis dataKey="name" type="category" stroke="#94a3b8" width={150} />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "#1e293b",
                      border: "none",
                      borderRadius: "8px",
                      color: "#fff",
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
