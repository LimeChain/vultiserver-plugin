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
import { mapData } from "../../schema/dcaTableSchema"; // todo these should be dynamic once we have the marketplace
import PolicyFilters from "../policy-filters/PolicyFilters";
import "./PolicyTable.css";
import columnJson from "../../schema/tableSchema.json";
import TokenPair from "@/modules/shared/token-pair/TokenPair";
import PolicyActions from "../policy-actions/PolicyActions";

const componentMap: Record<string, React.FC<any>> = {
  TokenPair,
};

const columns: ColumnDef<any>[] = columnJson.map((col) => {
  const column: ColumnDef<any> = {
    accessorKey: col.accessorKey,
    header: col.header,
  };

  if (col.cellComponent) {
    [
      (column.cell = ({ getValue }) => {
        const Component = componentMap[col.cellComponent];
        return Component ? <Component data={getValue()} /> : getValue();
      }),
    ];
  }

  return column;
});

// all policies must have these actions Pause/Play, Edit, Tx history, Delete
columns.push({
  header: "Actions",
  cell: (info: any) => {
    const policyId = info.row.original.policyId;
    return <PolicyActions policyId={policyId} />;
  },
});

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
