import type { Product } from "./product";

export interface CartItem {
  product: Product;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

export type CartItemId = CartItem['product']['id'];
export type CartQuantity = CartItem['quantity'];
export type CartSubtotal = CartItem['subtotal'];
