import { PluginFee } from "../models/plugin";

const pricingMetric = {
  FIXED: "USDC",
  PERCENTAGE: "%",
};

const pricingType = {
  FREE: "free",
  SINGLE: "per installation",
  RECURRING: "per",
  PER_TX: "per trade",
};

const pricingFrequency = {
  ANNUAL: "year",
  MONTHLY: "month",
  WEEKLY: "week",
};

export const getPluginFee = (pricing: PluginFee): string => {
  return `Plugin fee: ${pricing.amount}${pricingMetric[pricing.metric]} ${pricingType[pricing.type]} ${pricing.frequency ? pricingFrequency[pricing.frequency] : ""}`;
};
