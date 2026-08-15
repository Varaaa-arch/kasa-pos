"use client";

import { useCallback, useMemo, useState } from "react";
import type { CartItem } from "@/types/cart";
import type { Product } from "@/types/product";

export function useCart() {
  const [items, setItems] = useState<CartItem[]>([]);

  const total = useMemo(
    () => items.reduce((sum, item) => sum + item.subtotal, 0),
    [items],
  );

  const subtotal = total;

  const getItemQuantity = useCallback(
    (productId: string) => items.find((item) => item.product.id === productId)?.quantity ?? 0,
    [items],
  );

  const addItem = useCallback((product: Product, quantity = 1) => {
    if (product.stock <= 0) return;

    setItems((current) => {
      const existing = current.find((item) => item.product.id === product.id);

      if (existing) {
        const nextQuantity = existing.quantity + quantity;
        if (nextQuantity > product.stock) return current;

        return current.map((item) =>
          item.product.id === product.id
            ? {
                ...item,
                product,
                quantity: nextQuantity,
                subtotal: nextQuantity * item.unitPrice,
              }
            : item,
        );
      }

      if (quantity > product.stock) return current;

      return [
        ...current,
        {
          product,
          quantity,
          unitPrice: product.price,
          subtotal: quantity * product.price,
        },
      ];
    });
  }, []);

  const increaseQuantity = useCallback((product: Product) => {
    setItems((current) =>
      current.map((item) => {
        if (item.product.id !== product.id) return item;
        if (item.quantity >= product.stock) return item;

        const nextQuantity = item.quantity + 1;
        return {
          ...item,
          product,
          quantity: nextQuantity,
          subtotal: nextQuantity * item.unitPrice,
        };
      }),
    );
  }, []);

  const decreaseQuantity = useCallback((productId: string) => {
    setItems((current) =>
      current
        .map((item) => {
          if (item.product.id !== productId) return item;

          const nextQuantity = item.quantity - 1;
          if (nextQuantity <= 0) return null;

          return {
            ...item,
            quantity: nextQuantity,
            subtotal: nextQuantity * item.unitPrice,
          };
        })
        .filter((item): item is CartItem => item !== null),
    );
  }, []);

  const removeItem = useCallback((productId: string) => {
    setItems((current) => current.filter((item) => item.product.id !== productId));
  }, []);

  const clearCart = useCallback(() => {
    setItems([]);
  }, []);

  return {
    items,
    subtotal,
    total,
    addItem,
    increaseQuantity,
    decreaseQuantity,
    removeItem,
    clearCart,
    getItemQuantity,
  };
}
