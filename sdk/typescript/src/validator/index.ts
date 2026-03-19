/**
 * Validate .bknd DataTable rows against Object/Relation schema definitions.
 */

import type { BknNetwork, BknObject, DataProperty, DataTable } from "../models/index.js";
import { allObjects, allDataTables } from "../models/index.js";

export interface ValidationError {
  table: string;
  row: number | null;
  column: string;
  code: string;
  message: string;
}

export interface ValidationResult {
  errors: ValidationError[];
  get ok(): boolean;
}

function createResult(errors: ValidationError[]): ValidationResult {
  return {
    errors,
    get ok() {
      return this.errors.length === 0;
    },
  };
}

const CONSTRAINT_SPLIT_RE = /;\s*/;

function parseConstraints(raw: string): string[] {
  if (!raw?.trim()) return [];
  return raw
    .trim()
    .split(CONSTRAINT_SPLIT_RE)
    .map((c) => c.trim())
    .filter(Boolean);
}

function tryFloat(val: string): number | null {
  const n = parseFloat(val);
  return Number.isNaN(n) ? null : n;
}

function checkCell(
  value: string,
  prop: DataProperty,
  tableName: string,
  rowIdx: number,
  errors: ValidationError[]
): void {
  const constraints = parseConstraints(prop.constraint);
  const col = prop.property;

  for (const cst of constraints) {
    if (cst === "not_null") {
      if (!value.trim()) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "not_null",
          message: "value must not be empty",
        });
      }
    } else if (cst.startsWith("regex:")) {
      const pattern = cst.slice(6);
      const re = new RegExp(`^${pattern}$`);
      if (value.trim() && !re.test(value.trim())) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "regex",
          message: `'${value}' does not match /${pattern}/`,
        });
      }
    } else if (cst.startsWith("in(") && cst.endsWith(")")) {
      const allowed = cst
        .slice(3, -1)
        .split(",")
        .map((v) => v.trim());
      if (value.trim() && !allowed.includes(value.trim())) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "in",
          message: `'${value}' not in ${JSON.stringify(allowed)}`,
        });
      }
    } else if (cst.startsWith("not_in(") && cst.endsWith(")")) {
      const forbidden = cst
        .slice(7, -1)
        .split(",")
        .map((v) => v.trim());
      if (value.trim() && forbidden.includes(value.trim())) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "not_in",
          message: `'${value}' is forbidden (${JSON.stringify(forbidden)})`,
        });
      }
    } else if (cst.startsWith("range(") && cst.endsWith(")")) {
      const parts = cst.slice(6, -1).split(",");
      if (parts.length === 2 && value.trim()) {
        const lo = tryFloat(parts[0]);
        const hi = tryFloat(parts[1]);
        const v = tryFloat(value.trim());
        if (lo != null && hi != null && v != null && (v < lo || v > hi)) {
          errors.push({
            table: tableName,
            row: rowIdx,
            column: col,
            code: "range",
            message: `${v} not in [${lo}, ${hi}]`,
          });
        }
      }
    } else if (cst.startsWith(">=")) {
      const threshold = tryFloat(cst.slice(2).trim());
      const v = tryFloat(value.trim());
      if (threshold != null && v != null && v < threshold) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: ">=",
          message: `${v} < ${threshold}`,
        });
      }
    } else if (cst.startsWith("<=")) {
      const threshold = tryFloat(cst.slice(2).trim());
      const v = tryFloat(value.trim());
      if (threshold != null && v != null && v > threshold) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "<=",
          message: `${v} > ${threshold}`,
        });
      }
    } else if (cst.startsWith(">") && !cst.startsWith(">=")) {
      const threshold = tryFloat(cst.slice(1).trim());
      const v = tryFloat(value.trim());
      if (threshold != null && v != null && v <= threshold) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: ">",
          message: `${v} <= ${threshold}`,
        });
      }
    } else if (cst.startsWith("<") && !cst.startsWith("<=")) {
      const threshold = tryFloat(cst.slice(1).trim());
      const v = tryFloat(value.trim());
      if (threshold != null && v != null && v >= threshold) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "<",
          message: `${v} >= ${threshold}`,
        });
      }
    } else if (cst.startsWith("== ")) {
      const expected = cst.slice(3).trim();
      if (value.trim() && value.trim() !== expected) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "==",
          message: `'${value}' != '${expected}'`,
        });
      }
    } else if (cst.startsWith("!= ")) {
      const forbiddenVal = cst.slice(3).trim();
      if (value.trim() === forbiddenVal) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "!=",
          message: `value must not be '${forbiddenVal}'`,
        });
      }
    }
  }

  const propType = prop.type.trim().toLowerCase();
  if (value.trim()) {
    if (propType === "bool") {
      if (!["true", "false"].includes(value.trim().toLowerCase())) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "type_bool",
          message: `'${value}' is not a valid bool`,
        });
      }
    } else if (
      ["int32", "int64", "integer", "float32", "float64", "float"].includes(
        propType
      )
    ) {
      if (tryFloat(value.trim()) === null) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "type_numeric",
          message: `'${value}' is not a valid ${propType}`,
        });
      }
    } else if (propType.startsWith("decimal")) {
      if (tryFloat(value.trim()) === null) {
        errors.push({
          table: tableName,
          row: rowIdx,
          column: col,
          code: "type_numeric",
          message: `'${value}' is not a valid ${propType}`,
        });
      }
    }
  }
}

export interface ValidateOptions {
  mode?: "strict" | "compat";
}

export function validateDataTable(
  table: DataTable,
  schema?: BknObject | null,
  network?: BknNetwork | null
): ValidationResult {
  const result = createResult([]);
  const tableName = table.object_or_relation || table.source_path;

  if (table.is_relation) {
    return result;
  }

  if (!schema && network) {
    schema = allObjects(network).find((o) => o.id === table.object_or_relation) ?? undefined;
  }

  if (!schema) {
    result.errors.push({
      table: tableName,
      row: null,
      column: "",
      code: "no_schema",
      message: `no Object schema found for '${table.object_or_relation}'`,
    });
    return result;
  }

  if (schema.data_source) {
    const sourceType = schema.data_source.type.trim().toLowerCase();
    if (["data_view", "connection"].includes(sourceType)) {
      result.errors.push({
        table: tableName,
        row: null,
        column: "",
        code: "readonly_data_source",
        message: `object data source type '${schema.data_source.type}' cannot be materialized in .bknd`,
      });
      return result;
    }
  }

  const schemaProps = new Map(
    schema.data_properties.map((dp) => [dp.property, dp])
  );
  const schemaPropNames = new Set(schemaProps.keys());

  for (const col of table.columns) {
    if (!schemaPropNames.has(col)) {
      result.errors.push({
        table: tableName,
        row: null,
        column: col,
        code: "extra_column",
        message: `column '${col}' not defined in Object schema`,
      });
    }
  }

  for (const col of schemaPropNames) {
    if (!table.columns.includes(col)) {
      result.errors.push({
        table: tableName,
        row: null,
        column: col,
        code: "missing_column",
        message: `schema property '${col}' not present in data`,
      });
    }
  }

  const pkProps = schema.data_properties.filter((dp) => dp.primary_key);
  const pkSeen: Record<string, number[]> = {};

  table.rows.forEach((row, rowIdx) => {
    const idx = rowIdx + 1;
    for (const [colName, dp] of schemaProps) {
      if (!(colName in row)) continue;
      const value = row[colName] ?? "";
      checkCell(value, dp, tableName, idx, result.errors);
    }

    if (pkProps.length > 0) {
      const pkVal = pkProps.map((dp) => row[dp.property] ?? "").join("|");
      if (!pkSeen[pkVal]) pkSeen[pkVal] = [];
      pkSeen[pkVal].push(idx);
    }
  });

  for (const [pkKey, rows] of Object.entries(pkSeen)) {
    if (rows.length > 1) {
      const pkColNames = pkProps.map((dp) => dp.property).join(", ");
      result.errors.push({
        table: tableName,
        row: rows[0],
        column: pkColNames,
        code: "pk_duplicate",
        message: `duplicate primary key '${pkKey}' in rows ${JSON.stringify(rows)}`,
      });
    }
  }

  return result;
}

export function validateDocument(
  _doc: import("../models/index.js").BknDocument,
  _options?: ValidateOptions
): ValidationResult {
  return createResult([]);
}

export function validateNetwork(
  network: BknNetwork,
  _options?: ValidateOptions
): ValidationResult {
  const result = createResult([]);
  for (const table of allDataTables(network)) {
    const tableResult = validateDataTable(table, undefined, network);
    result.errors.push(...tableResult.errors);
  }
  return result;
}
