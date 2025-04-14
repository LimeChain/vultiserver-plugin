import PluginCard from "@/modules/plugin/components/plugin-card/PluginCard";
import "./Marketplace.css";
import MarketplaceFilters from "../marketplace-filters/MarketplaceFilters";
import { useEffect, useState } from "react";
import { PluginMap, ViewFilter } from "../../models/marketplace";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import Pagination from "@/modules/core/components/ui/pagination/Pagination";
import { publish } from "@/utils/eventBus";

const getSavedView = (): string => {
  return localStorage.getItem("view") || "grid";
};

const ITEMS_PER_PAGE = 6;

const Marketplace = () => {
  const [view, setView] = useState<string>(getSavedView());

  const [currentPage, setCurrentPage] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  const changeView = (view: ViewFilter) => {
    localStorage.setItem("view", view);
    setView(view);
  };

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
      } catch (error) {
        if (error instanceof Error) {
          console.error("Failed to get plugins:", error.message);
          publish("onToast", {
            type: "error",
            message: "Failed to get plugins",
          });
        }
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
        <div className="only-section" data-testid="marketplace-wrapper">
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
                data-testid="marketplace-plugin-card"
              >
                <PluginCard
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
        </div>
      )}
    </>
  );
};

export default Marketplace;
