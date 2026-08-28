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
      <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground transition-colors" />
      <Input
        id="product-search"
        type="search"
        placeholder="Cari produk..."
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="pl-9 transition-all duration-200 focus:ring-2 focus:ring-primary/20"
        autoComplete="off"
      />
    </div>
  );
}
