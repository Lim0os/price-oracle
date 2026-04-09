package domain

func validateOrderBook(book *OrderBook) error {
	if book == nil {
		return ErrEmptyOrderBook
	}
	if len(book.Asks) == 0 || len(book.Bids) == 0 {
		return ErrEmptyOrderBook
	}
	return nil
}
