package main

import (
	"GOcoolprop/pkg/props"
	"fmt"
)

func main() {
	// Test Water at 300K, 101325 Pa
	fmt.Println("=== Water ===")
	rho, err := props.PropSI("D", "T", 300, "P", 101325, "Water")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Density: %v mol/m3\n", rho)

		// Get other properties
		h, _ := props.PropSI("H", "T", 300, "P", 101325, "Water")
		s, _ := props.PropSI("S", "T", 300, "P", 101325, "Water")
		fmt.Printf("Enthalpy: %v J/mol\n", h)
		fmt.Printf("Entropy: %v J/mol/K\n", s)
	}

	// Test Nitrogen at 300K, 101325 Pa (gas)
	fmt.Println("\n=== Nitrogen ===")
	rho, err = props.PropSI("D", "T", 300, "P", 101325, "Nitrogen")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Density: %v mol/m3\n", rho)

		// Verify pressure
		p, _ := props.PropSI("P", "T", 300, "D", rho, "Nitrogen")
		fmt.Printf("Verification - Pressure: %v Pa\n", p)
	}

	// Test Hydrogen at 300K, 101325 Pa (gas)
	fmt.Println("\n=== Hydrogen ===")
	rho, err = props.PropSI("D", "T", 300, "P", 101325, "Hydrogen")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Density: %v mol/m3\n", rho)

		// Verify pressure
		p, _ := props.PropSI("P", "T", 300, "D", rho, "Hydrogen")
		fmt.Printf("Verification - Pressure: %v Pa\n", p)
	}
}
