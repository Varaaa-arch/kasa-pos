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
        "w-full text-left transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "group relative",
        (outOfStock || atMaxStock) && "cursor-not-allowed opacity-50",
      )}
      aria-label={`Tambah ${product.name} ke keranjang`}
    >
      <Card className={cn(
        "h-full transition-all duration-300 ease-out",
        "glass-card hover:shadow-glow hover:-translate-y-1 hover:scale-[1.02]",
        "border-border/50 hover:border-primary/50",
        !(outOfStock || atMaxStock) && "active:scale-[0.98]"
      )}>
        <CardContent className="flex h-full flex-col gap-3 p-4">
          <div className="flex items-start justify-between gap-2">
            <h3 className="line-clamp-2 text-sm font-semibold leading-tight text-foreground">
              {product.name}
            </h3>
            {cartQuantity > 0 && (
              <Badge 
                variant="secondary" 
                className="shrink-0 text-xs font-medium bg-primary/10 text-primary border-primary/20"
              >
                ×{cartQuantity}
              </Badge>
            )}
          </div>
          <p className="text-xs text-muted-foreground font-mono">{product.sku}</p>
          <div className="mt-auto flex items-end justify-between gap-2 pt-2 border-t border-border/50">
            <span className="text-sm font-bold text-foreground">{formatIDR(product.price)}</span>
            <span
              className={cn(
                "text-xs font-medium",
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
