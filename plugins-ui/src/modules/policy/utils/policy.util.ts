import { PluginPolicy, Policy } from "../models/policy";

export const generatePolicy = (
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key:
      "034850b75430c133b35a51b2b1540ad1e975e80bb996704e6e9aa67db64fe7c280", // TODO: get Vault's pub key
    plugin_type,
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
