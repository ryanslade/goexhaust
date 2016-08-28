package main

import "fmt"

type Food int

const (
	Apple Food = iota
	Grape
)

type Drink string

const (
	Beer Drink = "Beer"
	Wine Drink = "Wine"
)

const (
	Water Drink = "Water"
)

const (
	Life  = 42
	Cards = 52
)

func exhaustive(f Food) {
	switch f {
	case Apple:
		fmt.Println("Apple")
	case Grape:
		fmt.Println("Grape")
	}
}

func notExhaustive(f Food) {
	switch f {
	case Apple:
		fmt.Println("Apple")
	}
}
