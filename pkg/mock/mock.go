package mock

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	"ecommerce/pkg/storage"
	"fmt"
	"io"
	"strings"
	"syreclabs.com/go/faker"
)

type Mock struct {
	DB                 *sql.DB
	W                  io.Writer
	ScriptPaths        []string
	ProductService ecommerce.ProductService
}

func (m *Mock) SeedDB() error {
	err := storage.ExecScripts(m.DB, m.ScriptPaths...)
	if err != nil {
		return err
	}

	fmt.Printf("Creating 10 random categories\n")
	catIDs, err := m.createCategories(10)
	if err != nil {
		return err
	}
	fmt.Printf("\tcategories created with ids: %v\n", catIDs)

	fmt.Println("Creating products")
	numberOfProductPerCat := 10
	err = m.createProducts(catIDs, numberOfProductPerCat)
	if err != nil {
		return err
	}
	fmt.Printf("\t %d products created each for %d categories\n", numberOfProductPerCat, len(catIDs))

	return nil
}

func (m *Mock) createCategories(number int) ([]int, error) {
	var ids []int

	for i := 0; i < number; i++ {
		name := faker.Commerce().Department()
		id, err := m.ProductService.CreateCategory(name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (m *Mock) createProducts(categoryIDs []int, number int) error {
	for _, catID := range categoryIDs {
		for i := 0; i < len(categoryIDs); i++ {
			p := &ecommerce.Product{
				Name:        faker.Commerce().ProductName(),
				CategoryID:  catID,
				Price:       ecommerce.Price{Current: faker.Commerce().Price()},
				Description: strings.Join(faker.Lorem().Paragraphs(3), "\n"),
				Quantity:    faker.RandomInt(10, 1000),
			}
			_, err := m.ProductService.CreateProduct(p)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
