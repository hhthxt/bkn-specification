import { describe, it, expect } from "vitest";
import { validateDataTable, validateNetwork } from "../src/validator/index.js";
import type { DataTable, BknObject } from "../src/models/index.js";

describe("validateDataTable", () => {
  it("returns ok for valid table", () => {
    const table: DataTable = {
      object_or_relation: "pod",
      is_relation: false,
      columns: ["id", "name"],
      rows: [{ id: "p1", name: "pod-1" }],
      source_path: "",
      network: "",
    };
    const schema: BknObject = {
      id: "pod",
      name: "Pod",
      description: "",
      tags: [],
      owner: "",
      data_properties: [
        { property: "id", display_name: "ID", type: "string", constraint: "not_null", description: "", primary_key: true, display_key: false, index: false },
        { property: "name", display_name: "Name", type: "string", constraint: "", description: "", primary_key: false, display_key: false, index: false },
      ],
      property_overrides: [],
      logic_properties: [],
      business_semantics: "",
    };
    const result = validateDataTable(table, schema);
    expect(result.ok).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it("reports missing column", () => {
    const table: DataTable = {
      object_or_relation: "pod",
      is_relation: false,
      columns: ["id"],
      rows: [{ id: "p1" }],
      source_path: "",
      network: "",
    };
    const schema: BknObject = {
      id: "pod",
      name: "Pod",
      description: "",
      tags: [],
      owner: "",
      data_properties: [
        { property: "id", display_name: "ID", type: "string", constraint: "", description: "", primary_key: true, display_key: false, index: false },
        { property: "name", display_name: "Name", type: "string", constraint: "not_null", description: "", primary_key: false, display_key: false, index: false },
      ],
      property_overrides: [],
      logic_properties: [],
      business_semantics: "",
    };
    const result = validateDataTable(table, schema);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.code === "missing_column")).toBe(true);
  });

  it("reports not_null violation", () => {
    const table: DataTable = {
      object_or_relation: "pod",
      is_relation: false,
      columns: ["id"],
      rows: [{ id: "" }],
      source_path: "",
      network: "",
    };
    const schema: BknObject = {
      id: "pod",
      name: "Pod",
      description: "",
      tags: [],
      owner: "",
      data_properties: [
        { property: "id", display_name: "ID", type: "string", constraint: "not_null", description: "", primary_key: true, display_key: false, index: false },
      ],
      property_overrides: [],
      logic_properties: [],
      business_semantics: "",
    };
    const result = validateDataTable(table, schema);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.code === "not_null")).toBe(true);
  });
});
