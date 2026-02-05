package calculator

import "jombG/goblast/example/shop"

func Add(a, b int) int {
	return a + b
}

func Subtract(a, b int) int {
	return a - b
}

func Multiply(a, b int) int {
	return a * b
}

func Divide(a, b int) int {
	return a / b
}

func AddPriceProdcut(pr1, pr2 shop.Product) shop.Product {
	pr1.AddPrice(pr2.Price)
	return pr1
}
