export type ViewFilter = "grid" | "list";

export type PluginType = {
  id: string;
  type: string;
  title: string;
  description: string;
  metadata: {};
  server_endpoint: string;
  pricing_id: string;
};
