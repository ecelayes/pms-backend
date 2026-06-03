package vo

import (
	"errors"
	"fmt"
)

var (
	ErrMismatchingCurrency = errors.New("cannot operate on money with different currencies")
)

type Money struct {
	amount   int64
	currency string
}

func NewMoney(amount int64, currency string) Money {
	return Money{
		amount:   amount,
		currency: currency,
	}
}
func (m Money) Amount() int64 {
	return m.amount
}
func (m Money) Currency() string {
	return m.currency
}
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrMismatchingCurrency
	}
	return NewMoney(m.amount+other.amount, m.currency), nil
}
func (m Money) Sub(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrMismatchingCurrency
	}
	return NewMoney(m.amount-other.amount, m.currency), nil
}
func (m Money) Multiply(factor int64) Money {
	return NewMoney(m.amount*factor, m.currency)
}
func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.amount, m.currency)
}
