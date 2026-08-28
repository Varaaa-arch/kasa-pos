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
    <div className="space-y-4">
      <div>
        <Label htmlFor="paid-amount" className="mb-1.5 block text-sm font-medium text-foreground">
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
          className="text-lg font-semibold tabular-nums transition-all duration-200 focus:ring-2 focus:ring-primary/20"
          aria-describedby="change-amount"
        />
      </div>

      <div className="flex items-center justify-between text-sm p-3 rounded-lg bg-muted/50">
        <span className="text-muted-foreground font-medium">Kembali</span>
        <span
          id="change-amount"
          className="font-semibold tabular-nums text-foreground"
        >
          {formatIDR(change > 0 ? change : 0)}
        </span>
      </div>

      {insufficient && (
        <Alert variant="destructive" className="animate-slide-up">
          <AlertDescription>Pembayaran kurang</AlertDescription>
        </Alert>
      )}
    </div>
  );
}
