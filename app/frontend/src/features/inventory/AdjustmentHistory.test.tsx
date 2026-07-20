import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import AdjustmentHistory from "./AdjustmentHistory";

const gatewayMocks = vi.hoisted(() => ({
  adjustmentGateway: {
    listAdjustments: vi.fn(),
  },
}));

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

const adjustmentPage = {
  items: [
    {
      id: 41,
      idempotencyKey: "adjustment-1",
      postingSequence: 2,
      occurredOn: "2026-07-16",
      postedAtMs: 1_700_000_000_100,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      reasonCode: "WASTE" as const,
      notes: "Perda no preparo",
      lines: [
        {
          id: 51,
          lineOrder: 1,
          itemId: 10,
          direction: "OUT" as const,
          quantityAtomic: 250,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          inventoryValueMicro: 1_250_000,
          allocations: [{ id: 70, lotId: 60, quantityAtomic: 250 }],
        },
      ],
    },
  ],
  next: null,
};

afterEach(() => {
  cleanup();
  gatewayMocks.adjustmentGateway.listAdjustments.mockReset();
});

describe("AdjustmentHistory", () => {
  it("shows recent documents and selects one for exact reversal", async () => {
    const user = userEvent.setup();
    const onReverse = vi.fn();
    gatewayMocks.adjustmentGateway.listAdjustments.mockImplementation(async () => adjustmentPage);

    render(
      <AdjustmentHistory
        items={[{ itemId: 10, itemName: "Manteiga" }]}
        refreshKey={0}
        onReverse={onReverse}
      />,
    );

    expect(await screen.findByText("Documento #41")).toBeInTheDocument();
    expect(screen.getByText("Perda")).toBeInTheDocument();
    expect(screen.getByText(/Manteiga · saída de 250 unidades atômicas/)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Reverter ajuste #41" }));
    expect(onReverse).toHaveBeenCalledWith(41);
  });
});
