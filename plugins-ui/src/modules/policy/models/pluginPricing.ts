export type PluginPricing = {
  id: string;
  public_key: string;
  plugin_type: string;
  is_ecdsa: boolean,
  chain_code_hex: string,
  derive_path: string,
  signature: string;
  pricing: string;
}
