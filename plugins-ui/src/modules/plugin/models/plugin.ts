export type Plugin = {
  id: string;
  type: string;
  title: string;
  description: string;
  metadata: {};
  server_endpoint: string;
  pricing_id: string;
  ratings: PluginRatings[];
};

export type PluginRatings = {
  rating: number;
  count: number;
};

export type PluginFee = {
  id: string;
  amount: number;
  metric: "FIXED" | "PERCENTAGE";
  type: "FREE" | "SINGLE" | "RECURRING" | "PER_TX";
  frequency?: "ANNUAL" | "MONTHLY" | "WEEKLY";
};
