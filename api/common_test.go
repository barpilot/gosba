package api

import (
	"fmt"

	"github.com/barpilot/gosba/http/filter"
	"github.com/barpilot/gosba/services/fake"
	memoryStorage "github.com/barpilot/gosba/storage/memory"
	fakeAsync "github.com/deis/async/fake"
	uuid "github.com/satori/go.uuid"
)

type ArbitraryType struct {
	Foo string `json:"foo"`
}

const fooValue = "bar"

var (
	testArbitraryObject = &ArbitraryType{
		Foo: fooValue,
	}
	testArbitraryObjectJSON = []byte(fmt.Sprintf(`{"foo":"%s"}`, fooValue))
	testArbitraryMap        = map[string]interface{}{
		"foo": "bar",
	}
	testArbitraryMapJSON = []byte(fmt.Sprintf(`{"foo":"%s"}`, fooValue))
)

func getDisposableInstanceID() string {
	return uuid.NewV4().String()
}

func getDisposableServiceID() string {
	return uuid.NewV4().String()
}

func getDisposablePlanID() string {
	return uuid.NewV4().String()
}

func getDisposableBindingID() string {
	return uuid.NewV4().String()
}

func getTestServer() (*server, *fake.Module, error) {
	fakeModule, err := fake.New()
	if err != nil {
		return nil, nil, err
	}
	fakeCatalog, err := fakeModule.GetCatalog()
	if err != nil {
		return nil, nil, err
	}
	s, err := NewServer(
		NewConfigWithDefaults(),
		memoryStorage.NewStore(fakeCatalog),
		fakeAsync.NewEngine(),
		filter.NewChain(),
		fakeCatalog,
	)
	if err != nil {
		return nil, nil, err
	}
	return s.(*server), fakeModule, nil
}
