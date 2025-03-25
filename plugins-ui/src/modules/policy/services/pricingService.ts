import { post } from "@/modules/core/services/httpService";
import { PluginPricing } from "../models/pluginPricing";

const getPluginUrl = () => import.meta.env.VITE_BASE_URL; // todo this is to be deleted and instead fetched with the policy from DB

const PricingService = {
  createPricing: async (pricing: Omit<PluginPricing, "id">) => {
    try {
      const endpoint = `${getPluginUrl()}/plugin/${encodeURIComponent(pricing.plugin_type)}/pricings`;
      const newPricing = await post(endpoint, pricing);
      return newPricing;
    } catch (error) {
      console.error("Error creating pricing:", error);
      throw error;
    }
  }
}

export default PricingService;
