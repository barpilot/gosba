package service

// ProvisioningParameters wraps a map containing provisioning parameters.
type ProvisioningParameters struct {
	Parameters
}

// InstanceDetails is an alias for the emoty interface. It exists only to
// improve the clarity of function signatures and documentation.
type InstanceDetails interface{}

// BindingParameters wraps a map containing binding parameters.
type BindingParameters struct {
	Parameters
}

// BindingDetails is an alias for the empty interface. It exists only to improve
// the clarity of function signatures and documentation.
type BindingDetails interface{}

// Credentials is an interface to be implemented by service-specific types
// that represent service credentials. This interface doesn't require any
// functions to be implemented. It exists to improve the clarity of function
// signatures and documentation.
type Credentials interface{}
