import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import DashboardPage from "./DashboardPage";

const gatewayMocks = vi.hoisted(() => ({
  catalogGateway: {
    listItems: vi.fn(),
  },
  reportingGateway: {
    getSalesReport: vi.fn(),
    getInventoryReport: vi.fn(),
    getPurchaseReport: vi.fn(),
    getProductionReport: vi.fn(),
    getAdjustmentReport: vi.fn(),
    getCategoryMixReport: vi.fn(),
  },
}));

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

vi.mock("recharts", () => {
  const Chart = ({
    children,
    data,
    testId,
  }: {
    children?: ReactNode;
    data?: unknown;
    testId: string;
  }) => (
    <div data-testid={testId}>
      {JSON.stringify(data)}
      {children}
    </div>
  );
  const Container = ({ children }: { children?: ReactNode }) => <div>{children}</div>;
  const Empty = () => null;
  return {
    ResponsiveContainer: Container,
    LineChart: ({ children, data }: { children?: ReactNode; data?: unknown }) => (
      <Chart data={data} testId="line-chart">
        {children}
      </Chart>
    ),
    BarChart: ({ children, data }: { children?: ReactNode; data?: unknown }) => (
      <Chart data={data} testId="bar-chart">
        {children}
      </Chart>
    ),
    PieChart: Container,
    Line: Empty,
    Bar: Empty,
    Pie: Container,
    Cell: Empty,
    XAxis: Empty,
    YAxis: Empty,
    CartesianGrid: Empty,
    Tooltip: Empty,
    Legend: Empty,
  };
});

describe("DashboardPage", () => {
  beforeEach(() => {
    gatewayMocks.reportingGateway.getSalesReport.mockResolvedValue({
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      totalSalesCount: 12,
      commercialTotalMinor: 245_000,
      growthBasisPoints: 1_250,
      salesRevenueSeries: [
        { label: "20 jul", salesCount: 4, commercialTotalMinor: 80_000 },
        { label: "21 jul", salesCount: 8, commercialTotalMinor: 165_000 },
      ],
      monthlyRevenueSeries: [{ label: "jul 2026", salesCount: 12, commercialTotalMinor: 245_000 }],
      monthlySalesSeries: [{ label: "jul 2026", salesCount: 12, commercialTotalMinor: 245_000 }],
      topProductsByQuantity: [
        { itemName: "Brigadeiro real", quantityAtomic: 48 },
        { itemName: "Bolo real", quantityAtomic: 21 },
      ],
    });
    gatewayMocks.reportingGateway.getInventoryReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getPurchaseReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getProductionReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getAdjustmentReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getCategoryMixReport.mockResolvedValue({
      available: false,
      unavailableReason: "Categories are not modeled.",
      rows: [],
    });
    gatewayMocks.catalogGateway.listItems.mockImplementation(
      ({ after }: { after?: { name: string; id: number } }) =>
        Promise.resolve(
          after
            ? { items: [{ id: 3, name: "Torta" }], next: null }
            : {
                items: [
                  { id: 1, name: "Bolo" },
                  { id: 2, name: "Brigadeiro" },
                ],
                next: { name: "Brigadeiro", id: 2 },
              },
        ),
    );
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("shows real sales metrics and counts every active catalog page", async () => {
    render(<DashboardPage />);

    const revenueCard = screen.getByText("Receita no período").parentElement;
    const salesCard = screen.getByText("Vendas no período").parentElement;
    const productsCard = screen.getByText("Produtos ativos").parentElement;
    const growthCard = screen.getByText("Crescimento").parentElement;

    expect(revenueCard).not.toBeNull();
    expect(salesCard).not.toBeNull();
    expect(productsCard).not.toBeNull();
    expect(growthCard).not.toBeNull();

    await waitFor(() => {
      expect(within(revenueCard!).getByText(/2\.450,00/)).toBeInTheDocument();
      expect(within(salesCard!).getByText("12")).toBeInTheDocument();
      expect(within(productsCard!).getByText("3")).toBeInTheDocument();
      expect(within(growthCard!).getByText("+12,5%")).toBeInTheDocument();
    });

    expect(gatewayMocks.reportingGateway.getSalesReport).toHaveBeenCalledTimes(2);
    expect(gatewayMocks.reportingGateway.getSalesReport).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({ granularity: "DAY" }),
    );
    expect(gatewayMocks.reportingGateway.getSalesReport).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ granularity: "MONTH" }),
    );
    expect(gatewayMocks.catalogGateway.listItems).toHaveBeenCalledTimes(2);
    expect(gatewayMocks.catalogGateway.listItems).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ after: { name: "Brigadeiro", id: 2 } }),
    );
  });

  it("uses real report series in every chart and shows the category placeholder", async () => {
    const user = userEvent.setup();

    render(<DashboardPage />);

    await waitFor(() => {
      expect(screen.getByTestId("line-chart")).toHaveTextContent("20 jul");
      expect(screen.getByTestId("bar-chart")).toHaveTextContent("Brigadeiro real");
    });
    expect(screen.getByText("Mix por categoria indisponível")).toBeInTheDocument();
    expect(screen.queryByText("Bolo de Chocolate")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Receita" }));
    expect(screen.getByTestId("line-chart")).toHaveTextContent("jul 2026");
    expect(screen.getByTestId("line-chart")).toHaveTextContent("2450");

    await user.click(screen.getByRole("button", { name: "Vendas" }));
    expect(screen.getByTestId("bar-chart")).toHaveTextContent("jul 2026");
    expect(screen.getByTestId("bar-chart")).toHaveTextContent('"sales":12');

    await user.click(screen.getByRole("button", { name: "Produtos" }));
    expect(screen.getByTestId("bar-chart")).toHaveTextContent("Bolo real");

    expect(gatewayMocks.reportingGateway.getSalesReport).toHaveBeenCalledTimes(2);
    expect(gatewayMocks.reportingGateway.getCategoryMixReport).toHaveBeenCalledTimes(1);
  });
});
