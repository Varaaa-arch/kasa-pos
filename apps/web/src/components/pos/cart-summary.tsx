import { formatIDR } from "@/lib/currency";
import { Separator } from "@/components/ui/separator";

interface CartSummaryProps {
  subtotal: number;
  total: number;
}

export function CartSummary({ subtotal, total }: CartSummaryProps) {
  return (
    <div className="space-y-2 text-sm">
      <div className="flex items-center justify-between text-muted-foreground">
        <span>Subtotal</span>
        <span className="tabular-nums">{formatIDR(subtotal)}</span>
      </div>
      <Separator />
      <div className="flex items-center justify-between text-base font-bold">
        <span>Total</span>
        <span className="tabular-nums">{formatIDR(total)}</span>
      </div>
    </div>
  );
}
