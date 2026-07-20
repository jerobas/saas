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
  type CategoryMixReportResponse,
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
  monthlySalesReport: SalesReportResponse | null;
  activeProductCount: number | null;
  categoryMixReport: CategoryMixReportResponse | null;
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
    granularity: "DAY",
  };
};

const monthlyReportingPeriod = (): ReportingPeriodRequest => {
  const today = new Date();
  const firstMonth = new Date(today.getFullYear(), today.getMonth() - 5, 1);
  return {
    fromOccurredOn: businessDate(firstMonth),
    toOccurredOn: businessDate(today),
    granularity: "MONTH",
  };
};

const loadHiddenReportingDump = async (
  period: ReportingPeriodRequest,
  salesReportPromise: Promise<SalesReportResponse>,
  categoryMixReportPromise: Promise<CategoryMixReportResponse>,
): Promise<Record<string, ReportingEndpointResult>> => {
  const endpoints: ReportingEndpoint[] = [
    ["salesReport", () => salesReportPromise],
    ["inventoryReport", () => reportingGateway.getInventoryReport(period)],
    ["purchaseReport", () => reportingGateway.getPurchaseReport(period)],
    ["productionReport", () => reportingGateway.getProductionReport(period)],
    ["adjustmentReport", () => reportingGateway.getAdjustmentReport(period)],
    ["categoryMixReport", () => categoryMixReportPromise],
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
  monthlySalesReportPromise: Promise<SalesReportResponse>,
  categoryMixReportPromise: Promise<CategoryMixReportResponse>,
): Promise<DashboardMetricData> => {
  const [salesResult, monthlySalesResult, productsResult, categoryMixResult] =
    await Promise.allSettled([
      salesReportPromise,
      monthlySalesReportPromise,
      loadActiveProductCount(),
      categoryMixReportPromise,
    ]);
  return {
    salesReport: salesResult.status === "fulfilled" ? salesResult.value : null,
    monthlySalesReport: monthlySalesResult.status === "fulfilled" ? monthlySalesResult.value : null,
    activeProductCount: productsResult.status === "fulfilled" ? productsResult.value : null,
    categoryMixReport: categoryMixResult.status === "fulfilled" ? categoryMixResult.value : null,
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
    monthlySalesReport: null,
    activeProductCount: null,
    categoryMixReport: null,
  }));

  useEffect(() => {
    let cancelled = false;
    const period = defaultReportingPeriod();
    const salesReportPromise = reportingGateway.getSalesReport(period);
    const monthlySalesReportPromise = reportingGateway.getSalesReport(monthlyReportingPeriod());
    const categoryMixReportPromise = reportingGateway.getCategoryMixReport(period);

    setHiddenReportingDump({
      status: "loading",
      period,
      endpoints: {},
    });
    setDashboardMetrics({
      status: "loading",
      salesReport: null,
      monthlySalesReport: null,
      activeProductCount: null,
      categoryMixReport: null,
    });

    Promise.all([
      loadHiddenReportingDump(period, salesReportPromise, categoryMixReportPromise),
      loadVisibleMetrics(salesReportPromise, monthlySalesReportPromise, categoryMixReportPromise),
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

  const COLORS = ["#ec4899", "#f472b6", "#fbcfe8", "#fce7f3"];

  const salesReport = dashboardMetrics.salesReport;
  const monthlySalesReport = dashboardMetrics.monthlySalesReport;
  const currencyDivisor = salesReport ? 10 ** salesReport.currencyMinorDigits : 100;
  const monthlyCurrencyDivisor = monthlySalesReport
    ? 10 ** monthlySalesReport.currencyMinorDigits
    : 100;
  const salesRevenueData =
    salesReport?.salesRevenueSeries.map((row) => ({
      period: row.label,
      sales: row.salesCount,
      revenue: row.commercialTotalMinor / currencyDivisor,
    })) ?? [];
  const monthlyRevenueData =
    monthlySalesReport?.monthlyRevenueSeries.map((row) => ({
      period: row.label,
      revenue: row.commercialTotalMinor / monthlyCurrencyDivisor,
    })) ?? [];
  const monthlySalesData =
    monthlySalesReport?.monthlySalesSeries.map((row) => ({
      period: row.label,
      sales: row.salesCount,
    })) ?? [];
  const topProductsData =
    salesReport?.topProductsByQuantity.map((row) => ({
      name: row.itemName,
      sales: row.quantityAtomic,
    })) ?? [];
  const categoriesData =
    dashboardMetrics.categoryMixReport?.available === true
      ? dashboardMetrics.categoryMixReport.rows.map((row) => ({
          name: row.categoryName,
          value: row.shareBasisPoints / 100,
        }))
      : [];
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
          Indicadores e gráficos usam relatórios reais. O mix por categoria permanece indisponível
          porque o catálogo V2 ainda não possui categorias.
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
                  <LineChart data={salesRevenueData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                    <XAxis dataKey="period" stroke="#94a3b8" />
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
                      name="Vendas"
                      stroke="#ec4899"
                      strokeWidth={2}
                      dot={{ fill: "#ec4899" }}
                      activeDot={{ r: 6 }}
                    />
                    <Line
                      type="monotone"
                      dataKey="revenue"
                      name="Receita"
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
                {categoriesData.length === 0 ? (
                  <div className="flex h-[300px] flex-col items-center justify-center rounded-xl border border-dashed border-slate-300 bg-slate-50 px-6 text-center">
                    <p className="font-semibold text-slate-800">Mix por categoria indisponível</p>
                    <p className="mt-2 text-sm text-slate-500">
                      O catálogo V2 ainda não possui categorias. O gráfico será ativado quando essa
                      dimensão existir.
                    </p>
                  </div>
                ) : (
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
                )}
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
                  <Bar
                    dataKey="sales"
                    name="Quantidade vendida"
                    fill="#ec4899"
                    radius={[8, 8, 0, 0]}
                  />
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
                <LineChart data={monthlyRevenueData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="period" stroke="#94a3b8" />
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
                    name="Receita"
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
                <BarChart data={monthlySalesData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="period" stroke="#94a3b8" />
                  <YAxis stroke="#94a3b8" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "#1e293b",
                      border: "none",
                      borderRadius: "8px",
                      color: "#fff",
                    }}
                  />
                  <Bar dataKey="sales" name="Vendas" fill="#3b82f6" radius={[8, 8, 0, 0]} />
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
                  <Bar
                    dataKey="sales"
                    name="Quantidade vendida"
                    fill="#a855f7"
                    radius={[0, 8, 8, 0]}
                  />
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
