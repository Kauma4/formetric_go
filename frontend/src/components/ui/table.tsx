import * as React from "react"

export const TableHeader = ({ children }: { children: React.ReactNode }) => (
  <thead className="bg-gray-100">{children}</thead>
)

export const TableBody = ({ children }: { children: React.ReactNode }) => (
  <tbody>{children}</tbody>
)


export const TableHead = ({ children }: { children: React.ReactNode }) => (
  <th className="px-4 py-2 text-left">{children}</th>
)

export const TableCell = ({ children }: { children: React.ReactNode }) => (
  <td className="px-4 py-2">{children}</td>
)

// Добавляем HTML атрибуты и поддержку className
type TableProps = React.HTMLAttributes<HTMLTableElement>

export const Table = React.forwardRef<HTMLTableElement, TableProps>(
  ({ className, ...props }, ref) => (
    <table
      ref={ref}
      className={`w-full ${className || ""}`}
      {...props}
    />
  )
)
Table.displayName = "Table"

// Аналогично для остальных компонентов таблицы
type TableRowProps = React.HTMLAttributes<HTMLTableRowElement>

export const TableRow = React.forwardRef<HTMLTableRowElement, TableRowProps>(
  ({ className, ...props }, ref) => (
    <tr
      ref={ref}
      className={`border-b ${className || ""}`}
      {...props}
    />
  )
)
TableRow.displayName = "TableRow"