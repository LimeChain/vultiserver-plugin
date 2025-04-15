import {
  PolicyContext,
  PolicyContextType,
} from "@/modules/policy/context/PolicyProvider";
import { PluginPolicy } from "@/modules/policy/models/policy";
import { ReactNode } from "react";
import { mockPolicyFunctions } from "./global-mocks";
import { render } from "@testing-library/react";

export const policyContextRender = (
  ui: ReactNode,
  policyMap: Map<string, PluginPolicy>
) => {
  const mockValue: PolicyContextType = {
    pluginType: "pluginType",
    policyMap: policyMap,
    policySchemaMap: new Map(),
    addPolicy: mockPolicyFunctions.mockAddPolicy,
    updatePolicy: mockPolicyFunctions.mockUpdatePolicy,
    removePolicy: mockPolicyFunctions.mockRemovePolicy,
    getPolicyHistory: mockPolicyFunctions.mockGetPolicyHistory,
  };

  return render(
    <PolicyContext.Provider value={mockValue}>{ui}</PolicyContext.Provider>
  );
};
