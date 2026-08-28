"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { CartPanel } from "@/components/pos/cart-panel";
import { ProductList } from "@/components/pos/product-list";
import { useCart } from "@/hooks/use-cart";
import { ApiError, processCheckout, fetchProducts } from "@/lib/api";
import type { CheckoutResponse } from "@/types/checkout";
import type { Product } from "@/types/product";
import { Separator } from "@/components/ui/separator";

export function PosScreen() {
  const [products, setProducts] = useState<Product[]>([]);
  const [search, setSearch] = useState("");
  const [paidAmount, setPaidAmount] = useState(0);
  const [isLoadingProducts, setIsLoadingProducts] = useState(true);
  const [productError, setProductError] = useState<string | null>(null);
  const [isCheckingOut, setIsCheckingOut] = useState(false);
  const [checkoutError, setCheckoutError] = useState<string | null>(null);
  const [checkoutResult, setCheckoutResult] = useState<CheckoutResponse | null>(null);

  const {
    items,
    subtotal,
    total,
    addItem,
    increaseQuantity,
    decreaseQuantity,
    removeItem,
    clearCart,
    getItemQuantity,
  } = useCart();

  const productsById = useMemo(
    () => new Map(products.map((product) => [product.id, product])),
    [products],
  );

  const loadProducts = useCallback(async (showLoading = false) => {
    if (showLoading) {
      setIsLoadingProducts(true);
    }
    setProductError(null);

    try {
      const data = await fetchProducts();
      setProducts(data);
    } catch (error) {
      const message =
        error instanceof ApiError ? error.message : "Server tidak tersedia";
      setProductError(message);
    } finally {
      setIsLoadingProducts(false);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function fetchInitialProducts() {
      setProductError(null);

      try {
        const data = await fetchProducts();
        if (!cancelled) {
          setProducts(data);
        }
      } catch (error) {
        if (!cancelled) {
          const message =
            error instanceof ApiError ? error.message : "Server tidak tersedia";
          setProductError(message);
        }
      } finally {
        if (!cancelled) {
          setIsLoadingProducts(false);
        }
      }
    }

    void fetchInitialProducts();

    return () => {
      cancelled = true;
    };
  }, []);

  const handleAddProduct = useCallback(
    (product: Product) => {
      const latest = productsById.get(product.id) ?? product;
      addItem(latest);
      setCheckoutResult(null);
      setCheckoutError(null);
    },
    [addItem, productsById],
  );

  const handleIncrease = useCallback(
    (product: Product) => {
      const latest = productsById.get(product.id) ?? product;
      increaseQuantity(latest);
    },
    [increaseQuantity, productsById],
  );

  const handleCheckout = useCallback(async () => {
    if (items.length === 0 || paidAmount < total) return;

    setIsCheckingOut(true);
    setCheckoutError(null);
    setCheckoutResult(null);

    try {
      const response = await processCheckout({
        items: items.map((item) => ({
          product_id: item.product.id,
          sku: item.product.sku,
          name: item.product.name,
          quantity: item.quantity,
          unit_price: item.unitPrice,
        })),
        discount: 0,
        tax: 0,
        service_charge: 0,
        paid_amount: paidAmount,
        payment_method: "CASH",
      });

      setCheckoutResult(response);
      clearCart();
      setPaidAmount(0);
      await loadProducts(true);
    } catch (error) {
      const message =
        error instanceof ApiError ? error.message : "Transaksi gagal";
      setCheckoutError(message);
    } finally {
      setIsCheckingOut(false);
    }
  }, [items, paidAmount, total, clearCart, loadProducts]);

  return (
    <div className="flex h-screen flex-col bg-linear-to-br from-background via-background to-muted/20">
      <header className="flex items-center justify-between border-b border-border/50 bg-background/80 backdrop-blur-xl px-6 py-4 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
            <span className="text-lg font-bold text-primary">K</span>
          </div>
          <h1 className="text-xl font-bold tracking-tight gradient-text">KASA POS</h1>
        </div>
        <span className="text-sm font-medium text-muted-foreground px-3 py-1 rounded-full bg-muted/50">CASHIER</span>
      </header>

      <main className="grid min-h-0 flex-1 grid-cols-1 lg:grid-cols-[1fr_380px] xl:grid-cols-[1fr_420px]">
        <section className="min-h-0 border-r border-border/50 bg-background/50 backdrop-blur-sm p-4 lg:p-6">
          <ProductList
            products={products}
            search={search}
            onSearchChange={setSearch}
            cartQuantityByProduct={getItemQuantity}
            onAddProduct={handleAddProduct}
            isLoading={isLoadingProducts}
            error={productError}
          />
        </section>

        <Separator className="lg:hidden" />

        <section className="flex min-h-0 flex-col bg-background/80 backdrop-blur-md p-4 lg:p-6">
          <CartPanel
            items={items}
            productsById={productsById}
            subtotal={subtotal}
            total={total}
            paidAmount={paidAmount}
            onPaidAmountChange={setPaidAmount}
            onIncrease={handleIncrease}
            onDecrease={decreaseQuantity}
            onRemove={removeItem}
            onCheckout={() => void handleCheckout()}
            isCheckingOut={isCheckingOut}
            checkoutError={checkoutError}
            checkoutResult={checkoutResult}
          />
        </section>
      </main>
    </div>
  );
}
