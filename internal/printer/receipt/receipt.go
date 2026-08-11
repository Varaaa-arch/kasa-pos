package receipt

func (i Item) Total() int64 {
	if i.SubTotal != 0 {
		return i.SubTotal
	}

	return int64(i.Quantity) * i.UnitPrice
}

func (r Receipt) Subtotal() int64 {
	if r.Summary.SubTotal != 0 {
		return r.Summary.SubTotal
	}

	var total int64

	for _, item := range r.Items {
		total += item.Total()
	}

	return total
}

func (r Receipt) Change() int64 {
	if r.Payment.Change != 0 {
		return r.Payment.Change
	}

	return r.Payment.Paid - r.Subtotal()
}
