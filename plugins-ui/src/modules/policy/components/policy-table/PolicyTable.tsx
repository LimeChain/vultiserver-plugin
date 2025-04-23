import {
  CellContext,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  Row,
  RowData,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useState } from "react";
import { usePolicies } from "@/modules/policy/context/PolicyProvider";
import PolicyFilters from "@/modules/policy/components/policy-filters/PolicyFilters";
import "@/modules/policy/components/policy-table/PolicyTable.css";
import TokenPair, {
  type TokenPairProps,
} from "@/modules/shared/token-pair/TokenPair";
import PolicyActions from "@/modules/policy/components/policy-actions/PolicyActions";
import TokenName, {
  type TokenNameProps,
} from "@/modules/shared/token-name/TokenName";
import TokenAmount, {
  type TokenAmountProps,
} from "@/modules/shared/token-amount/TokenAmount";
import { mapTableColumnData } from "@/modules/policy/utils/policy.util";
import ActiveStatus, {
  type ActiveStatusProps,
} from "@/modules/shared/active-status/ActiveStatus";
import {
  PolicySchema,
  PolicyTableColumn,
} from "@/modules/policy/models/policy";

const componentMap: Record<
  string,
  ({ data, row }: { data: unknown; row: Row<unknown> }) => JSX.Element
> = {
  TokenPair: (props) => <TokenPair {...(props as TokenPairProps)} />,
  TokenName: (props) => <TokenName {...(props as TokenNameProps)} />,
  TokenAmount: (props) => <TokenAmount {...(props as TokenAmountProps)} />,
  ActiveStatus: (props) => <ActiveStatus {...(props as ActiveStatusProps)} />,
};

const getTableColumns = (schema: PolicySchema) => {
  const columns: PolicyTableColumn[] = schema.table.columns.map(
    (col: PolicyTableColumn) => {
      const column: PolicyTableColumn = {
        accessorKey: col.accessorKey,
        header: col.header,
      };
      if (col.cellComponent) {
        column.cell = ({ getValue, row }) => {
          const func = componentMap[`${col.cellComponent}`];
          return func ? func({ data: getValue(), row }) : getValue();
        };
      }
      return column;
    }
  );

  // all policies must have these actions Pause/Play, Edit, Tx history, Delete
  columns.push({
    header: "Actions",
    cell: (info: CellContext<RowData, unknown>) => {
      const policyId = (info.row.original as Record<string, unknown>).policyId;
      return <PolicyActions policyId={`${policyId}`} />;
    },
  } as PolicyTableColumn);

  return columns;
};

const PolicyTable = () => {
  const [data, setData] = useState<unknown[]>(() => []);
  const { policyMap, policySchemaMap, pluginType } = usePolicies();
  const [columns, setColumns] = useState<PolicyTableColumn[]>([]);

  useEffect(() => {
    const savedSchema = policySchemaMap.get(pluginType);

    if (
      savedSchema &&
      savedSchema.table &&
      savedSchema.table.columns &&
      savedSchema.table.mapping
    ) {
      const mappedColumns: PolicyTableColumn[] = getTableColumns(savedSchema);

      setColumns(mappedColumns);

      const transformedData = [];
      for (const [, value] of policyMap) {
        const obj: Record<string, unknown> = mapTableColumnData(
          value,
          savedSchema.table.mapping
        );
        transformedData.push(obj);
      }
      setData(transformedData);
    }
  }, [policySchemaMap, policyMap, pluginType]);

  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]); // can set initial column filter state here

  const table = useReactTable({
    data,
    columns,
    state: {
      columnFilters,
    },
    onColumnFiltersChange: setColumnFilters,
    getFilteredRowModel: getFilteredRowModel(), // needed for client-side filtering
    getCoreRowModel: getCoreRowModel(),
  });

  if (columns.length === 0) return;

  return (
    <div data-testid="policy-table-wrapper">
      <PolicyFilters onFiltersChange={setColumnFilters} />

      {policySchemaMap.has(pluginType) && (
        <table data-testid="policy-table" className="policy-table">
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th data-testid="policy-table-headers" key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row) => (
              <tr key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <td data-testid="policy-table-cells" key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
            {table.getRowModel().rows.length === 0 && (
              <tr data-testid="policy-form-empty-row">
                <td
                  colSpan={table.getAllColumns().length}
                  className="empty-message-row"
                >
                  Nothing to see here yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default PolicyTable;
