import { get } from "@/modules/core/services/httpService";

const getMarketplaceUrl = () => import.meta.env.VITE_MARKETPLACE_URL;

const MarketplaceService = {
  /**
   * Get plugins from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched plugins.
   */
  getPlugins: async (): Promise<any> => {
    try {
      const endpoint = `${getMarketplaceUrl()}/plugins`;
      const plugins = await get(endpoint);
      return plugins;
    } catch (error) {
      console.error("Error getting plugins:", error);
      throw error;
    }
  },
};

export default MarketplaceService;
