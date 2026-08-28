import type { Product } from "@/types/product";
import type { CheckoutRequest, CheckoutResponse } from "@/types/checkout";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (response.ok) {
    return response.json() as Promise<T>;
  }

  const text = await response.text();
  const message = mapErrorMessage(response.status, text);
  throw new ApiError(message, response.status);
}

function mapErrorMessage(status: number, body: string): string {
  const normalized = body.trim().toLowerCase();

  if (normalized.includes("stock tidak cukup") || normalized.includes("insufficient stock")) {
    return "Stock tidak cukup";
  }
  if (
    normalized.includes("pembayaran kurang") ||
    normalized.includes("insufficient cash") ||
    normalized.includes("payment is insufficient")
  ) {
    return "Pembayaran kurang";
  }
  if (normalized.includes("cart is empty")) {
    return "Keranjang kosong";
  }
  if (normalized.includes("product not found")) {
    return "Produk tidak ditemukan";
  }
  if (normalized.includes("transaksi gagal")) {
    return "Transaksi gagal";
  }
  if (body.trim()) {
    return body.trim();
  }
  if (status === 0 || status >= 500) {
    return "Server tidak tersedia";
  }
  return "Terjadi kesalahan";
}

export async function fetchProducts(): Promise<Product[]> {
  if (!API_BASE_URL) {
    throw new ApiError("Server tidak tersedia", 0);
  }

  const response = await fetch(`${API_BASE_URL}/products`, {
    method: "GET",
    headers: { Accept: "application/json" },
  });

  return handleResponse<Product[]>(response);
}

export async function processCheckout(request: CheckoutRequest): Promise<CheckoutResponse> {
  if (!API_BASE_URL) {
    throw new ApiError("Server tidak tersedia", 0);
  }

  const response = await fetch(`${API_BASE_URL}/checkout`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify(request),
  });

  return handleResponse<CheckoutResponse>(response);
}
