import { post } from "@/modules/core/services/httpService";
import { PluginPricing } from "../models/pluginPricing";

const PricingService = {
  createPricing: async (pricing: Omit<PluginPricing, "id">) => {
    try {
      const endpoint = `/plugin/${encodeURIComponent(pricing.plugin_type)}/pricings`;
      const newPricing = await post(endpoint, pricing);
      return newPricing;
    } catch (error) {
      console.error("Error creating pricing:", error);
      throw error;
    }
  }
}

export default PricingService;
