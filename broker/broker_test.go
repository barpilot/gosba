package broker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/barpilot/gosba/api"
	fakeAPI "github.com/barpilot/gosba/api/fake"
	"github.com/barpilot/gosba/http/filter"
	"github.com/barpilot/gosba/service"
	fakeAsync "github.com/deis/async/fake"
	"github.com/stretchr/testify/assert"
)

var errSome = errors.New("an error")

func TestBrokerStartBlocksUntilAsyncEngineErrors(t *testing.T) {
	apiServerStopped := false
	svr := fakeAPI.NewServer()
	svr.RunBehavior = func(ctx context.Context) error {
		<-ctx.Done()
		apiServerStopped = true
		return ctx.Err()
	}
	e := fakeAsync.NewEngine()
	e.RunBehavior = func(context.Context) error {
		return errSome
	}
	b, err := getTestBroker()
	assert.Nil(t, err)
	b.asyncEngine = e
	b.apiServer = svr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = b.Run(ctx)
	assert.Equal(t, &errAsyncEngineStopped{err: errSome}, err)
	time.Sleep(time.Second)
	assert.True(t, apiServerStopped)
}

func TestBrokerStartBlocksUntilAsyncEngineReturns(t *testing.T) {
	apiServerStopped := false
	svr := fakeAPI.NewServer()
	svr.RunBehavior = func(ctx context.Context) error {
		<-ctx.Done()
		apiServerStopped = true
		return ctx.Err()
	}
	e := fakeAsync.NewEngine()
	e.RunBehavior = func(context.Context) error {
		return nil
	}
	b, err := getTestBroker()
	assert.Nil(t, err)
	b.asyncEngine = e
	b.apiServer = svr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = b.Run(ctx)
	assert.Equal(t, &errAsyncEngineStopped{}, err)
	time.Sleep(time.Second)
	assert.True(t, apiServerStopped)
}

func TestBrokerStartBlocksUntilAPIServerErrors(t *testing.T) {
	svr := fakeAPI.NewServer()
	svr.RunBehavior = func(context.Context) error {
		return errSome
	}
	asyncEngineStopped := false
	e := fakeAsync.NewEngine()
	e.RunBehavior = func(ctx context.Context) error {
		<-ctx.Done()
		asyncEngineStopped = true
		return ctx.Err()
	}
	b, err := getTestBroker()
	assert.Nil(t, err)
	b.asyncEngine = e
	b.apiServer = svr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = b.Run(ctx)
	assert.Equal(t, &errAPIServerStopped{err: errSome}, err)
	time.Sleep(time.Second)
	assert.True(t, asyncEngineStopped)
}

func TestBrokerStartBlocksUntilAPIServerReturns(t *testing.T) {
	svr := fakeAPI.NewServer()
	svr.RunBehavior = func(context.Context) error {
		return nil
	}
	asyncEngineStopped := false
	e := fakeAsync.NewEngine()
	e.RunBehavior = func(ctx context.Context) error {
		<-ctx.Done()
		asyncEngineStopped = true
		return ctx.Err()
	}
	b, err := getTestBroker()
	assert.Nil(t, err)
	b.asyncEngine = e
	b.apiServer = svr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = b.Run(ctx)
	assert.Equal(t, &errAPIServerStopped{}, err)
	time.Sleep(time.Second)
	assert.True(t, asyncEngineStopped)
}

func TestBrokerStartBlocksUntilContextCanceled(t *testing.T) {
	apiServerStopped := false
	svr := fakeAPI.NewServer()
	svr.RunBehavior = func(ctx context.Context) error {
		<-ctx.Done()
		apiServerStopped = true
		return ctx.Err()
	}
	asyncEngineStopped := false
	e := fakeAsync.NewEngine()
	e.RunBehavior = func(ctx context.Context) error {
		<-ctx.Done()
		asyncEngineStopped = true
		return ctx.Err()
	}
	b, err := getTestBroker()
	assert.Nil(t, err)
	b.asyncEngine = e
	b.apiServer = svr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = b.Run(ctx)
	assert.Equal(t, ctx.Err(), err)
	time.Sleep(time.Second)
	assert.True(t, apiServerStopped)
	assert.True(t, asyncEngineStopped)
}

func getTestBroker() (*broker, error) {
	asyncEngine := fakeAsync.NewEngine()
	catalog := service.NewCatalog(nil)
	apiServer, err := api.NewServer(
		api.NewConfigWithDefaults(),
		nil,
		asyncEngine,
		filter.NewChain(),
		catalog,
	)
	if err != nil {
		return nil, err
	}
	b, err := NewBroker(
		apiServer,
		asyncEngine,
		nil,
		catalog,
	)
	if err != nil {
		return nil, err
	}
	return b.(*broker), nil
}
