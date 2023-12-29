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

func CreateContainer(client *azcosmos.Client, databaseName, containerName, partitionKey string) error {
	slog.Debug("creating container")

	databaseClient, err := client.NewDatabase(databaseName)
	if err != nil {
		return fmt.Errorf("failed to create database client: %v", err)
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
		return fmt.Errorf("failed to create container: %v", err)
	}

	return nil
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

func CreateItem(client *azcosmos.Client, databaseName string, containerName string, itemID ItemID, item *Item) error {
	slog.Debug("creating item")
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client: %s", err)
	}

	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString("/id")

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

func ReadItem(client *azcosmos.Client, databaseName string, containerName string, itemID ItemID) (*Item, error) {
	slog.Debug("creating item")
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a container client: %s", err)
	}

	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString("/id")

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

func DeleteItem(client *azcosmos.Client, databaseName string, containerName string, itemID ItemID) error {
	slog.Debug("deleting item")
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client: %s", err)
	}

	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString("/id")

	ctx := context.TODO()
	itemResponse, err := containerClient.DeleteItem(ctx, pk, string(itemID), nil)
	if err != nil {
		return fmt.Errorf("failed to delete item: %v", err)
	}
	slog.Info("Item deleted", "id", itemID, "activityId", itemResponse.ActivityID, "requestCharge", itemResponse.RequestCharge)

	return nil
}

// func GetAllItems() {
// 	pk := azcosmos.NewPartitionKeyString("myPartitionKeyValue")
// 	queryPager := container.NewQueryItemsPager("select * from docs c", pk, nil)
// 	for queryPager.More() {
// 		queryResponse, err := queryPager.NextPage(context)
// 		if err != nil {
// 			handle(err)
// 		}

// 		for _, item := range queryResponse.Items {
// 			var itemResponseBody map[string]interface{}
// 			json.Unmarshal(item, &itemResponseBody)
// 		}
// 	}
// }
