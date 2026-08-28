import { CartItemRow } from "@/components/pos/cart-item";
import { CartSummary } from "@/components/pos/cart-summary";
import { CheckoutResult } from "@/components/pos/checkout-result";
import { PaymentPanel } from "@/components/pos/payment-panel";
import type { CartItem } from "@/types/cart";
import type { CheckoutResponse } from "@/types/checkout";
import type { Product } from "@/types/product";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

interface CartPanelProps {
  items: CartItem[];
  productsById: Map<string, Product>;
  subtotal: number;
  total: number;
  paidAmount: number;
  onPaidAmountChange: (amount: number) => void;
  onIncrease: (product: Product) => void;
  onDecrease: (productId: string) => void;
  onRemove: (productId: string) => void;
  onCheckout: () => void;
  isCheckingOut: boolean;
  checkoutError: string | null;
  checkoutResult: CheckoutResponse | null;
}

export function CartPanel({
  items,
  productsById,
  subtotal,
  total,
  paidAmount,
  onPaidAmountChange,
  onIncrease,
  onDecrease,
  onRemove,
  onCheckout,
  isCheckingOut,
  checkoutError,
  checkoutResult,
}: CartPanelProps) {
  const canCheckout =
    items.length > 0 && paidAmount >= total && total > 0 && !isCheckingOut;

  return (
    <div className="flex h-full flex-col gap-4 animate-fade-in">
      <div className="flex items-center justify-between px-1">
        <h2 className="text-lg font-semibold text-foreground">Keranjang</h2>
        <span className="text-sm text-muted-foreground font-medium">{items.length} item</span>
      </div>

      {checkoutResult && <CheckoutResult result={checkoutResult} />}

      <ScrollArea className="min-h-0 flex-1 scrollbar-modern">
        {items.length === 0 ? (
          <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
            Keranjang kosong
          </div>
        ) : (
          <div className="space-y-2">
            {items.map((item) => {
              const product = productsById.get(item.product.id) ?? item.product;
              return (
                <CartItemRow
                  key={item.product.id}
                  item={item}
                  product={product}
                  onIncrease={onIncrease}
                  onDecrease={onDecrease}
                  onRemove={onRemove}
                />
              );
            })}
          </div>
        )}
      </ScrollArea>

      <Separator className="bg-border/50" />

      <CartSummary subtotal={subtotal} total={total} />

      <PaymentPanel
        paidAmount={paidAmount}
        total={total}
        onPaidAmountChange={onPaidAmountChange}
        disabled={isCheckingOut || items.length === 0}
      />

      {checkoutError && (
        <Alert variant="destructive" className="animate-slide-up">
          <AlertDescription>{checkoutError}</AlertDescription>
        </Alert>
      )}

      <Button
        type="button"
        size="lg"
        className={cn(
          "h-12 w-full text-base font-bold btn-glow",
          "transition-all duration-300 ease-out",
          canCheckout && "hover:shadow-glow hover:-translate-y-0.5",
          !canCheckout && "opacity-50 cursor-not-allowed"
        )}
        onClick={onCheckout}
        disabled={!canCheckout}
      >
        {isCheckingOut ? "Memproses..." : "BAYAR"}
      </Button>
    </div>
  );
}
