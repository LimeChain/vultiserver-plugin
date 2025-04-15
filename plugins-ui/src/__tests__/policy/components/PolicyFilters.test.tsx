import PolicyFilters from "@/modules/policy/components/policy-filters/PolicyFilters";
import { describe, expect, it, vi } from "vitest";
import { render, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

describe("PolicyFilters", () => {
  const mockOnChange = vi.fn();

  it("should render closed select with default option All", async () => {
    const { findByTestId } = render(
      <PolicyFilters onFiltersChange={mockOnChange} />
    );
    await findByTestId("select-box-wrapper");
    const selectedElement = await findByTestId("select-box-selected");
    expect(selectedElement.innerHTML).includes("all");
  });

  it("should change filters & close dropdown upon click", async () => {
    const { findByTestId, findAllByTestId } = render(
      <PolicyFilters onFiltersChange={mockOnChange} />
    );

    const selectBoxTrigger = await findByTestId("select-box-trigger");
    await userEvent.click(selectBoxTrigger);
    await waitFor(async () => {
      const selectBoxOptions = await findAllByTestId("select-box-option");
      await userEvent.click(selectBoxOptions[1]);
      expect(mockOnChange).toBeCalledWith([
        {
          id: "status",
          value: true,
        },
      ]);
    });
  });
});
