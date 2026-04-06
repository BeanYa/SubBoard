import * as React from 'react'
import { cn } from '@/lib/utils'

const Dialog = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('relative', className)} {...props} />
  )
)
Dialog.displayName = 'Dialog'

export { Dialog }
