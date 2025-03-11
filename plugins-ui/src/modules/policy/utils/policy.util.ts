import { PluginPolicy, Policy } from "../models/policy";

export const generatePolicy = (
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key:
      "023c3f2a75c5a90c9316b102a5d6be9a270a30ff4b171536a9246f02db4545468c", // TODO: get Vault's pub key
    plugin_type,
    active: true,
    policy: convertToStrings(policy),
    signature: "", // todo this should be implemented
  };
};

function convertToStrings<T extends Record<string, any>>(
  obj: T
): Record<string, string> {
  return Object.fromEntries(
    Object.entries(obj).map(([key, value]) => [
      key,
      typeof value === "object" && value !== null
        ? convertToStrings(value)
        : String(value),
    ])
  ) as Record<string, string>;
}
