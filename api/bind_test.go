package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/barpilot/gosba/service"
	"github.com/barpilot/gosba/services/fake"
	"github.com/stretchr/testify/assert"
)

func TestBindingWithInstanceThatDoesNotExist(t *testing.T) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	req, err := getBindingRequest(
		getDisposableInstanceID(),
		getDisposableBindingID(),
		nil,
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBindingWithInstanceThatIsNotFullyProvisioned(t *testing.T) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioning,
	})
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		getDisposableBindingID(),
		nil,
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBindingWithServiceIDDifferentFromInstanceServiceID(t *testing.T) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		getDisposableBindingID(),
		&BindingRequest{
			ServiceID: getDisposableServiceID(),
			PlanID:    fake.StandardPlanID,
		},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBindingWithPlanIDDifferentFromInstancePlanID(t *testing.T) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		getDisposableBindingID(),
		&BindingRequest{
			ServiceID: fake.ServiceID,
			PlanID:    getDisposablePlanID(),
		},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBindingModuleNotFoundForServiceID(t *testing.T) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID := getDisposableInstanceID()
	serviceID := getDisposableServiceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  serviceID,
		PlanID:     getDisposablePlanID(),
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		getDisposableBindingID(),
		&BindingRequest{},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBindingWithExistingBindingWithDifferentInstanceID(
	t *testing.T,
) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID1 := getDisposableInstanceID()
	instanceID2 := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID1,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	bindingID := getDisposableBindingID()
	// This binding will already be bound to the first instance
	err = s.store.WriteBinding(service.Binding{
		InstanceID: instanceID1,
		BindingID:  bindingID,
		ServiceID:  fake.ServiceID,
	})
	assert.Nil(t, err)
	// Here's a second instance that we can try to bind to
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID2,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	// This should be a conflict because we're trying to bind to the second
	// instance, but this binding already exists and is bound to a different
	// instance.
	req, err := getBindingRequest(
		instanceID2,
		bindingID,
		&BindingRequest{},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBindingWithExistingBindingWithDifferentParameters(
	t *testing.T,
) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	bindingID := getDisposableBindingID()
	existingBinding := service.Binding{
		InstanceID: instanceID,
		BindingID:  bindingID,
		ServiceID:  fake.ServiceID,
		BindingParameters: &service.BindingParameters{
			Parameters: service.Parameters{
				Schema: &service.InputParametersSchema{
					PropertySchemas: map[string]service.PropertySchema{
						"someParameter": &service.StringPropertySchema{},
					},
				},
				Data: map[string]interface{}{
					"someParameter": "foo",
				},
			},
		},
	}
	err = s.store.WriteBinding(existingBinding)
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		bindingID,
		&BindingRequest{
			Parameters: map[string]interface{}{
				"someParameter": "bar",
			},
		},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBindingWithExistingBoundBindingWithSameAttributes(
	t *testing.T,
) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	bindingID := getDisposableBindingID()
	err = s.store.WriteBinding(service.Binding{
		InstanceID: instanceID,
		BindingID:  bindingID,
		ServiceID:  fake.ServiceID,
		BindingParameters: &service.BindingParameters{
			Parameters: service.Parameters{
				Schema: &service.InputParametersSchema{
					PropertySchemas: map[string]service.PropertySchema{
						"someParameter": &service.StringPropertySchema{},
					},
				},
				Data: map[string]interface{}{
					"someParameter": "foo",
				},
			},
		},
		Status: service.BindingStateBound,
	})
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		bindingID,
		&BindingRequest{
			Parameters: map[string]interface{}{
				"someParameter": "foo",
			},
		},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	// TODO: Test the response body
}

func TestBindingWithExistingFailedBindingWithSameAttributes(
	t *testing.T,
) {
	s, _, err := getTestServer()
	assert.Nil(t, err)
	instanceID := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	bindingID := getDisposableBindingID()
	err = s.store.WriteBinding(service.Binding{
		InstanceID: instanceID,
		BindingID:  bindingID,
		ServiceID:  fake.ServiceID,
		Status:     service.BindingStateBindingFailed,
	})
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		bindingID,
		&BindingRequest{},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Equal(t, responseEmptyJSON, rr.Body.Bytes())
}

func TestBrandNewBinding(t *testing.T) {
	s, m, err := getTestServer()
	assert.Nil(t, err)
	bindCalled := false
	m.ServiceManager.BindBehavior = func(
		service.Instance,
		service.BindingParameters,
	) (service.BindingDetails, error) {
		bindCalled = true
		return nil, nil
	}
	instanceID := getDisposableInstanceID()
	err = s.store.WriteInstance(service.Instance{
		InstanceID: instanceID,
		ServiceID:  fake.ServiceID,
		PlanID:     fake.StandardPlanID,
		Status:     service.InstanceStateProvisioned,
	})
	assert.Nil(t, err)
	req, err := getBindingRequest(
		instanceID,
		getDisposableBindingID(),
		&BindingRequest{},
	)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, bindCalled)
	// TODO: Test the response body
}

func getBindingRequest(
	instanceID string,
	bindingID string,
	br *BindingRequest,
) (*http.Request, error) {
	var body []byte
	if br != nil {
		var err error
		body, err = br.ToJSON()
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf(
			"/v2/service_instances/%s/service_bindings/%s",
			instanceID,
			bindingID,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}
