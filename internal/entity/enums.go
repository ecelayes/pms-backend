package entity

type MealType int

const (
	MealTypeRoomOnly    MealType = iota
	MealTypeContinental
	MealTypeBuffet
	MealTypeAmerican
	MealTypeHalfBoard
	MealTypeFullBoard
)

type PenaltyType int

const (
	PenaltyFixedAmount PenaltyType = iota
	PenaltyPercentage
	PenaltyNights
)

type PaymentTiming int

const (
	PayOnArrival PaymentTiming = iota
	PayPrepaid
)

type PaymentMethod int

const (
	PaymentMethodNone          PaymentMethod = iota
	PaymentMethodCreditCard
	PaymentMethodBankTransfer
)
