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
    <div className="flex h-full flex-col gap-4 animate-fade-in">
      <div className="flex items-center justify-between px-1">
        <h2 className="text-lg font-semibold text-foreground">Produk</h2>
        <span className="text-sm text-muted-foreground font-medium">
          {filtered.length} item
        </span>
      </div>

      <ProductSearch value={search} onChange={onSearchChange} />

      {error && (
        <Alert variant="destructive" className="animate-slide-up">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {isLoading ? (
        <div className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
            <span>Memuat produk...</span>
          </div>
        </div>
      ) : (
        <ScrollArea className="flex-1 pr-2 scrollbar-modern">
          {filtered.length === 0 ? (
            <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
              Produk tidak ditemukan
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4 pb-2 xl:grid-cols-3">
              {filtered.map((product, index) => (
                <div 
                  key={product.id} 
                  className="animate-scale-in"
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <ProductCard
                    product={product}
                    cartQuantity={cartQuantityByProduct(product.id)}
                    onAdd={onAddProduct}
                  />
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      )}
    </div>
  );
}
