package db

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type CosmosHandler struct {
	// Cosmos Client
	Client *azcosmos.Client
	// Cosmos database client
	DatabaseClient *azcosmos.DatabaseClient
	// Cosmos container client
	ContainerClient *azcosmos.ContainerClient
	// Data
	DatabaseName  string
	ContainerName string
	PartitionKey  string
}

func NewCosmosHandler(cfg *CosmosConfig) (*CosmosHandler, error) {
	slog.Debug("creating cosmos handler")
	client, err := GetCostmosClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create cosmos client: %v", err)
	}

	return &CosmosHandler{
		Client:          client,
		DatabaseClient:  nil,
		ContainerClient: nil,
		DatabaseName:    cfg.DatabaseName,
		ContainerName:   cfg.ContainerName,
		PartitionKey:    cfg.PartitionKeyPath,
	}, nil
}

func (h *CosmosHandler) Init() error {
	databaseClient, err := CreateDatabase(h.Client, h.DatabaseName)
	if err != nil {
		return fmt.Errorf("failed to create database client: %v", err)
	}
	h.DatabaseClient = databaseClient

	containerClient, err := CreateContainer(h.Client, h.DatabaseName, h.ContainerName, h.PartitionKey)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}
	h.ContainerClient = containerClient

	return nil
}

type Item struct {
	Id            ItemID      `json:"id"`
	LifetimeHours int         `json:"lifetimeHours"`
	Content       ItemContent `json:"content"`
	DeleteOnRead  bool        `json:"deleteOnRead"`
	Created       time.Time   `json:"created"`
}

func GetCostmosClient(cfg *CosmosConfig) (*azcosmos.Client, error) {
	slog.Debug("getting cosmos client")
	cred, err := azcosmos.NewKeyCredential(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create a credential: %v", err)
	}

	// Create a CosmosDB client
	client, err := azcosmos.NewClientWithKey(cfg.Endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Cosmos DB client: %v", err)
	}

	return client, nil
}

func CreateDatabase(client *azcosmos.Client, databaseName string) (*azcosmos.DatabaseClient, error) {
	slog.Debug("creating database")
	databaseProperties := azcosmos.DatabaseProperties{ID: databaseName}

	ctx := context.TODO()
	_, err := client.CreateDatabase(ctx, databaseProperties, nil)
	// Parse the error as we expect a 409 when the database already exists
	azRespErr, ok := err.(*azcore.ResponseError)
	if ok {
		if azRespErr.StatusCode == 409 {
			slog.Debug("database already exists")
			err = nil
		}
	}
	// If we get here, we have an error that is not a 409
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %v", err)
	}
	databaseClient, err := client.NewDatabase(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %v", err)
	}
	return databaseClient, nil
}

func CreateContainer(client *azcosmos.Client, databaseName, containerName, partitionKey string) (*azcosmos.ContainerClient, error) {
	slog.Debug("creating container")

	databaseClient, err := client.NewDatabase(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %v", err)
	}

	// Setting container properties
	containerProperties := azcosmos.ContainerProperties{
		ID: containerName,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{partitionKey},
		},
	}

	options := &azcosmos.CreateContainerOptions{}
	ctx := context.TODO()
	// Parse the error as we expect a 409 when the database already exists
	_, err = databaseClient.CreateContainer(ctx, containerProperties, options)
	azRespErr, ok := err.(*azcore.ResponseError)
	if ok {
		if azRespErr.StatusCode == 409 {
			slog.Debug("container already exists")
			err = nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	containerClient, err := databaseClient.NewContainer(containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container client: %v", err)
	}

	return containerClient, nil
}

type ItemID string
type ItemContent string

func GetRandomID() ItemID {
	i := rand.Intn(100000000)
	hashedI := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", i)))
	return ItemID(hashedI[:8])
}

func EncodeContent(content string) ItemContent {
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
	return ItemContent(encodedContent)
}

func GetCurrentTime() time.Time {
	return time.Now().UTC()
}

func ParseTime(t string) (time.Time, error) {
	return time.Parse(time.RFC3339, t)
}

func DecodeContent(content ItemContent) (string, error) {
	decodedContent, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %v", err)
	}
	return string(decodedContent), nil
}

func (h *CosmosHandler) CreateItem(itemID ItemID, item *Item) error {
	slog.Debug("creating item")
	containerClient, err := h.Client.NewContainer(h.DatabaseName, h.ContainerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client: %s", err)
	}

	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(h.PartitionKey)

	b, err := json.Marshal(item)
	if err != nil {
		return err
	}
	// setting item options upon creating ie. consistency level
	itemOptions := azcosmos.ItemOptions{
		ConsistencyLevel: azcosmos.ConsistencyLevelSession.ToPtr(),
	}
	ctx := context.TODO()
	itemResponse, err := containerClient.CreateItem(ctx, pk, b, &itemOptions)

	if err != nil {
		return err
	}
	slog.Info("Item created", "id", item.Id, "activityId", itemResponse.ActivityID, "requestCharge", itemResponse.RequestCharge)

	return nil
}

func (h *CosmosHandler) ReadItem(itemID ItemID) (*Item, error) {
	slog.Debug("creating item")
	containerClient, err := h.Client.NewContainer(h.DatabaseName, h.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a container client: %s", err)
	}

	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(h.PartitionKey)

	ctx := context.TODO()
	itemResponse, err := containerClient.ReadItem(ctx, pk, string(itemID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read item: %v", err)
	}

	var item Item
	err = json.Unmarshal(itemResponse.Value, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %v", err)
	}

	return &item, nil
}

func (h *CosmosHandler) DeleteItem(itemID ItemID) error {
	slog.Debug("deleting item")
	containerClient, err := h.Client.NewContainer(h.DatabaseName, h.ContainerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client: %s", err)
	}

	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(h.PartitionKey)

	ctx := context.TODO()
	itemResponse, err := containerClient.DeleteItem(ctx, pk, string(itemID), nil)
	if err != nil {
		return fmt.Errorf("failed to delete item: %v", err)
	}
	slog.Info("Item deleted", "id", itemID, "activityId", itemResponse.ActivityID, "requestCharge", itemResponse.RequestCharge)

	return nil
}

func (h *CosmosHandler) GetAllItems() ([]Item, error) {
	slog.Debug("getting all items")
	pk := azcosmos.NewPartitionKeyString(h.PartitionKey)
	queryPager := h.ContainerClient.NewQueryItemsPager("SELECT * FROM docs c", pk, nil)
	allItems := []Item{}
	for queryPager.More() {
		slog.Debug("found items")
		queryResponse, err := queryPager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get next page: %v", err)
		}
		fmt.Printf("%+v\n", queryResponse)
		for _, respItem := range queryResponse.Items {
			var item Item
			err := json.Unmarshal(respItem, &item)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal item: %v", err)
			}
			slog.Debug("got item", "id", item.Id)
			allItems = append(allItems, item)
		}
	}
	return allItems, nil
}
