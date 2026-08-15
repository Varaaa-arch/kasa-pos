import { Alert, AlertDescription } from "@/components/ui/alert";
import { ProductCard } from "@/components/pos/product-card";
import { ProductSearch } from "@/components/pos/product-search";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { Product } from "@/types/product";

interface ProductListProps {
  products: Product[];
  search: string;
  onSearchChange: (value: string) => void;
  cartQuantityByProduct: (productId: string) => number;
  onAddProduct: (product: Product) => void;
  isLoading: boolean;
  error: string | null;
}

export function ProductList({
  products,
  search,
  onSearchChange,
  cartQuantityByProduct,
  onAddProduct,
  isLoading,
  error,
}: ProductListProps) {
  const filtered = products.filter((product) => {
    const query = search.trim().toLowerCase();
    if (!query) return true;
    return (
      product.name.toLowerCase().includes(query) ||
      product.sku.toLowerCase().includes(query)
    );
  });

  return (
    <div className="flex h-full flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Produk</h2>
        <span className="text-sm text-muted-foreground">
          {filtered.length} item
        </span>
      </div>

      <ProductSearch value={search} onChange={onSearchChange} />

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {isLoading ? (
        <div className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
          Memuat produk...
        </div>
      ) : (
        <ScrollArea className="flex-1 pr-2">
          {filtered.length === 0 ? (
            <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
              Produk tidak ditemukan
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-3 pb-2 xl:grid-cols-3">
              {filtered.map((product) => (
                <ProductCard
                  key={product.id}
                  product={product}
                  cartQuantity={cartQuantityByProduct(product.id)}
                  onAdd={onAddProduct}
                />
              ))}
            </div>
          )}
        </ScrollArea>
      )}
    </div>
  );
}
