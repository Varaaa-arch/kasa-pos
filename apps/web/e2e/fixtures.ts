import {
  APIRequestContext,
  expect,
  Page,
  request,
  test as base,
} from "@playwright/test";

export const API_URL = process.env.API_URL ?? "http://localhost:8080";

export interface TestProduct {
  id: string;
  sku: string;
  name: string;
  price: number;
  stock: number;
}

export interface ApiTransaction {
  ID: string;
  InvoiceNumber: string;
  Items?: ApiTransactionItem[];
  Subtotal: number;
  Total: number;
  PaidAmount: number;
  Change: number;
  PaymentMethod: string;
  Status: string;
}

export interface ApiTransactionItem {
  ProductID: string;
  SKU: string;
  Name: string;
  Quantity: number;
  UnitPrice: number;
  Subtotal: number;
}

export function formatIDR(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function uniqueSuffix(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export async function createApiContext(): Promise<APIRequestContext> {
  return request.newContext({ baseURL: API_URL });
}

export async function createTestProduct(
  api: APIRequestContext,
  options?: { price?: number; stock?: number; suffix?: string },
): Promise<TestProduct> {
  const suffix = options?.suffix ?? uniqueSuffix();
  const response = await api.post("/products", {
    data: {
      sku: `E2E-PW-${suffix}`,
      name: `E2E Product ${suffix}`,
      price: options?.price ?? 15_000,
      stock: options?.stock ?? 50,
    },
  });

  expect(response.ok(), `create product failed: ${await response.text()}`).toBeTruthy();
  return response.json() as Promise<TestProduct>;
}

export async function deleteTestProduct(
  api: APIRequestContext,
  productId: string,
): Promise<void> {
  const response = await api.delete(`/products/${productId}`);
  if (!response.ok()) {
    // Products used in checkout may be referenced by transaction_items.
    console.warn(
      `Could not delete test product ${productId}: ${await response.text()}`,
    );
  }
}

export async function getProduct(
  api: APIRequestContext,
  productId: string,
): Promise<TestProduct> {
  const response = await api.get(`/products/${productId}`);
  expect(response.ok(), `get product failed: ${await response.text()}`).toBeTruthy();
  return response.json() as Promise<TestProduct>;
}

export async function getTransaction(
  api: APIRequestContext,
  transactionId: string,
): Promise<ApiTransaction> {
  const response = await api.get(`/transactions/${transactionId}`);
  expect(
    response.ok(),
    `get transaction failed: ${await response.text()}`,
  ).toBeTruthy();
  return response.json() as Promise<ApiTransaction>;
}

export async function waitForProductsLoaded(page: Page): Promise<void> {
  await expect(page.getByText("Memuat produk...")).toBeHidden();
  await expect(page.getByRole("heading", { name: "Produk", level: 2 })).toBeVisible();
}

export function productCardButton(page: Page, productName: string) {
  return page.getByRole("button", {
    name: `Tambah ${productName} ke keranjang`,
  });
}

export function cartIncreaseButton(page: Page, productName: string) {
  return page.getByRole("button", { name: `Tambah ${productName}`, exact: true });
}

export function totalRow(page: Page) {
  return page.getByText("Total", { exact: true }).locator("..");
}

export async function createAndTrackProduct(
  api: APIRequestContext,
  testProducts: TestProduct[],
  options?: { price?: number; stock?: number; suffix?: string },
): Promise<TestProduct> {
  const product = await createTestProduct(api, options);
  testProducts.push(product);
  return product;
}

export const test = base.extend<{
  api: APIRequestContext;
  testProducts: TestProduct[];
}>({
  api: async ({}, use) => {
    const api = await createApiContext();
    await use(api);
    await api.dispose();
  },
  testProducts: async ({ api }, use) => {
    const created: TestProduct[] = [];
    await use(created);
    for (const product of created) {
      await deleteTestProduct(api, product.id);
    }
  },
});

export { expect } from "@playwright/test";
