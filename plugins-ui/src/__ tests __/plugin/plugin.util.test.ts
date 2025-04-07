import { PluginFee } from "@/modules/plugin/models/plugin";
import { getPluginFee } from "@/modules/plugin/utils/plugin.util";
import { describe, expect, it } from "vitest";

describe("getPluginFee", () => {
  it("should render correct price for free pricing", () => {
    const pluginFee: PluginFee = {
      id: "1",
      type: "FREE",
      metric: "FIXED",
      amount: 0,
    };

    const result = getPluginFee(pluginFee);

    expect(result).toStrictEqual("Plugin fee: 0USDC free ");
  });

  it("should render correct price for fixed single time pricing", () => {
    const pluginFee: PluginFee = {
      id: "1",
      type: "SINGLE",
      metric: "FIXED",
      amount: 5,
    };

    const result = getPluginFee(pluginFee);

    expect(result).toStrictEqual("Plugin fee: 5USDC per installation ");
  });

  it("should render correct price for precentage single time pricing", () => {
    const pluginFee: PluginFee = {
      id: "1",
      type: "SINGLE",
      metric: "PERCENTAGE",
      amount: 5,
    };

    const result = getPluginFee(pluginFee);

    expect(result).toStrictEqual("Plugin fee: 5% per installation ");
  });

  it("should render correct price for fixed per transaction pricing", () => {
    const pluginFee: PluginFee = {
      id: "1",
      type: "PER_TX",
      metric: "FIXED",
      amount: 5,
    };

    const result = getPluginFee(pluginFee);

    expect(result).toStrictEqual("Plugin fee: 5USDC per trade ");
  });

  it("should render correct price for precentage per transaction pricing", () => {
    const pluginFee: PluginFee = {
      id: "1",
      type: "PER_TX",
      metric: "PERCENTAGE",
      amount: 5,
    };

    const result = getPluginFee(pluginFee);

    expect(result).toStrictEqual("Plugin fee: 5% per trade ");
  });

  it("should render correct price for fixed recurring pricing", () => {
    const pluginFee: PluginFee = {
      id: "1",
      type: "RECURRING",
      metric: "FIXED",
      frequency: "ANNUAL",
      amount: 5,
    };

    const result = getPluginFee(pluginFee);

    expect(result).toStrictEqual("Plugin fee: 5USDC per year");
  });

  it("should render correct price for precentage recurring pricing", () => {
    const pluginFee: PluginFee = {
      id: "1",
      type: "RECURRING",
      metric: "PERCENTAGE",
      frequency: "MONTHLY",
      amount: 5,
    };

    const result = getPluginFee(pluginFee);

    expect(result).toStrictEqual("Plugin fee: 5% per month");
  });
});
