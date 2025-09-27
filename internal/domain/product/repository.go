package product

type Repository interface {
	Save(p *Product) error
	FindAll() ([]Product, error)
}
