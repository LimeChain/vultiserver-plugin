import React, { createContext, useContext, useEffect, useState } from "react";
import {
  PluginPolicy,
  PolicySchema,
  PolicyTransactionHistory,
} from "../models/policy";
import PolicyService from "../services/policyService";
import {
  derivePathMap,
  isSupportedChainType,
  toHex,
} from "@/modules/shared/wallet/wallet.utils";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import { useParams } from "react-router-dom";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import { publish } from "@/utils/eventBus";

export interface PolicyContextType {
  pluginType: string;
  policyMap: Map<string, PluginPolicy>;
  policySchemaMap: Map<string, PolicySchema>;
  addPolicy: (policy: PluginPolicy) => Promise<boolean>;
  updatePolicy: (policy: PluginPolicy) => Promise<boolean>;
  removePolicy: (policyId: string) => Promise<void>;
  getPolicyHistory: (policyId: string) => Promise<PolicyTransactionHistory[]>;
}

export const PolicyContext = createContext<PolicyContextType | undefined>(
  undefined
);

export const PolicyProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [policyMap, setPolicyMap] = useState(new Map<string, PluginPolicy>());
  const [policySchemaMap, setPolicySchemaMap] = useState(
    new Map<string, PolicySchema>()
  );

  const { pluginId } = useParams();
  const [pluginType, setPluginType] = useState("");
  const [serverEndpoint, setServerEndpoint] = useState("");
  const [authToken, setAuthToken] = useState(
    localStorage.getItem("authToken") || ""
  );

  useEffect(() => {
    const handleStorageChange = () => {
      setAuthToken(localStorage.getItem("authToken") || "");
    };

    // Listen for storage changes
    window.addEventListener("storage", handleStorageChange);

    const fetchPlugin = async (): Promise<void> => {
      if (!pluginId) return;

      try {
        const fetchedPlugin = await MarketplaceService.getPlugin(pluginId);

        if (fetchedPlugin) {
          setPluginType(fetchedPlugin.type);
          setServerEndpoint(fetchedPlugin.server_endpoint);
          const fetchPolicies = async (): Promise<void> => {
            try {
              const fetchedPolicies = await MarketplaceService.getPolicies(
                fetchedPlugin.type
              );

              const constructPolicyMap: Map<string, PluginPolicy> = new Map(
                fetchedPolicies?.map((p: PluginPolicy) => [p.id, p]) // Convert the array into [key, value] pairs
              );

              setPolicyMap(constructPolicyMap);
            } catch (error) {
              if (error instanceof Error) {
                console.error("Failed to get policies:", error.message);
                publish("onToast", {
                  message: error.message || "Failed to get policies",
                  type: "error",
                });
              }
            }
          };

          fetchPolicies();

          const fetchPolicySchema = async (
            pluginType: string
          ): Promise<unknown> => {
            if (policySchemaMap.has(pluginType)) {
              return Promise.resolve(policySchemaMap.get(pluginType));
            }

            try {
              const fetchedSchemas = await PolicyService.getPolicySchema(
                fetchedPlugin.server_endpoint,
                fetchedPlugin.type
              );

              setPolicySchemaMap((prev) =>
                new Map(prev).set(fetchedPlugin.type, fetchedSchemas)
              );

              return Promise.resolve(fetchedSchemas);
            } catch (error) {
              if (error instanceof Error) {
                console.error("Failed to fetch policy schema:", error.message);
                publish("onToast", {
                  message: error.message || "Failed to fetch policy schema",
                  type: "error",
                });
              }

              return Promise.resolve(null);
            }
          };

          fetchPolicySchema(fetchedPlugin.type);
        }
      } catch (error) {
        if (error instanceof Error) {
          console.error("Plugin not found:", error.message);
          publish("onToast", {
            message: "Plugin not found",
            type: "error",
          });
        }

        return;
      }
    };

    fetchPlugin();

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [authToken]);

  const addPolicy = async (policy: PluginPolicy): Promise<boolean> => {
    try {
      const signature = await signPolicy(policy);
      if (signature && typeof signature === "string") {
        policy.signature = signature;
        const newPolicy = await PolicyService.createPolicy(
          serverEndpoint,
          policy
        );
        setPolicyMap((prev) => new Map(prev).set(newPolicy.id, newPolicy));
        publish("onToast", {
          message: "Policy created successfully!",
          type: "success",
        });
        return Promise.resolve(true);
      }
      return Promise.resolve(false);
    } catch (error) {
      if (error instanceof Error) {
        console.error("Failed to create policy:", error.message);
        publish("onToast", {
          message: error.message || "Failed to create policy",
          type: "error",
        });
      }
      return Promise.resolve(false);
    }
  };

  const updatePolicy = async (policy: PluginPolicy): Promise<boolean> => {
    try {
      const signature = await signPolicy(policy);

      if (signature && typeof signature === "string") {
        policy.signature = signature;
        const updatedPolicy = await PolicyService.updatePolicy(
          serverEndpoint,
          policy
        );

        setPolicyMap((prev) =>
          new Map(prev).set(updatedPolicy.id, updatedPolicy)
        );
        publish("onToast", {
          message: "Policy updated successfully!",
          type: "success",
        });
        return Promise.resolve(true);
      }

      return Promise.resolve(false);
    } catch (error) {
      if (error instanceof Error) {
        console.error("Failed to update policy:", error.message, error);
        publish("onToast", {
          message: error.message || "Failed to update policy",
          type: "error",
        });
      }

      return Promise.resolve(false);
    }
  };

  const removePolicy = async (policyId: string): Promise<void> => {
    const policy = policyMap.get(policyId);

    if (!policy) return;

    try {
      const signature = await signPolicy(policy);
      if (signature && typeof signature === "string") {
        await PolicyService.deletePolicy(serverEndpoint, policyId, signature);

        setPolicyMap((prev) => {
          const updatedPolicyMap = new Map(prev);
          updatedPolicyMap.delete(policyId);

          return updatedPolicyMap;
        });
        publish("onToast", {
          message: "Policy deleted successfully!",
          type: "success",
        });
      }
    } catch (error) {
      if (error instanceof Error) {
        console.error("Failed to delete policy:", error);
        publish("onToast", {
          message: error.message,
          type: "error",
        });
      }
    }
  };

  const signPolicy = async (policy: PluginPolicy): Promise<string> => {
    const chain = localStorage.getItem("chain") as string;

    if (isSupportedChainType(chain)) {
      let accounts = [];
      if (chain === "ethereum") {
        accounts = await VulticonnectWalletService.getConnectedEthAccounts();
      }

      if (!accounts || accounts.length === 0) {
        throw new Error("Need to connect to wallet");
      }

      const vaults = await VulticonnectWalletService.getVaults();

      policy.public_key = vaults[0].publicKeyEcdsa;
      policy.signature = "";
      policy.is_ecdsa = true;
      policy.chain_code_hex = vaults[0].hexChainCode;
      policy.derive_path = derivePathMap[chain];
      const serializedPolicy = JSON.stringify(policy);
      const hexMessage = toHex(serializedPolicy);

      const signature = await VulticonnectWalletService.signCustomMessage(
        hexMessage,
        accounts[0]
      );

      console.log("Public key ecdsa: ", policy.public_key);
      console.log("Chain code hex: ", policy.chain_code_hex);
      console.log("Derive path: ", policy.derive_path);
      console.log("Hex message: ", hexMessage);
      console.log("Account[0]: ", accounts[0]);
      console.log("Signature: ", signature);

      return signature;
    }
    return "";
  };

  const getPolicyHistory = async (
    policyId: string
  ): Promise<PolicyTransactionHistory[]> => {
    try {
      const history =
        await MarketplaceService.getPolicyTransactionHistory(policyId);
      return history;
    } catch (error) {
      if (error instanceof Error) {
        console.error("Failed to get policy history:", error);
        publish("onToast", {
          message: error.message,
          type: "error",
        });
      }

      return [];
    }
  };

  return (
    <PolicyContext.Provider
      value={{
        pluginType,
        policyMap,
        policySchemaMap,
        addPolicy,
        updatePolicy,
        removePolicy,
        getPolicyHistory,
      }}
    >
      {children}
    </PolicyContext.Provider>
  );
};

export const usePolicies = (): PolicyContextType => {
  const context = useContext(PolicyContext);
  if (!context) {
    throw new Error("usePolicies must be used within a PolicyProvider");
  }
  return context;
};
