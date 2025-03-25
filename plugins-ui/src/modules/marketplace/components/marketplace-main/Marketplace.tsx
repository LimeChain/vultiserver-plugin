import Button from "@/modules/core/components/ui/button/Button";
import PluginCard from "@/modules/plugin/components/plugin-card/PluginCard";
import { isSupportedChainType } from "@/modules/shared/wallet/wallet.utils";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import PricingService from '@/modules/policy/services/pricingService';
import { useNavigate } from "react-router-dom";
import "./Marketplace.css";
import MarketplaceFilters from "../marketplace-filters/MarketplaceFilters";
import { useState } from "react";
import { ViewFilter } from "../../models/marketplace";

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

const getAccountVault = async (): Promise<[string, { publicKeyEcdsa: string }]> => {
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

const Marketplace = () => {
  const navigate = useNavigate();
  const [view, setView] = useState<string>(getSavedView());

  const changeView = (view: ViewFilter) => {
    localStorage.setItem("view", view);
    setView(view);
  };

  const approveDcaPricingTerms = async () => {
    // dca pricing
    const pricingPolicy = JSON.stringify({
      type: 'PER_TX',
      amount: 0.1,
      metric: 'PERCENTAGE'
    })

    const [account, vault] = await getAccountVault();
    const signature = await sign(account, pricingPolicy)

    const pricing = await PricingService.createPricing({
      public_key: vault.publicKeyEcdsa, // TODO: what if not ecdsa?
      plugin_type: 'dca',
      signature,
      pricing: pricingPolicy
    })

    console.log('pricing', pricing)
  }

  return (
    <>
      <div className="only-section">
        <h2>Plugins Marketplace</h2>
        <MarketplaceFilters
          viewFilter={view as ViewFilter}
          onChange={changeView}
        />
        <section className="cards">
          {[1, 2, 3, 4, 5].map((_, index) => (
            <div className={view === "list" ? "list-card" : ""} key={index}>
              <PluginCard
                pluginType="dca" // todo remove hardcoding once we have the marketplace
                uiStyle={view as ViewFilter}
                id={index.toString()}
                title="DCA Plugin"
                description="The DCA Plugin allows you to dollar cost average into any supported token like Bitcoin. "
              />
            </div>
          ))}
        </section>

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
    </>
  );
};

export default Marketplace;
