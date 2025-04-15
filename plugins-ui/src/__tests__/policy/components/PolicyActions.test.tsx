import { describe, it, expect } from "vitest";
import { waitFor } from "@testing-library/react";
import PolicyActions from "@/modules/policy/components/policy-actions/PolicyActions";
import "@testing-library/jest-dom";
import { PluginPolicy } from "@/modules/policy/models/policy";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { USDC_TOKEN, WETH_TOKEN } from "@/modules/shared/data/tokens";
import userEvent from "@testing-library/user-event";
import { mockPolicyFunctions } from "@/__tests__/utils/global-mocks";
import { policyContextRender } from "@/__tests__/utils/custom-renders";

const mockPolicyActive: PluginPolicy = generatePolicy(
  "",
  "",
  "pluginType",
  "1",
  {
    source_token_id: WETH_TOKEN,
    destination_token_id: USDC_TOKEN,
  }
);

const presetPolicies = new Map();
presetPolicies.set(mockPolicyActive.id, mockPolicyActive);

describe("PolicyActions", () => {
  it("should render Pause/Pause button for each policy depending on its state", async () => {
    const { findByTestId } = policyContextRender(
      <PolicyActions policyId="1" />,
      presetPolicies
    );

    const pauseButton = await findByTestId("policy-actions-update-btn");

    await userEvent.click(pauseButton);

    await waitFor(() => {
      expect(mockPolicyFunctions.mockUpdatePolicy).toBeCalledWith(
        presetPolicies.get(mockPolicyActive.id)
      );
    });
  });

  it("should open edit modal for policy", async () => {
    const { findByTestId, getByRole } = policyContextRender(
      <PolicyActions policyId="1" />,
      presetPolicies
    );

    const editButton = await findByTestId("policy-actions-edit-btn");

    await userEvent.click(editButton);

    const modal = getByRole("dialog");
    expect(modal).toBeInTheDocument();
  });

  it("should open transaction history for policy", async () => {
    const { findByTestId, getByRole, getByText } = policyContextRender(
      <PolicyActions policyId="1" />,
      presetPolicies
    );

    const historyButton = await findByTestId("policy-actions-history-btn");

    await userEvent.click(historyButton);

    const modal = getByRole("dialog");
    expect(modal).toBeInTheDocument();
    const modalHeader = getByText("Transaction History");
    expect(modalHeader).toBeInTheDocument();
  });

  it("should delete policy", async () => {
    const { findByTestId } = policyContextRender(
      <PolicyActions policyId="1" />,
      presetPolicies
    );

    const removeBtn = await findByTestId("policy-actions-delete-btn");
    await userEvent.click(removeBtn);
    expect(mockPolicyFunctions.mockRemovePolicy).toBeCalledWith("1");
  });
});
