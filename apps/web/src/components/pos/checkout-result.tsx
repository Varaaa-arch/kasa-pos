import { CheckCircle2 } from "lucide-react";
import { formatIDR } from "@/lib/currency";
import type { CheckoutResponse } from "@/types/checkout";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface CheckoutResultProps {
  result: CheckoutResponse;
}

export function CheckoutResult({ result }: CheckoutResultProps) {
  return (
    <Alert className="border-green-200 bg-green-50 text-green-900">
      <CheckCircle2 className="text-green-600" />
      <AlertTitle>Transaksi berhasil</AlertTitle>
      <AlertDescription className="space-y-1">
        <p>ID transaksi: {result.id}</p>
        <p>No. invoice: {result.invoice_number}</p>
        <p>Kembalian: {formatIDR(result.change)}</p>
      </AlertDescription>
    </Alert>
  );
}
