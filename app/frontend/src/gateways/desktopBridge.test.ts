import { afterEach, describe, expect, it, vi } from "vitest";
import { GetAllItems } from "./desktopBridge";

const originalBridge = window.go;

afterEach(() => {
  window.go = originalBridge;
});

describe("desktop bridge", () => {
  it("forwards calls to the Wails runtime", async () => {
    const getAllItems = vi.fn().mockResolvedValue([{ id: "item-1" }]);
    window.go = { service: { ItemService: { GetAllItems: getAllItems } } };

    await expect(GetAllItems()).resolves.toEqual([{ id: "item-1" }]);
    expect(getAllItems).toHaveBeenCalledOnce();
  });

  it("fails clearly when the desktop runtime is unavailable", async () => {
    window.go = undefined;

    await expect(GetAllItems()).rejects.toThrow(
      "Desktop bridge method ItemService.GetAllItems is unavailable.",
    );
  });
});
