package domain

import "errors"

// ErrInvalidStrategy is returned for unsupported strategies.
var ErrInvalidStrategy = errors.New("invalid strategy")

// ErrEmptyOrderBook is returned when asks or bids are missing.
var ErrEmptyOrderBook = errors.New("order book is empty")

// ErrIndexOutOfBounds is returned when the index exceeds the order book length.
var ErrIndexOutOfBounds = errors.New("index out of bounds")

// ErrInvalidRange is returned when N > M in avgNM.
var ErrInvalidRange = errors.New("invalid range: n must be <= m")
