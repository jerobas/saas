import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import InventoryPage from "./InventoryPage";

const gatewayMocks = vi.hoisted(() => ({
  adjustmentGateway: {
    postAdjustment: vi.fn(),
  },
  inventoryGateway: {
    listInventoryBalances: vi.fn(),
  },
  referenceDataGateway: {
    listMeasurementUnits: vi.fn(),
  },
}));

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

const balancePage = {
  items: [
    {
      itemId: 10,
      itemName: "Butter",
      baseUnitCode: "g",
      quantityAtomic: 1000,
      inventoryValueMicro: 5_000_000,
      updatedAtMs: 1_700_000_000_000,
      capabilities: { purchasable: true, producible: false, sellable: true },
      reorderQuantityAtomic: null,
    },
  ],
  next: null,
};

const gram = {
  code: "g",
  name: "gram",
  symbol: "g",
  dimension: "MASS",
  numeratorAtomic: 1000,
  denominator: 1,
  isItemBase: true,
  isSeeded: true,
};

describe("InventoryPage", () => {
  beforeEach(() => {
    gatewayMocks.inventoryGateway.listInventoryBalances.mockResolvedValue(balancePage);
    gatewayMocks.referenceDataGateway.listMeasurementUnits.mockResolvedValue([gram]);
    gatewayMocks.adjustmentGateway.postAdjustment.mockResolvedValue({
      id: 40,
      idempotencyKey: "adjustment-test",
      postingSequence: 2,
      occurredOn: "2026-07-16",
      postedAtMs: 1_700_000_000_001,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      reasonCode: "WASTE",
      lines: [
        {
          id: 50,
          lineOrder: 1,
          itemId: 10,
          direction: "OUT",
          quantityAtomic: 250,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1000,
          conversionDenominator: 1,
          inventoryValueMicro: 1_250_000,
          allocations: [{ id: 60, lotId: 70, quantityAtomic: 250 }],
        },
      ],
    });
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("posts a negative stock adjustment through the V2 gateway", async () => {
    const user = userEvent.setup();

    render(<InventoryPage />);

    expect((await screen.findAllByText("Butter")).length).toBeGreaterThan(0);
    await user.type(screen.getByLabelText("Quantidade atomica"), "250");
    await user.click(screen.getByRole("button", { name: "Postar ajuste" }));

    expect(gatewayMocks.adjustmentGateway.postAdjustment).toHaveBeenCalledWith(
      expect.objectContaining({
        reasonCode: "WASTE",
        lines: [
          expect.objectContaining({
            itemId: 10,
            direction: "OUT",
            quantityAtomic: 250,
            enteredUnitCode: "g",
            conversionNumeratorAtomic: 1000,
            conversionDenominator: 1,
          }),
        ],
      }),
    );
    expect(await screen.findByText("Ajuste #40 postado.")).toBeInTheDocument();
  });
});
