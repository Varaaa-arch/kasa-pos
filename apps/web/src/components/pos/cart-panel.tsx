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
    <div className="flex h-full flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Keranjang</h2>
        <span className="text-sm text-muted-foreground">{items.length} item</span>
      </div>

      {checkoutResult && <CheckoutResult result={checkoutResult} />}

      <ScrollArea className="min-h-0 flex-1">
        {items.length === 0 ? (
          <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
            Keranjang kosong
          </div>
        ) : (
          <div>
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

      <Separator />

      <CartSummary subtotal={subtotal} total={total} />

      <PaymentPanel
        paidAmount={paidAmount}
        total={total}
        onPaidAmountChange={onPaidAmountChange}
        disabled={isCheckingOut || items.length === 0}
      />

      {checkoutError && (
        <Alert variant="destructive">
          <AlertDescription>{checkoutError}</AlertDescription>
        </Alert>
      )}

      <Button
        type="button"
        size="lg"
        className="h-12 w-full text-base font-bold"
        onClick={onCheckout}
        disabled={!canCheckout}
      >
        {isCheckingOut ? "Memproses..." : "BAYAR"}
      </Button>
    </div>
  );
}
