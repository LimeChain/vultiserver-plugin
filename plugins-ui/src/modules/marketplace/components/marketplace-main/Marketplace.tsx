import Button from "@/modules/core/components/ui/button/Button";
import PluginCard from "@/modules/plugin/components/plugin-card/PluginCard";
import { useNavigate } from "react-router-dom";
import "./Marketplace.css";
import MarketplaceFilters from "../marketplace-filters/MarketplaceFilters";
import { useEffect, useState } from "react";
import { Plugin, ViewFilter } from "../../models/marketplace";
import Toast from "@/modules/core/components/ui/toast/Toast";
import MarketplaceService from "../../services/marketplaceService";

const getSavedView = (): string => {
  return localStorage.getItem("view") || "grid";
};

const Marketplace = () => {
  const navigate = useNavigate();
  const [view, setView] = useState<string>(getSavedView());

  const changeView = (view: ViewFilter) => {
    localStorage.setItem("view", view);
    setView(view);
  };

  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  const [plugins, setPlugins] = useState<Plugin[] | null>(null);

  useEffect(() => {
    const fetchPlugins = async (): Promise<void> => {
      try {
        const fetchedPlugins = await MarketplaceService.getPlugins();
        console.log("fetchedPlugins", fetchedPlugins);
        setPlugins(fetchedPlugins);
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
  }, []);

  return (
    <>
      {plugins && (
        <div className="only-section">
          <h2>Plugins Marketplace</h2>
          <MarketplaceFilters
            viewFilter={view as ViewFilter}
            onChange={changeView}
          />
          <section className="cards">
            {plugins.map((plugin) => (
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

          <Button
            size="small"
            type="button"
            styleType="primary"
            onClick={() => navigate(`/plugin-detail/1`)}
          >
            Open Detail view
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
