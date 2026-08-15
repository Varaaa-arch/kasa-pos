import { formatIDR } from "@/lib/currency";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface PaymentPanelProps {
  paidAmount: number;
  total: number;
  onPaidAmountChange: (amount: number) => void;
  disabled: boolean;
}

export function PaymentPanel({
  paidAmount,
  total,
  onPaidAmountChange,
  disabled,
}: PaymentPanelProps) {
  const change = paidAmount - total;
  const insufficient = paidAmount > 0 && paidAmount < total;

  const handleChange = (value: string) => {
    const digits = value.replace(/\D/g, "");
    onPaidAmountChange(digits ? parseInt(digits, 10) : 0);
  };

  return (
    <div className="space-y-3">
      <div>
        <Label htmlFor="paid-amount" className="mb-1.5 block text-sm font-medium">
          Bayar
        </Label>
        <Input
          id="paid-amount"
          type="text"
          inputMode="numeric"
          placeholder="Rp0"
          value={paidAmount > 0 ? paidAmount.toString() : ""}
          onChange={(e) => handleChange(e.target.value)}
          disabled={disabled}
          className="text-lg font-semibold tabular-nums"
          aria-describedby="change-amount"
        />
      </div>

      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">Kembali</span>
        <span
          id="change-amount"
          className="font-semibold tabular-nums"
        >
          {formatIDR(change > 0 ? change : 0)}
        </span>
      </div>

      {insufficient && (
        <Alert variant="destructive">
          <AlertDescription>Pembayaran kurang</AlertDescription>
        </Alert>
      )}
    </div>
  );
}
