package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	controller Controller
}

func NewResolver() (*Resolver, error) {

	controller, err := NewController()
	if err != nil {
		return nil, err
	}

	r := &Resolver{
		controller: controller,
	}

	return r, nil
}
