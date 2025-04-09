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
