import { formatIDR } from "@/lib/currency";
import { cn } from "@/lib/utils";
import type { Product } from "@/types/product";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

interface ProductCardProps {
  product: Product;
  cartQuantity: number;
  onAdd: (product: Product) => void;
}

export function ProductCard({ product, cartQuantity, onAdd }: ProductCardProps) {
  const outOfStock = product.stock <= 0;
  const atMaxStock = cartQuantity >= product.stock;

  return (
    <button
      type="button"
      onClick={() => onAdd(product)}
      disabled={outOfStock || atMaxStock}
      className={cn(
        "w-full text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        (outOfStock || atMaxStock) && "cursor-not-allowed opacity-50",
      )}
      aria-label={`Tambah ${product.name} ke keranjang`}
    >
      <Card className="h-full hover:border-primary/40 hover:shadow-sm">
        <CardContent className="flex h-full flex-col gap-2 p-3">
          <div className="flex items-start justify-between gap-2">
            <h3 className="line-clamp-2 text-sm font-semibold leading-tight">
              {product.name}
            </h3>
            {cartQuantity > 0 && (
              <Badge variant="secondary" className="shrink-0 text-xs">
                ×{cartQuantity}
              </Badge>
            )}
          </div>
          <p className="text-xs text-muted-foreground">{product.sku}</p>
          <div className="mt-auto flex items-end justify-between gap-2 pt-1">
            <span className="text-sm font-bold">{formatIDR(product.price)}</span>
            <span
              className={cn(
                "text-xs",
                outOfStock ? "text-destructive" : "text-muted-foreground",
              )}
            >
              {outOfStock ? "Habis" : `Stok ${product.stock}`}
            </span>
          </div>
        </CardContent>
      </Card>
    </button>
  );
}
