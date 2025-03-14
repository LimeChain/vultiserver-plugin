import { PluginPolicy, Policy } from "@/modules/policy/models/policy";
import { supportedTokens } from "@/modules/shared/data/tokens";
import { ethers } from "ethers";

export const mapData = (
  pluginPolicy: PluginPolicy
): { [key: string]: unknown } => {
  console.log("pluginPolicy", pluginPolicy);

  const weiAmount = ethers
    .formatUnits(
      pluginPolicy.policy.total_amount as string,
      supportedTokens[pluginPolicy.policy.source_token_id as string].decimals
    )
    .toString();
  return {
    policyId: pluginPolicy.id,
    pair: [
      pluginPolicy.policy.source_token_id as string,
      pluginPolicy.policy.destination_token_id as string,
    ],
    sell: `${weiAmount} ${supportedTokens[pluginPolicy.policy.source_token_id as string].name}`,
    orders: pluginPolicy.policy.total_orders as string,
    toBuy:
      supportedTokens[pluginPolicy.policy.destination_token_id as string].name,
    orderInterval: `${(pluginPolicy.policy.schedule as Policy).interval} ${(pluginPolicy.policy.schedule as Policy).frequency}`,
    status: pluginPolicy.active,
  };
};
