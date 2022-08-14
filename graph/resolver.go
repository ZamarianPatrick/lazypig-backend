package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	version    string
	controller Controller
}

func NewResolver(version string) (*Resolver, error) {

	controller, err := NewController(true)
	if err != nil {
		return nil, err
	}

	r := &Resolver{
		controller: controller,
		version:    version,
	}

	return r, nil
}
