package product

type Product struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:100"`
	Price    float64
	Category string `gorm:"size:200"`
	Tags     string `gorm:"size:200"`
}
