import Button from "@/modules/core/components/ui/button/Button";
import PluginCard from "@/modules/plugin/components/plugin-card/PluginCard";
import { isSupportedChainType } from "@/modules/shared/wallet/wallet.utils";
import VulticonnectWalletService, { IVulticonnectVault }  from "@/modules/shared/wallet/vulticonnectWalletService";
import PricingService from '@/modules/policy/services/pricingService';
import { useNavigate } from "react-router-dom";
import "./Marketplace.css";
import MarketplaceFilters from "../marketplace-filters/MarketplaceFilters";
import { useEffect, useState } from "react";
import { PluginMap, ViewFilter } from "../../models/marketplace";
import Toast from "@/modules/core/components/ui/toast/Toast";
import MarketplaceService from "../../services/marketplaceService";
import Pagination from "@/modules/core/components/ui/pagination/Pagination";

const getSavedView = (): string => {
  return localStorage.getItem("view") || "grid";
};

const toHex = (str: string): string => {
  return (
    "0x" +
    Array.from(str)
      .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
      .join("")
  );
};

const getAccountVault = async (): Promise<[string, IVulticonnectVault]> => {
  const chain = localStorage.getItem("chain") as string;

  let accounts = [];
  if (!isSupportedChainType(chain)) {
    throw new Error('Chain not supported');
  }

  if (chain === "ethereum") {
    accounts = await VulticonnectWalletService.getConnectedEthAccounts();
  }

  if (!accounts || accounts.length === 0) {
    throw new Error("Need to connect to wallet");
  }

  const vaults = await window.vultisig?.getVaults();
  if (!vaults || vaults.length === 0) {
    throw new Error("No vaults found");
  }

  const [account] = accounts;
  const [vault] = vaults;

  return [account, vault];
}

const sign = async (account: string, content: string): Promise<string> => {
  const hexMessage = toHex(content);

  const signature = await VulticonnectWalletService.signCustomMessage(
    hexMessage,
    account
  );

  return signature;
}
const ITEMS_PER_PAGE = 6;

const Marketplace = () => {
  const navigate = useNavigate();
  const [view, setView] = useState<string>(getSavedView());

  const [currentPage, setCurrentPage] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  const changeView = (view: ViewFilter) => {
    localStorage.setItem("view", view);
    setView(view);
  };

  const approveDcaPricingTerms = async () => {
    // dca pricing
    const pricingPolicy = {
      type: 'PER_TX',
      amount: 0.1,
      metric: 'PERCENTAGE'
    }

    const [account, vault] = await getAccountVault();
    const signature = await sign(account, JSON.stringify(pricingPolicy));

    const pricing = await PricingService.createPricing({
      public_key: vault.publicKeyEcdsa,
      plugin_type: 'dca',
      is_ecdsa: true,
      chain_code_hex: vault.hexChainCode,
      derive_path: "m/44'/60'/0'/0/0", // TODO: should be per chain
      signature,
      pricing: pricingPolicy
    })

    console.log('pricing', pricing)
  }
  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  const [pluginsMap, setPlugins] = useState<PluginMap | null>(null);

  useEffect(() => {
    const fetchPlugins = async (): Promise<void> => {
      try {
        const fetchedPlugins = await MarketplaceService.getPlugins(
          currentPage > 1 ? (currentPage - 1) * ITEMS_PER_PAGE : 0,
          ITEMS_PER_PAGE
        );
        setPlugins(fetchedPlugins);
        setTotalPages(Math.ceil(fetchedPlugins.total_count / ITEMS_PER_PAGE));

        if (
          fetchedPlugins.total_count / ITEMS_PER_PAGE > 1 &&
          currentPage === 0
        ) {
          setCurrentPage(1);
        }
      } catch (error: any) {
        console.error("Failed to get plugins:", error.message);
        setToast({
          message: "Failed to get plugins",
          error: error.error,
          type: "error",
        });
      }
    };

    fetchPlugins();
  }, [currentPage]);

  const onCurrentPageChange = (page: number): void => {
    setCurrentPage(page);
  };

  return (
    <>
      {pluginsMap && (
        <div className="only-section">
          <h2>Plugins Marketplace</h2>
          <MarketplaceFilters
            viewFilter={view as ViewFilter}
            onChange={changeView}
          />
          <section className="cards">
            {pluginsMap.plugins?.map((plugin) => (
              <div
                className={view === "list" ? "list-card" : ""}
                key={plugin.id}
              >
                <PluginCard
                  pluginType={plugin.type}
                  uiStyle={view as ViewFilter}
                  id={plugin.id}
                  title={plugin.title}
                  description={plugin.description}
                />
              </div>
            ))}
          </section>

          {totalPages > 1 && (
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={onCurrentPageChange}
            />
          )}

          <Button
            size="small"
            type="button"
            styleType="primary"
            onClick={() => navigate(`/plugin-detail/1`)}
          >
            Open Detail view
          </Button>

          <Button
            size="small"
            type="button"
            styleType="primary"
            onClick={approveDcaPricingTerms}
          >
            Approve DCA pricing terms
          </Button>
        </div>
      )}

      {toast && (
        <Toast
          title={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </>
  );
};

export default Marketplace;
