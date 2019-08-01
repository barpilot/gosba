package fake

import "github.com/barpilot/gosba/service"

type fakeInstanceDetails struct {
}

type fakeBindingDetails struct {
}

// GetEmptyInstanceDetails returns an "empty" service-specific object that
// can be populated with data during unmarshaling of JSON to an Instance
func (s *ServiceManager) GetEmptyInstanceDetails() service.InstanceDetails {
	return GetEmptyInstanceDetails()
}

// GetEmptyInstanceDetails is invoked in testing.
func GetEmptyInstanceDetails() service.InstanceDetails {
	return fakeInstanceDetails{}
}

// GetEmptyBindingDetails returns an "empty" service-specific object that
// can be populated with data during unmarshaling of JSON to a Binding
func (s *ServiceManager) GetEmptyBindingDetails() service.BindingDetails {
	return GetEmptyBindingDetails()
}

// GetEmptyBindingDetails is invoked in testing.
func GetEmptyBindingDetails() service.BindingDetails {
	return fakeBindingDetails{}
}
