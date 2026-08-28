export interface Product {
  id: string;
  sku: string;
  name: string;
  price: number;
  stock: number;
  createdAt: string;
  updatedAt: string;
}

export type ProductId = Product['id'];
export type ProductSku = Product['sku'];
export type ProductStock = Product['stock'];
