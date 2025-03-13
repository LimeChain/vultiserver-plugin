import {
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useState } from "react";
import { usePolicies } from "@/modules/policy/context/PolicyProvider";
import { mapData, dcaPolicyColumns } from "../../schema/dcaTableSchema"; // todo these should be dynamic once we have the marketplace
import PolicyFilters from "../policy-filters/PolicyFilters";
import "./PolicyTable.css";
import columnJson from "../../schema/tableSchema.json";
import TokenPair from "@/modules/shared/token-pair/TokenPair";

const componentMap: Record<string, React.FC<any>> = {
  TokenPair,
};

const columns: ColumnDef<any>[] = columnJson.map((col) => ({
  accessorKey: col.accessorKey,
  header: col.header,
  cell: col.cellComponent
    ? ({ getValue }) => {
        const Component = componentMap[col.cellComponent]; // Resolve Component
        return Component ? <Component pair={getValue()} /> : getValue();
      }
    : undefined, // Default to normal rendering if no custom component
}));

const PolicyTable = () => {
  const [data, setData] = useState<any>(() => []);
  const { policyMap } = usePolicies();

  useEffect(() => {
    const policies = [];
    for (const [_, value] of policyMap) {
      policies.push(mapData(value));
    }
    setData(policies);
    console.log("policies", policies);
  }, [policyMap]);

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

  return (
    <div>
      <PolicyFilters onFiltersChange={setColumnFilters} />
      <table className="policy-table">
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th key={header.id}>
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
                <td key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
          {table.getRowModel().rows.length === 0 && (
            <tr>
              <td colSpan={table.getAllColumns().length}>
                Nothing to see here yet.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
};

export default PolicyTable;
