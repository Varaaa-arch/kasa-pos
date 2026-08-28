import { Minus, Plus, Trash2 } from "lucide-react";
import { formatIDR } from "@/lib/currency";
import { cn } from "@/lib/utils";
import type { CartItem } from "@/types/cart";
import type { Product } from "@/types/product";
import { Button } from "@/components/ui/button";

interface CartItemRowProps {
  item: CartItem;
  product: Product;
  onIncrease: (product: Product) => void;
  onDecrease: (productId: string) => void;
  onRemove: (productId: string) => void;
}

export function CartItemRow({
  item,
  product,
  onIncrease,
  onDecrease,
  onRemove,
}: CartItemRowProps) {
  const atMaxStock = item.quantity >= product.stock;

  return (
    <div className="flex items-start justify-between gap-3 border-b border-border/50 py-3 last:border-b-0 transition-all duration-200 hover:bg-muted/30 rounded-lg px-2 -mx-2">
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-foreground">{item.product.name}</p>
        <p className="text-xs text-muted-foreground">
          {formatIDR(item.unitPrice)} × {item.quantity}
        </p>
        <div className="mt-2 flex items-center gap-1">
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="size-7 transition-all duration-200 hover:bg-primary/10 hover:border-primary/30"
            onClick={() => onDecrease(item.product.id)}
            aria-label={`Kurangi ${item.product.name}`}
          >
            <Minus className="size-3.5" />
          </Button>
          <span className="w-8 text-center text-sm font-medium text-foreground">{item.quantity}</span>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className={cn(
              "size-7 transition-all duration-200 hover:bg-primary/10 hover:border-primary/30",
              atMaxStock && "opacity-50 cursor-not-allowed"
            )}
            onClick={() => onIncrease(product)}
            disabled={atMaxStock}
            aria-label={`Tambah ${item.product.name}`}
          >
            <Plus className="size-3.5" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-7 text-destructive hover:text-destructive hover:bg-destructive/10 transition-all duration-200"
            onClick={() => onRemove(item.product.id)}
            aria-label={`Hapus ${item.product.name}`}
          >
            <Trash2 className="size-3.5" />
          </Button>
        </div>
      </div>
      <p className="shrink-0 text-sm font-semibold tabular-nums text-foreground">
        {formatIDR(item.subtotal)}
      </p>
    </div>
  );
}
