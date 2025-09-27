package csv

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var adjectives = []string{
	"Super", "Deluxe", "Eco", "Smart", "Premium", "Classic", "Ultra", "Golden",
}

var materials = []string{
	"Wooden", "Metal", "Plastic", "Glass", "Leather", "Silk", "Cotton", "Steel",
}

var nouns = []string{
	"Widget", "Chair", "Bottle", "Lamp", "Table", "Bag", "Shoe", "Phone",
}

var suffixes = []string{
	"Pro", "X", "Edition", "2025", "Limited", "Max",
}

var categories = []string{
	"electronics", "furniture", "fashion", "kitchen", "sport", "toys",
}

var allTags = []string{
	"new", "sale", "popular", "eco-friendly", "limited", "trending", "handmade",
}

func GenerateCSV(path string, total int) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	writer.Write([]string{"name", "price", "category", "tags"})

	for i := 0; i < total; i++ {
		name := adjectives[r.Intn(len(adjectives))]

		if r.Intn(2) == 1 { // 50% chance pakai material
			name += " " + materials[r.Intn(len(materials))]
		}

		name += " " + nouns[r.Intn(len(nouns))]

		if r.Intn(3) == 1 { // 33% chance pakai suffix
			name += " " + suffixes[r.Intn(len(suffixes))]
		}

		price := 50 + r.Float64()*(500-50)
		priceStr := strconv.FormatFloat(price, 'f', 2, 64)

		category := categories[r.Intn(len(categories))]

		tagCount := r.Intn(3) + 1
		tagSet := make(map[string]struct{})
		for len(tagSet) < tagCount {
			tagSet[allTags[r.Intn(len(allTags))]] = struct{}{}
		}
		var tags []string
		for t := range tagSet {
			tags = append(tags, t)
		}

		tagStr := strings.Join(tags, ",")
		writer.Write([]string{name, priceStr, category, tagStr})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV: %w", err)
	}

	fmt.Printf("File %s berhasil dibuat dengan %d produk\n", path, total)
	return nil
}
