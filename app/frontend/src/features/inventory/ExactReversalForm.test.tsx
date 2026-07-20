import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import ExactReversalForm from "./ExactReversalForm";

const gatewayMocks = vi.hoisted(() => ({
  reversalGateway: {
    postReversal: vi.fn(),
  },
}));

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

const postedReversal = {
  id: 90,
  idempotencyKey: "reversal-test",
  postingSequence: 8,
  targetDocumentId: 42,
  occurredOn: "2026-07-20",
  postedAtMs: 1_700_000_000_000,
  currencyCode: "BRL",
  currencyMinorDigits: 2,
  reasonCode: "EXACT_REVERSAL" as const,
  notes: "Correção de lançamento",
  lines: [
    {
      id: 91,
      lineOrder: 1,
      itemId: 10,
      direction: "IN" as const,
      quantityAtomic: 250,
      enteredUnitCode: "g",
      enteredPackagingName: null,
      conversionNumeratorAtomic: 1000,
      conversionDenominator: 1,
      inventoryValueMicro: 1_250_000,
      commercialTotalMinor: null,
      reversesLineId: 50,
      allocations: [],
    },
  ],
};

afterEach(() => {
  cleanup();
  gatewayMocks.reversalGateway.postReversal.mockReset();
});

describe("ExactReversalForm", () => {
  it("creates an explicitly confirmed exact reversal", async () => {
    const user = userEvent.setup();
    const onPosted = vi.fn();
    gatewayMocks.reversalGateway.postReversal.mockImplementation(async () => postedReversal);

    render(<ExactReversalForm onPosted={onPosted} />);

    const occurredOn = screen.getByLabelText("Data da reversão");
    await user.type(screen.getByLabelText("ID do documento"), "42");
    await user.type(screen.getByLabelText("Observações"), "Correção de lançamento");
    await user.click(screen.getByLabelText(/Confirmo que é uma correção de dados/));
    await user.click(screen.getByRole("button", { name: "Reverter documento" }));

    expect(gatewayMocks.reversalGateway.postReversal).toHaveBeenCalledWith({
      idempotencyKey: expect.stringMatching(/^reversal-/),
      targetDocumentId: 42,
      occurredOn: (occurredOn as HTMLInputElement).value,
      notes: "Correção de lançamento",
    });
    expect(
      await screen.findByText("Reversão #90 criada para o documento #42."),
    ).toBeInTheDocument();
    expect(screen.getByText("Item #10 · Entrada")).toBeInTheDocument();
    expect(screen.getByText(/reverte linha #50/)).toBeInTheDocument();
    expect(onPosted).toHaveBeenCalledWith(postedReversal);
  });

  it("shows backend eligibility failures", async () => {
    const user = userEvent.setup();
    gatewayMocks.reversalGateway.postReversal.mockImplementation(async () => {
      throw new Error("O documento não é o último lançamento dos itens afetados.");
    });

    render(<ExactReversalForm />);

    await user.type(screen.getByLabelText("ID do documento"), "41");
    await user.click(screen.getByLabelText(/Confirmo que é uma correção de dados/));
    await user.click(screen.getByRole("button", { name: "Reverter documento" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "O documento não é o último lançamento dos itens afetados.",
    );
  });
});
