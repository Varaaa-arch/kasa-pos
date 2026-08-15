import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface ProductSearchProps {
  value: string;
  onChange: (value: string) => void;
}

export function ProductSearch({ value, onChange }: ProductSearchProps) {
  return (
    <div className="relative">
      <Label htmlFor="product-search" className="sr-only">
        Cari produk
      </Label>
      <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        id="product-search"
        type="search"
        placeholder="Cari produk..."
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="pl-9"
        autoComplete="off"
      />
    </div>
  );
}
