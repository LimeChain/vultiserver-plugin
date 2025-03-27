export type PluginPricing = {
  id: string;
  public_key: string;
  plugin_type: string;
  is_ecdsa: boolean,
  chain_code_hex: string,
  derive_path: string,
  signature: string;
  pricing: PluginPricingPolicy;
}

type PluginPricingPolicy = {
  type: string;
  frequency?: string;
  amount: number;
  metric: string;
}
