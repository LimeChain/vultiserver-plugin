import { get, post } from "@/modules/core/services/httpService";
import { PluginPricing } from "../models/pluginPricing";

const getPublicKey = () => localStorage.getItem("publicKey");
const getPluginUrl = () => import.meta.env.VITE_PLUGIN_URL; // todo this is to be deleted and instead fetched with the policy from DB

const PricingService = {
  getPluginPricing: async (pluginType: string) => {
    try {
      const endpoint = `${getPluginUrl()}/plugin/${encodeURIComponent(pluginType)}/pricing-policy`;
      const pluginPricing = await get(endpoint, {
        headers: {
          plugin_type: pluginType,
          public_key: getPublicKey(),
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });
      return pluginPricing;
    } catch (error) {
      console.error("Error getting pricing:", error);
      throw error;
    }
  },

  createPricing: async (pricing: Omit<PluginPricing, "id">) => {
    try {
      const endpoint = `${getPluginUrl()}/plugin/${encodeURIComponent(pricing.plugin_type)}/pricing-policy`;
      const newPricing = await post(endpoint, pricing);
      return newPricing;
    } catch (error) {
      console.error("Error creating pricing:", error);
      throw error;
    }
  }
}

export default PricingService;
