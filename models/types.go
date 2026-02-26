package models

import (
	"fmt"
)

// Rupees is a custom type wrapping float64 for currency amounts.
type Rupees float64

func (r Rupees) String() string {
	return fmt.Sprintf("%.2f", float64(r))
}

func (r Rupees) ToFloat64() float64 {
	return float64(r)
}

// Category represents a transaction category.
type Category string

const (
	CategoryFood          Category = "Food"
	CategoryTravel        Category = "Travel"
	CategoryShopping      Category = "Shopping"
	CategoryBills         Category = "Bills"
	CategoryHealth        Category = "Health"
	CategoryEducation     Category = "Education"
	CategoryEntertainment Category = "Entertainment"
	CategorySalary        Category = "Salary"
	CategoryFreelance     Category = "Freelance"
	CategoryInvestment    Category = "Investment"
	CategoryOther         Category = "Other"
)

// TransactionType indicates income or expense.
type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)
