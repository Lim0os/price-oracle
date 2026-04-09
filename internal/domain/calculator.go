package domain

// StrategyCalculator defines a rate calculation strategy.
type StrategyCalculator interface {
	Calculate(book *OrderBook) (ask, bid string, err error)
}
