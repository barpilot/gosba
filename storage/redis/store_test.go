package redis

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"testing"

	"github.com/barpilot/gosba/service"
	"github.com/barpilot/gosba/services/fake"
	"github.com/go-redis/redis"
	"github.com/ory/dockertest"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// var (
// 	fakeServiceManager service.ServiceManager
// 	testStore          *store
// 	config             Config
// )

// func init() {
// 	var err error
// 	fakeModule, err := fake.New()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fakeCatalog, err := fakeModule.GetCatalog()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fakeServiceManager = fakeModule.ServiceManager
// 	config = NewConfigWithDefaults()
// 	config.RedisHost = os.Getenv("STORAGE_REDIS_HOST")
// 	config.RedisPrefix = uuid.NewV4().String()
// 	str, err := NewStore(
// 		fakeCatalog,
// 		config,
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	testStore = str.(*store)
// }

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type StorageTestSuite struct {
	suite.Suite
	fakeServiceManager service.ServiceManager
	testStore          *store
	config             Config
	pool               *dockertest.Pool
	resource           *dockertest.Resource
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *StorageTestSuite) SetupSuite() {
	var err error
	fakeModule, err := fake.New()
	if err != nil {
		log.Fatal(err)
	}
	fakeCatalog, err := fakeModule.GetCatalog()
	if err != nil {
		log.Fatal(err)
	}
	suite.fakeServiceManager = fakeModule.ServiceManager
	suite.config = NewConfigWithDefaults()
	// suite.config.RedisHost = os.Getenv("STORAGE_REDIS_HOST")
	suite.config.RedisPrefix = uuid.NewV4().String()
	if err != nil {
		log.Fatal(err)
	}

	suite.pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	suite.resource, err = suite.pool.Run("redis", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	suite.resource.Expire(60)

	// if run with docker-machine the hostname needs to be set
	u, err := url.Parse(suite.pool.Client.Endpoint())
	if err != nil {
		log.Fatalf("Could not parse endpoint: %s", suite.pool.Client.Endpoint())
	}
	if u.Hostname() == "" {
		// unix socket
		suite.config.RedisHost = "127.0.0.1"
	} else {
		suite.config.RedisHost = u.Hostname()
	}

	port, err := strconv.Atoi(suite.resource.GetPort("6379/tcp"))
	if err != nil {
		log.Fatalf("Could not parse port: %d", port)
	}
	suite.config.RedisPort = port

	str, err := NewStore(
		fakeCatalog,
		suite.config,
	)
	suite.testStore = str.(*store)
}

func (suite *StorageTestSuite) TearDownSuite() {
	if err := suite.pool.Purge(suite.resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

// All methods that begin with "Test" are run as tests within a
// suite.
// func (suite *StorageTestSuite) TestExample() {
// 	assert.Equal(suite.T(), 5, suite.VariableThatShouldStartAtFive)
// 	suite.Equal(5, suite.VariableThatShouldStartAtFive)
// }

func (suite *StorageTestSuite) TestWriteInstance() {
	t := suite.T()
	instance := getTestInstance()
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// First assert that the instance doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Store the instance
	err := suite.testStore.WriteInstance(instance)
	assert.Nil(t, err)
	// Assert that the instance is now in Redis
	strCmd = suite.testStore.redisClient.Get(key)
	assert.Nil(t, strCmd.Err())
	// TODO: krancour: This next assertion only holds true if the Redis DB is
	// reset between test runs. Need to fix this at some point.
	// sCmd := testStore.redisClient.SMembers(instances)
	// assert.Nil(t, strCmd.Err())
	// count, err := sCmd.Result()
	// assert.Nil(t, err)
	// assert.Equal(t, 1, len(count))
	boolCmd := suite.testStore.redisClient.SIsMember(
		wrapKey(
			suite.config.RedisPrefix,
			"instances",
		),
		key,
	)
	assert.Nil(t, boolCmd.Err())
	found, _ := boolCmd.Result()
	assert.True(t, found)
}

func (suite *StorageTestSuite) TestWriteInstanceWithAlias() {
	t := suite.T()
	instance := getTestInstance()
	instance.Alias = uuid.NewV4().String()
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// First assert that the instance doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Nor does its alias
	aliasKey := suite.testStore.getInstanceAliasKey(instance.Alias)
	strCmd = suite.testStore.redisClient.Get(aliasKey)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Store the instance
	err := suite.testStore.WriteInstance(instance)
	assert.Nil(t, err)
	// Assert that the instance is now in Redis
	strCmd = suite.testStore.redisClient.Get(key)
	assert.Nil(t, strCmd.Err())
	// Assert that the alias is as well
	strCmd = suite.testStore.redisClient.Get(aliasKey)
	assert.Nil(t, strCmd.Err())
	instanceID, err := strCmd.Result()
	assert.Nil(t, err)
	assert.Equal(t, instance.InstanceID, instanceID)
}

func (suite *StorageTestSuite) TestWriteInstanceWithParent() {
	t := suite.T()
	instance := getTestInstance()
	instance.ParentAlias = uuid.NewV4().String()
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// First assert that the instance doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Nor does any index of parent alias to children
	parentAliasChildrenKey := suite.testStore.getInstanceAliasChildrenKey(
		instance.ParentAlias,
	)
	boolCmd := suite.testStore.redisClient.SIsMember(parentAliasChildrenKey, instance.InstanceID)
	assert.Nil(t, boolCmd.Err())
	childFoundInIndex, err := boolCmd.Result()
	assert.Nil(t, err)
	assert.False(t, childFoundInIndex)
	// Store the instance
	err = suite.testStore.WriteInstance(instance)
	assert.Nil(t, err)
	// Assert that the instance is now in Redis
	strCmd = suite.testStore.redisClient.Get(key)
	assert.Nil(t, strCmd.Err())
	// And the index for parent alias to children contains this instance
	boolCmd = suite.testStore.redisClient.SIsMember(parentAliasChildrenKey, instance.InstanceID)
	assert.Nil(t, boolCmd.Err())
	childFoundInIndex, err = boolCmd.Result()
	assert.Nil(t, err)
	assert.True(t, childFoundInIndex)
}

func (suite *StorageTestSuite) TestGetNonExistingInstance() {
	t := suite.T()
	instanceID := uuid.NewV4().String()
	key := suite.testStore.getInstanceKey(instanceID)
	// First assert that the instance doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Try to retrieve the non-existing instance
	_, ok, err := suite.testStore.GetInstance(instanceID)
	// Assert that the retrieval failed
	assert.False(t, ok)
	assert.Nil(t, err)
}

func (suite *StorageTestSuite) TestGetExistingInstance() {
	t := suite.T()
	instance := getTestInstance()
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// First ensure the instance exists in Redis
	json, err := instance.ToJSON()
	assert.Nil(t, err)
	statCmd := suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// Retrieve the instance
	retrievedInstance, ok, err := suite.testStore.GetInstance(instance.InstanceID)
	// Assert that the retrieval was successful
	assert.Nil(t, err)
	assert.True(t, ok)
	// Blank out a few fields before we compare
	retrievedInstance.Service = nil
	retrievedInstance.Plan = nil
	assert.Equal(t, instance, retrievedInstance)
}

func (suite *StorageTestSuite) TestGetExistingInstanceWithParent() {
	t := suite.T()
	// Make a parent instance
	parentInstance := getTestInstance()
	parentInstance.Alias = uuid.NewV4().String()
	parentKey := suite.testStore.getInstanceKey(parentInstance.InstanceID)
	// Ensure the parent instance exists in Redis
	json, err := parentInstance.ToJSON()
	assert.Nil(t, err)
	statCmd := suite.testStore.redisClient.Set(parentKey, json, 0)
	assert.Nil(t, statCmd.Err())
	// Ensure the parent instance's alias also exists in Redis
	parentAliasKey := suite.testStore.getInstanceAliasKey(parentInstance.Alias)
	statCmd = suite.testStore.redisClient.Set(parentAliasKey, parentInstance.InstanceID, 0)
	assert.Nil(t, statCmd.Err())
	// Make a child instance
	instance := getTestInstance()
	instance.ParentAlias = parentInstance.Alias
	instance.Parent = &parentInstance
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// Ensure the child instance exists in Redis
	json, err = instance.ToJSON()
	assert.Nil(t, err)
	statCmd = suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// Retrieve the child instance
	retrievedInstance, ok, err := suite.testStore.GetInstance(instance.InstanceID)
	// Assert that the retrieval was successful
	assert.Nil(t, err)
	assert.True(t, ok)
	// Blank out a few fields before we compare
	retrievedInstance.Service = nil
	retrievedInstance.Parent.Service = nil
	retrievedInstance.Plan = nil
	retrievedInstance.Parent.Plan = nil
	assert.Equal(t, instance, retrievedInstance)
}

func (suite *StorageTestSuite) TestGetNonExistingInstanceByAlias() {
	t := suite.T()
	alias := uuid.NewV4().String()
	aliasKey := suite.testStore.getInstanceAliasKey(alias)
	// First assert that the alias doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(aliasKey)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Try to retrieve the non-existing instance by alias
	_, ok, err := suite.testStore.GetInstanceByAlias(aliasKey)
	// Assert that the retrieval failed
	assert.False(t, ok)
	assert.Nil(t, err)
}

func (suite *StorageTestSuite) TestGetExistingInstanceByAlias() {
	t := suite.T()
	instance := getTestInstance()
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// First ensure the instance exists in Redis
	json, err := instance.ToJSON()
	assert.Nil(t, err)
	statCmd := suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// And so does the alias
	aliasKey := suite.testStore.getInstanceAliasKey(instance.Alias)
	statCmd = suite.testStore.redisClient.Set(aliasKey, instance.InstanceID, 0)
	assert.Nil(t, statCmd.Err())
	// Retrieve the instance by alias
	retrievedInstance, ok, err := suite.testStore.GetInstanceByAlias(instance.Alias)
	// Assert that the retrieval was successful
	assert.Nil(t, err)
	assert.True(t, ok)
	// Blank out a few fields before we compare
	retrievedInstance.Service = nil
	retrievedInstance.Plan = nil
	assert.Equal(t, instance, retrievedInstance)
}

func (suite *StorageTestSuite) TestDeleteNonExistingInstance() {
	t := suite.T()
	instanceID := uuid.NewV4().String()
	key := suite.testStore.getInstanceKey(instanceID)
	// First assert that the instance doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Try to delete the non-existing instance
	ok, err := suite.testStore.DeleteInstance(instanceID)
	// Assert that the delete failed
	assert.False(t, ok)
	assert.Nil(t, err)
}

func (suite *StorageTestSuite) TestDeleteExistingInstance() {
	t := suite.T()
	instance := getTestInstance()
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// First ensure the instance exists in Redis
	json, err := instance.ToJSON()
	assert.Nil(t, err)
	statCmd := suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// Delete the instance
	ok, err := suite.testStore.DeleteInstance(instance.InstanceID)
	// Assert that the delete was successful
	assert.True(t, ok)
	assert.Nil(t, err)
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	boolCmd := suite.testStore.redisClient.SIsMember(
		wrapKey(
			suite.config.RedisPrefix,
			"instances",
		),
		key,
	)
	assert.Nil(t, boolCmd.Err())
	found, _ := boolCmd.Result()
	assert.False(t, found)
}

func (suite *StorageTestSuite) TestDeleteExistingInstanceWithAlias() {
	t := suite.T()
	instance := getTestInstance()
	instance.Alias = uuid.NewV4().String()
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// First ensure the instance exists in Redis
	json, err := instance.ToJSON()
	assert.Nil(t, err)
	statCmd := suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// And so does the alias
	aliasKey := suite.testStore.getInstanceAliasKey(instance.Alias)
	statCmd = suite.testStore.redisClient.Set(aliasKey, instance.InstanceID, 0)
	assert.Nil(t, statCmd.Err())
	// Delete the instance
	ok, err := suite.testStore.DeleteInstance(instance.InstanceID)
	// Assert that the delete was successful
	assert.True(t, ok)
	assert.Nil(t, err)
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Assert that the alias is also gone
	strCmd = suite.testStore.redisClient.Get(aliasKey)
	assert.Equal(t, redis.Nil, strCmd.Err())
}

func (suite *StorageTestSuite) TestDeleteExistingInstanceWithParent() {
	t := suite.T()
	// Make a parent instance
	parentInstance := getTestInstance()
	parentKey := suite.testStore.getInstanceKey(parentInstance.InstanceID)
	// Ensure the parent instance exists in Redis
	json, err := parentInstance.ToJSON()
	assert.Nil(t, err)
	statCmd := suite.testStore.redisClient.Set(parentKey, json, 0)
	assert.Nil(t, statCmd.Err())
	// Ensure the parent instance's alias also exists in Redis
	parentAliasKey := suite.testStore.getInstanceAliasKey(parentInstance.Alias)
	statCmd = suite.testStore.redisClient.Set(parentAliasKey, parentInstance.InstanceID, 0)
	assert.Nil(t, statCmd.Err())
	// Make a child instance
	instance := getTestInstance()
	instance.ParentAlias = parentInstance.Alias
	instance.Parent = &parentInstance
	key := suite.testStore.getInstanceKey(instance.InstanceID)
	// Ensure the child instance exists in Redis
	json, err = instance.ToJSON()
	assert.Nil(t, err)
	statCmd = suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// Delete the instance
	ok, err := suite.testStore.DeleteInstance(instance.InstanceID)
	// Assert that the delete was successful
	assert.True(t, ok)
	assert.Nil(t, err)
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// And the index of parent alias to children no longer contains this instance
	parentAliasChildrenKey := suite.testStore.getInstanceAliasChildrenKey(
		instance.ParentAlias,
	)
	boolCmd := suite.testStore.redisClient.SIsMember(parentAliasChildrenKey, instance.InstanceID)
	assert.Nil(t, boolCmd.Err())
	childFoundInIndex, err := boolCmd.Result()
	assert.Nil(t, err)
	assert.False(t, childFoundInIndex)
}

func (suite *StorageTestSuite) TestGetInstanceChildCountByAlias() {
	t := suite.T()
	const count = 5
	instanceAlias := uuid.NewV4().String()
	instanceAliasChildrenKey := suite.testStore.getInstanceAliasChildrenKey(
		instanceAlias,
	)
	for i := 0; i < count; i++ {
		// Add a new, unique, child instance ID to the index
		suite.testStore.redisClient.SAdd(instanceAliasChildrenKey, uuid.NewV4().String())
		// Count the children
		children, err := suite.testStore.GetInstanceChildCountByAlias(instanceAlias)
		assert.Nil(t, err)
		// Assert the size of the index is what we expect
		assert.Equal(t, int64(i+1), children)
	}
}

func (suite *StorageTestSuite) TestWriteBinding() {
	t := suite.T()
	binding := getTestBinding()
	key := suite.testStore.getBindingKey(binding.BindingID)
	// First assert that the binding doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Store the binding
	err := suite.testStore.WriteBinding(binding)
	assert.Nil(t, err)
	// Assert that the binding is now in Redis
	strCmd = suite.testStore.redisClient.Get(key)
	assert.Nil(t, strCmd.Err())
	boolCmd := suite.testStore.redisClient.SIsMember(
		wrapKey(
			suite.config.RedisPrefix,
			"bindings",
		),
		key,
	)
	assert.Nil(t, boolCmd.Err())
	found, _ := boolCmd.Result()
	assert.True(t, found)
}

func (suite *StorageTestSuite) TestGetNonExistingBinding() {
	t := suite.T()
	bindingID := uuid.NewV4().String()
	key := suite.testStore.getBindingKey(bindingID)
	// First assert that the binding doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Try to retrieve the non-existing binding
	_, ok, err := suite.testStore.GetBinding(bindingID)
	// Assert that the retrieval failed
	assert.False(t, ok)
	assert.Nil(t, err)
}

func (suite *StorageTestSuite) TestGetExistingBinding() {
	t := suite.T()
	binding := getTestBinding()
	key := suite.testStore.getBindingKey(binding.BindingID)
	// First ensure the binding exists in Redis
	json, err := binding.ToJSON()
	assert.Nil(t, err)

	statCmd := suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// Retrieve the binding
	retrievedBinding, ok, err := suite.testStore.GetBinding(binding.BindingID)
	// Assert that the retrieval was successful
	assert.True(t, ok)
	assert.Nil(t, err)
	// Blank out a few fields before we compare
	assert.Equal(t, binding, retrievedBinding)
}

func (suite *StorageTestSuite) TestDeleteNonExistingBinding() {
	t := suite.T()
	bindingID := uuid.NewV4().String()
	key := suite.testStore.getBindingKey(bindingID)
	// First assert that the binding doesn't exist in Redis
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
	// Try to delete the non-existing binding
	ok, err := suite.testStore.DeleteBinding(bindingID)
	// Assert that the delete failed
	assert.False(t, ok)
	assert.Nil(t, err)
}

func (suite *StorageTestSuite) TestDeleteExistingBinding() {
	t := suite.T()
	binding := getTestBinding()
	key := suite.testStore.getBindingKey(binding.BindingID)
	// First ensure the binding exists in Redis
	json, err := binding.ToJSON()
	assert.Nil(t, err)
	statCmd := suite.testStore.redisClient.Set(key, json, 0)
	assert.Nil(t, statCmd.Err())
	// Delete the binding
	ok, err := suite.testStore.DeleteBinding(binding.BindingID)
	// Assert that the delete was successful
	assert.True(t, ok)
	assert.Nil(t, err)
	strCmd := suite.testStore.redisClient.Get(key)
	assert.Equal(t, redis.Nil, strCmd.Err())
}

func (suite *StorageTestSuite) TestGetInstanceKey() {
	t := suite.T()
	const rawKey = "foo"
	expected := fmt.Sprintf("%s:instances:%s", suite.config.RedisPrefix, rawKey)
	assert.Equal(t, expected, suite.testStore.getInstanceKey(rawKey))
}

func (suite *StorageTestSuite) TestGetBindingKey() {
	t := suite.T()
	const rawKey = "foo"
	expected := fmt.Sprintf("%s:bindings:%s", suite.config.RedisPrefix, rawKey)
	assert.Equal(t, expected, suite.testStore.getBindingKey(rawKey))
}

func getTestInstance() service.Instance {
	return service.Instance{
		InstanceID:   uuid.NewV4().String(),
		ServiceID:    fake.ServiceID,
		PlanID:       fake.StandardPlanID,
		Status:       service.InstanceStateProvisioned,
		StatusReason: "",
	}
}

func getTestBinding() service.Binding {
	return service.Binding{
		BindingID:    uuid.NewV4().String(),
		InstanceID:   uuid.NewV4().String(),
		ServiceID:    fake.ServiceID,
		Status:       service.BindingStateBound,
		StatusReason: "",
	}
}
