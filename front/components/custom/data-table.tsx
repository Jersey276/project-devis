"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ArrowDownIcon,
  ArrowUpDownIcon,
  ArrowUpIcon,
  EllipsisVertical,
} from "lucide-react";
import { createContext, useContext } from "react";

export type DataTableRowAction = {
  type: "callback" | "link";
  label: string;
  icon?: React.ComponentType;
  href?: string;
  hidden?: boolean;
  disabled?: boolean;
  callback?: (row: object) => void;
};

// Create a context for the data
type DataContextType = {
  datas: object[];
  row_actions?: DataTableRowAction[];
  filters?: React.ReactNode;
  sortBy: string;
  sortDirection?: "asc" | "desc";
  setSort: (column: string, direction: "asc" | "desc") => void;
};

const DataContext = createContext<DataContextType | undefined>(undefined);

function useDataContext() {
  const context = useContext(DataContext);
  if (!context) {
    throw new Error("useDataContext must be used within a DataTable");
  }
  return context;
}

type DataTableParams = {
  datas: object[];
  row_actions?: DataTableRowAction[];
  filters?: React.ReactNode;
  sortBy: string;
  sortDirection?: "asc" | "desc";
  onSortChange?: (column: string, direction: "asc" | "desc") => void;
};

function DataTable({
  children,
  datas,
  row_actions,
  filters,
  sortBy,
  sortDirection,
  onSortChange,
}: DataTableParams & { children: React.ReactNode }) {
  const [currentSortBy, setCurrentSortBy] = React.useState(sortBy);
  const [currentSortDirection, setCurrentSortDirection] = React.useState<
    "asc" | "desc"
  >(sortDirection ?? "asc");

  React.useEffect(() => {
    setCurrentSortBy(sortBy);
  }, [sortBy]);

  React.useEffect(() => {
    if (sortDirection) {
      setCurrentSortDirection(sortDirection);
    }
  }, [sortDirection]);

  const setSort = React.useCallback(
    (column: string, direction: "asc" | "desc") => {
      setCurrentSortBy(column);
      setCurrentSortDirection(direction);
      onSortChange?.(column, direction);
    },
    [onSortChange],
  );

  const contextValue = React.useMemo(
    () => ({
      datas,
      row_actions: row_actions || [],
      filters,
      sortBy: currentSortBy,
      sortDirection: currentSortDirection,
      setSort,
    }),
    [datas, row_actions, filters, currentSortBy, currentSortDirection, setSort],
  );

  return (
    <DataContext.Provider value={contextValue}>
      <Table>{children}</Table>
    </DataContext.Provider>
  );
}

function DataTableHeader({ children }: { children: React.ReactNode }) {
  return <TableHeader>{children}</TableHeader>;
}

function DataTableHead({ children }: { children: React.ReactNode }) {
  return <TableHead>{children}</TableHead>;
}

function DataTableSortableHead({
  name,
  children,
}: {
  name: string;
  children: React.ReactNode;
}) {
  const context = useDataContext();

  const handleSort = () => {
    if (context.sortBy) {
      const newDirection =
        context.sortBy === name && context.sortDirection === "asc"
          ? "desc"
          : "asc";
      context.setSort(name, newDirection);
    }
  };

  return (
    <TableHead>
      <Button
        variant="ghost"
        className="flex flex-row justify-between items-center gap-2 p-0"
        onClick={handleSort}
      >
        {context.sortBy === name && context.sortDirection === "asc" ? (
          <ArrowUpIcon />
        ) : context.sortBy === name && context.sortDirection === "desc" ? (
          <ArrowDownIcon />
        ) : (
          <ArrowUpDownIcon />
        )}
        {children}
      </Button>
    </TableHead>
  );
}

function DataTableBody({ children }: { children: React.ReactNode }) {
  return <TableBody>{children}</TableBody>;
}

function DataTableRowActions({ id }: { id?: string | number }) {
  const { row_actions } = useDataContext();
  if (!row_actions || row_actions.length === 0) {
    return null;
  }
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <EllipsisVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {row_actions?.map((element) => {
          if (element.type === "link" && element.href) {
            return (
              <DropdownMenuItem key={element.label} asChild>
                <a
                  href={
                    id ? element.href.replace("{id}", String(id)) : element.href
                  }
                  className="flex items-center gap-2"
                >
                  {element.icon ? React.createElement(element.icon) : null}
                  {element.label}
                </a>
              </DropdownMenuItem>
            );
          } else if (element.type === "callback" && element.callback) {
            return (
              <DropdownMenuItem
                key={element.label}
                onClick={() => element.callback!(element)}
              >
                <div className="flex items-center gap-2">
                  {element.icon ? React.createElement(element.icon) : null}
                  {element.label}
                </div>
              </DropdownMenuItem>
            );
          }
          return null;
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function DataTableRow({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return <TableRow className={className}>{children}</TableRow>;
}

function DataTableCell({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return <TableCell className={className}>{children}</TableCell>;
}

export {
  DataTable,
  DataTableHead,
  DataTableSortableHead,
  DataTableBody,
  DataTableHeader,
  DataTableRowActions,
  useDataContext,
  DataTableRow,
  DataTableCell,
};
