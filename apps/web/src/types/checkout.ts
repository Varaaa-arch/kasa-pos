export interface CheckoutRequestItem {
  product_id: string;
  sku: string;
  name: string;
  quantity: number;
  unit_price: number;
}

export interface CheckoutRequest {
  items: CheckoutRequestItem[];
  discount: number;
  tax: number;
  service_charge: number;
  paid_amount: number;
  payment_method: string;
  invoice_number?: string;
}

export interface CheckoutPrintJob {
  id: string;
  status: string;
  error?: string;
}

export interface CheckoutResponse {
  id: string;
  invoice_number: string;
  subtotal: number;
  discount: number;
  tax: number;
  service_charge: number;
  total: number;
  paid_amount: number;
  change: number;
  payment_method: string;
  status: string;
  created_at: string;
  print_job?: CheckoutPrintJob;
}
