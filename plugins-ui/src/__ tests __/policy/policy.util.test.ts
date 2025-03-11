import { Policy } from "@/modules/policy/models/policy";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { describe, expect, it } from "vitest";

describe("generatePolicy", () => {
  it("should render Pause/Pause button for each policy depending on its state", () => {
    const policyData: Policy = {
      someNumberInput: 5,
      someBooleanInput: false,
      someStringInput: "text",
      someNestedInput: {
        someNumberInput: 5,
        someBooleanInput: false,
        someStringInput: "text",
      },
      someNullInput: null,
      someUndefinedInput: undefined,
    };

    const result = generatePolicy("pluginType", "", policyData);
    expect(result).toStrictEqual({
      active: true,
      id: "",
      plugin_type: "pluginType",
      policy: {
        someNumberInput: "5",
        someBooleanInput: "false",
        someStringInput: "text",
        someNestedInput: {
          someNumberInput: "5",
          someBooleanInput: "false",
          someStringInput: "text",
        },
        someNullInput: "null",
        someUndefinedInput: "undefined",
      },
      public_key:
        "023c3f2a75c5a90c9316b102a5d6be9a270a30ff4b171536a9246f02db4545468c",
      signature: "",
    });
  });
});
