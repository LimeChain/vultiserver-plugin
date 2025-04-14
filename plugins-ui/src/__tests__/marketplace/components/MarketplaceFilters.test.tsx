import MarketplaceFilters from "@/modules/marketplace/components/marketplace-filters/MarketplaceFilters";
import { render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

describe("MarketplaceFilters", async () => {
  it("should call onChange with 'grid'", async () => {
    const mockOnChange = vi.fn();
    const { findByTestId } = render(
      <MarketplaceFilters viewFilter="list" onChange={mockOnChange} />
    );
    const gridButton = await findByTestId("marketplace-filters-grid");
    await userEvent.click(gridButton);
    expect(mockOnChange).toBeCalledWith("grid");
  });
  it("should call onChange with 'list'", async () => {
    const mockOnChange = vi.fn();
    const { findByTestId } = render(
      <MarketplaceFilters viewFilter="grid" onChange={mockOnChange} />
    );
    const listButton = await findByTestId("marketplace-filters-list");
    await userEvent.click(listButton);
    expect(mockOnChange).toBeCalledWith("list");
  });
});
