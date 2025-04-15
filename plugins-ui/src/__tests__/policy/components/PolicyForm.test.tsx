import {
  mockedDCAPolicy,
  mockPluginPolicy,
} from "@/__tests__/utils/global-mocks";
import PolicyForm from "@/modules/policy/components/policy-form/PolicyForm";
import { RJSFSchema } from "@rjsf/utils";
import { render, waitFor, queryAllByAttribute } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

const hoisted = vi.hoisted(() => ({
  mockOnSubmitCallback: vi.fn(),
  mockAddPolicy: vi.fn(() => true),
  mockUpdatePolicy: vi.fn(() => true),
}));

const mockPolicySchemaMap: Map<string, RJSFSchema> = new Map();
mockPolicySchemaMap.set("dca", mockedDCAPolicy);

vi.mock("@/modules/policy/context/PolicyProvider", async (importActual) => ({
  ...(await importActual()),
  usePolicies: vi.fn(() => ({
    addPolicy: hoisted.mockAddPolicy,
    updatePolicy: hoisted.mockUpdatePolicy,
    policySchemaMap: mockPolicySchemaMap,
    pluginType: "dca",
  })),
}));

describe("PolicyForm", () => {
  it("should visualize form", async () => {
    const { container } = render(
      <PolicyForm onSubmitCallback={hoisted.mockOnSubmitCallback} />
    );
    const formElement = container.querySelector("form");
    expect(formElement).toBeInTheDocument();
  });
  it("should call addPolicy", async () => {
    const { findByText, findByTestId, container } = render(
      <PolicyForm onSubmitCallback={hoisted.mockOnSubmitCallback} />
    );
    const weiInput = await findByTestId("wei-converter");
    await userEvent.type(weiInput, "20");
    const elements = queryAllByAttribute("class", container, "form-control", {
      exact: false,
    });
    const inputs = elements.filter((el) => {
      const label = el.getAttribute("label");
      if (label === "Every" || label === "Over (orders)") {
        return true;
      }
      return false;
    });

    await userEvent.type(inputs[0], "20");
    await userEvent.type(inputs[1], "20");

    await waitFor(async () => {
      const submitBtn = await findByText("Save policy", { exact: false });
      await userEvent.click(submitBtn);
      expect(hoisted.mockAddPolicy).toBeCalled();
      expect(hoisted.mockOnSubmitCallback).toBeCalled();
    });
  });
  it("should call addPolicy and handle error", async () => {
    hoisted.mockAddPolicy.mockRejectedValueOnce(new Error("From test"));
    const consoleSpy = vi.spyOn(console, "error");
    const { findByText, findByTestId, container } = render(
      <PolicyForm onSubmitCallback={hoisted.mockOnSubmitCallback} />
    );
    const weiInput = await findByTestId("wei-converter");
    await userEvent.type(weiInput, "20");
    const elements = queryAllByAttribute("class", container, "form-control", {
      exact: false,
    });
    const inputs = elements.filter((el) => {
      const label = el.getAttribute("label");
      if (label === "Every" || label === "Over (orders)") {
        return true;
      }
      return false;
    });

    await userEvent.type(inputs[0], "20");
    await userEvent.type(inputs[1], "20");

    await waitFor(async () => {
      const submitBtn = await findByText("Save policy", { exact: false });
      await userEvent.click(submitBtn);
      expect(hoisted.mockAddPolicy).toBeCalled();
      expect(consoleSpy).toBeCalled();
    });
  });
  it("should call updatePolicy", async () => {
    const { findByText } = render(
      <PolicyForm
        onSubmitCallback={hoisted.mockOnSubmitCallback}
        data={mockPluginPolicy}
      />
    );

    await waitFor(async () => {
      const submitBtn = await findByText("Save policy", { exact: false });
      await userEvent.click(submitBtn);
      expect(hoisted.mockUpdatePolicy).toBeCalled();
      expect(hoisted.mockOnSubmitCallback).toBeCalled();
    });
  });
  it("should call updatePolicy and handle error", async () => {
    hoisted.mockUpdatePolicy.mockRejectedValueOnce(new Error("From test"));
    const consoleSpy = vi.spyOn(console, "error");
    const { findByText } = render(
      <PolicyForm
        onSubmitCallback={hoisted.mockOnSubmitCallback}
        data={mockPluginPolicy}
      />
    );

    await waitFor(async () => {
      const submitBtn = await findByText("Save policy", { exact: false });
      await userEvent.click(submitBtn);
      expect(hoisted.mockUpdatePolicy).toBeCalled();
      expect(consoleSpy).toBeCalled();
    });
  });
});
