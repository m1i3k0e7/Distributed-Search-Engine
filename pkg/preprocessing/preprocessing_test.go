package preprocessing

import (
	"testing"
	"fmt"
)

func TestPunctuationRemoval(t *testing.T) {
	tests := []string {
		"Lloyd 1.5 Ton 3 Star Inverter Split AC (5 in 1 Convertible, Copper, Anti-Viral + PM 2.5 Filter, White, GLS18I3FWAMC)",
		"LG 1.5 Ton 5 Star AI DUAL Inverter Split AC (Copper, Super Convertible 6-in-1 Cooling, HD Filter with Anti-Virus Protection, 2023 Model, RS-Q19YNZE, White)",
		"LG 1 Ton 4 Star AI DUAL Inverter Split AC (Copper, AI Convertible 6-in-1 Cooling, HD Filter with Anti Virus Protection, RS-Q13JNYE, White)",
		"LG 1.5 Ton 3 Star AI DUAL Inverter Split AC (Copper, Super Convertible 6-in-1 Cooling, HD Filter with Anti-Virus Protection, 2023 Model, RS-Q19JNXE, White)",
		"Carrier 1.5 Ton 3 Star Inverter Split AC (Copper,ESTER Dxi, 4-in-1 Flexicool Inverter, 2022 Model,R32,White)",
		"Lifelong LLMG23 Power Pro 500-Watt Mixer Grinder with 3 Jars (Liquidizing, Wet Grinding and Chutney Jar), Stainless Steel ...",
	}

	for _, test := range tests {
		result := PreprocessForLargeDataset(test)
		fmt.Println("Original:", test)
		for i, word := range result {
			fmt.Printf("Word %d: %s\n", i+1, word)
		}
	}
}
