package fake

import "github.com/barpilot/gosba/service"

const (
	// ServiceID is the service ID of the fake service
	ServiceID = "cdd1fb7a-d1e9-49e0-b195-e0bab747798a"
	// StandardPlanID is the plan ID for the standard (and only) variant of the
	// fake service
	StandardPlanID = "bd15e6f3-4ff5-477c-bb57-26313a368e74"
)

// GetCatalog returns a Catalog of service/plans offered by a module
func (m *Module) GetCatalog() (service.Catalog, error) {
	return service.NewCatalog([]service.Service{
		service.NewService(
			service.ServiceProperties{
				ID:          ServiceID,
				Name:        "fake",
				Description: "Fake Service",
				Metadata: service.ServiceMetadata{
					DisplayName:      "fake",
					ImageURL:         "fake",
					LongDescription:  "Fake Service",
					DocumentationURL: "fake",
					SupportURL:       "fake",
				},
				Bindable: true,
				Tags:     []string{"Fake"},
			},
			m.ServiceManager,
			service.NewPlan(service.PlanProperties{
				ID:          StandardPlanID,
				Name:        "standard",
				Description: "The ONLY sort of fake service-- one that's fake!",
				Free:        false,
				Metadata: service.ServicePlanMetadata{
					DisplayName: "Fake",
					Bullets: []string{"Fake 1",
						"Fake 2",
					},
				},
				Schemas: service.PlanSchemas{
					ServiceInstances: service.InstanceSchemas{
						ProvisioningParametersSchema: service.InputParametersSchema{
							PropertySchemas: map[string]service.PropertySchema{
								"someParameter": &service.StringPropertySchema{},
							},
						},
						UpdatingParametersSchema: service.InputParametersSchema{
							PropertySchemas: map[string]service.PropertySchema{
								"someParameter": &service.StringPropertySchema{},
							},
						},
					},
					ServiceBindings: service.BindingSchemas{
						BindingParametersSchema: service.InputParametersSchema{
							PropertySchemas: map[string]service.PropertySchema{
								"someParameter": &service.StringPropertySchema{},
							},
						},
					},
				},
			}),
		),
	}), nil
}
