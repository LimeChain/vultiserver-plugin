import { mockedDCAPolicy } from "@/__tests__/utils/global-mocks";
import Policy from "@/modules/policy/components/policy-main/Policy";
import { PluginPolicy } from "@/modules/policy/models/policy";
import { RJSFSchema } from "@rjsf/utils";
import { render } from "@testing-library/react";
import { describe, it, vi } from "vitest";

const hoisted = vi.hoisted(() => ({
  mockOnSubmitCallback: vi.fn(),
  mockAddPolicy: vi.fn(() => true),
  mockUpdatePolicy: vi.fn(() => true),
}));

const mockPolicySchemaMap: Map<string, RJSFSchema> = new Map();
mockPolicySchemaMap.set("dca", mockedDCAPolicy);

const mockPolicyMap: Map<string, PluginPolicy> = new Map();

vi.mock("@/modules/policy/context/PolicyProvider", async (importActual) => ({
  ...(await importActual()),
  usePolicies: vi.fn(() => ({
    addPolicy: hoisted.mockAddPolicy,
    updatePolicy: hoisted.mockUpdatePolicy,
    policySchemaMap: mockPolicySchemaMap,
    pluginType: "dca",
    policyMap: mockPolicyMap,
  })),
}));

describe("Policy", () => {
  it("should render all components", async () => {
    localStorage.setItem("authToken", "test");
    const { findByTestId } = render(<Policy />);

    await findByTestId("policy-form-wrapper");
    await findByTestId("policy-table-wrapper");
  });
});
