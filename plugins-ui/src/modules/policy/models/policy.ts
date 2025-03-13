export type Policy = {
  [key: string]: string | number | null | undefined | Policy;
};

export type PluginPolicy = {
  id: string;
  public_key: string;
  is_ecdsa: boolean;
  chain_code_hex: string;
  derive_path: string;
  plugin_id: string;
  plugin_version: string;
  policy_version: string;
  plugin_type: string;
  signature: string;
  policy: Policy;
  active: boolean;
};

export type PolicyTransactionHistory = {
  id: string;
  updated_at: string;
  status: string;
};
