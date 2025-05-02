import * as React from "react"

export const DialogHeader = ({ children }: { children: React.ReactNode }) => (
  <div className="mb-4">{children}</div>
)

export const DialogTitle = ({ children }: { children: React.ReactNode }) => (
  <h2 className="text-xl font-bold">{children}</h2>
)

type DialogProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    children: React.ReactNode
  }
  
  export const Dialog = ({ 
    open, 
    onOpenChange,
    children,
    ...props
  }: DialogProps & React.HTMLAttributes<HTMLDivElement>) => (
    <div 
      className={`fixed inset-0 bg-black/50 ${open ? "flex" : "hidden"} items-center justify-center`}
      {...props}
    >
      {children}
    </div>
  )
  
  type DialogContentProps = React.HTMLAttributes<HTMLDivElement>
  
  export const DialogContent = React.forwardRef<HTMLDivElement, DialogContentProps>(
    ({ className, ...props }, ref) => (
      <div
        ref={ref}
        className={`bg-white p-6 rounded-lg ${className || ""}`}
        {...props}
      />
    )
  )
  DialogContent.displayName = "DialogContent"