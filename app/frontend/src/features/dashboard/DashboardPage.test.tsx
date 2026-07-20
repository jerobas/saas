import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
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

describe("DashboardPage", () => {
  beforeEach(() => {
    gatewayMocks.reportingGateway.getSalesReport.mockResolvedValue({
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      totalSalesCount: 12,
      commercialTotalMinor: 245_000,
      growthBasisPoints: 1_250,
    });
    gatewayMocks.reportingGateway.getInventoryReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getPurchaseReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getProductionReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getAdjustmentReport.mockResolvedValue({});
    gatewayMocks.reportingGateway.getCategoryMixReport.mockResolvedValue({});
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

    expect(gatewayMocks.reportingGateway.getSalesReport).toHaveBeenCalledTimes(1);
    expect(gatewayMocks.catalogGateway.listItems).toHaveBeenCalledTimes(2);
    expect(gatewayMocks.catalogGateway.listItems).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ after: { name: "Brigadeiro", id: 2 } }),
    );
  });
});
