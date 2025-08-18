package common

import "slices"

// Keyword classes represented as bits
const (
	HOME_KITCHEN_PETS    = 1 << iota
	GROCERY_GOURMET_FOODS
	MENS_SHOES
	KIDS_FASHION
	WOMENS_SHOES
	ACCESSORIES
	BAGS_LUGGAGE
	INDUSTRIAL_SUPPLIES
	STORES
	MENS_CLOTHING
	WOMENS_CLOTHING
	TV_AUDIO_CAMERAS
	BEAUTY_HEALTH
	HOME_AND_KITCHEN
	PET_SUPPLIES
	MUSIC
	TOYS_AND_BABY_PRODUCTS
	SPORTS_AND_FITNESS
	CAR_AND_MOTORBIKE
	APPLIANCES
)

// Extract keywords from a search request
func GetClassBits(keywords []string) uint64 {
	var bits uint64
	if slices.Contains(keywords, "Home, Kitchen, Pets") {
		bits |= HOME_KITCHEN_PETS
	}
	if slices.Contains(keywords, "Grocery & Gourmet Foods") {
		bits |= GROCERY_GOURMET_FOODS
	}
	if slices.Contains(keywords, "Men's Shoes") {
		bits |= MENS_SHOES
	}
	if slices.Contains(keywords, "Kids's Fashion") {
		bits |= KIDS_FASHION
	}
	if slices.Contains(keywords, "Women's Shoes") {
		bits |= WOMENS_SHOES
	}
	if slices.Contains(keywords, "Accessories") {
		bits |= ACCESSORIES
	}
	if slices.Contains(keywords, "Bags & Luggage") {
		bits |= BAGS_LUGGAGE
	}
	if slices.Contains(keywords, "Industrial Supplies") {
		bits |= INDUSTRIAL_SUPPLIES
	}
	if slices.Contains(keywords, "Stores") {
		bits |= STORES
	}
	if slices.Contains(keywords, "Men's Clothing") {
		bits |= MENS_CLOTHING
	}
	if slices.Contains(keywords, "Women's Clothing") {
		bits |= WOMENS_CLOTHING
	}
	if slices.Contains(keywords, "TV, Audio & Cameras") {
		bits |= TV_AUDIO_CAMERAS
	}
	if slices.Contains(keywords, "Beauty & Health") {
		bits |= BEAUTY_HEALTH
	}
	if slices.Contains(keywords, "Home & Kitchen") {
		bits |= HOME_AND_KITCHEN
	}
	if slices.Contains(keywords, "Pet Supplies") {
		bits |= PET_SUPPLIES
	}
	if slices.Contains(keywords, "Music") {
		bits |= MUSIC
	}
	if slices.Contains(keywords, "Toys & Baby Products") {
		bits |= TOYS_AND_BABY_PRODUCTS
	}
	if slices.Contains(keywords, "Sports & Fitness") {
		bits |= SPORTS_AND_FITNESS
	}
	if slices.Contains(keywords, "Car & Motorbike") {
		bits |= CAR_AND_MOTORBIKE
	}
	if slices.Contains(keywords, "Appliances") {
		bits |= APPLIANCES
	}
	
	return bits
}
