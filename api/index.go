package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const mongoDefaultDB = "mapdata"

type Node struct {
	Profiles []map[string]interface{} `json:"data,omitempty"`
	Links    map[string]interface{}   `json:"links,omitempty"`
	Errors   map[string]interface{}   `json:"errors,omitempty"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// connect to MongoDB
	mongoClient, err := connectMongo()
	if err != nil {
		log.Fatal(err)
		return
	}

	// get current last_updated from mongo
	lastUpdated, err := getLastUpdated(mongoClient)

	// get data from MurmurationsServices
	now := time.Now().Unix()
	nodes, err := getNodes(lastUpdated, "")
	if err != nil {
		log.Fatal(err)
		return
	}

	// save profiles
	for _, profile := range nodes.Profiles {
		isDuplicated, err := saveOneProfile(mongoClient, profile)
		if isDuplicated {
			continue
		}
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	nextPage := nodes.Links["next"].(string)
	// todo: if there is the next page, continue to process
	// todo: bug: due to the limitation of getting data, this part needs to revise
	for nodes.Links["next"] != nil {
		fmt.Println(nextPage)
		nodes, err := getNodes(lastUpdated, nextPage)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, profile := range nodes.Profiles {
			isDuplicated, err := saveOneProfile(mongoClient, profile)
			if isDuplicated {
				continue
			}
			if err != nil {
				log.Fatal(err)
				return
			}
		}
		nextPage = nodes.Links["next"].(string)
	}

	// save current Timestamp to Mongo
	err = saveTimestamp(mongoClient, now)
	if err != nil {
		log.Fatal(err)
		return
	}

	// disconnect Mongo
	err = disconnectMongo(mongoClient)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Fprintf(w, "<h1>Hello from Go!</h1>")
}

func connectMongo() (*mongo.Client, error) {
	mongoUrl := os.Getenv("MONGO_URL")
	credential := options.Credential{
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	}
	clientOptions := options.Client().ApplyURI(mongoUrl).SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

func disconnectMongo(client *mongo.Client) error {
	err := client.Disconnect(context.TODO())
	if err != nil {
		return err
	}
	fmt.Println("Disconnected from MongoDB!")
	return nil
}

func getLastUpdated(client *mongo.Client) (int64, error) {
	coll := client.Database(mongoDefaultDB).Collection("settings")
	filter := bson.D{{"name", "current"}}
	var setting map[string]interface{}
	err := coll.FindOne(context.TODO(), filter).Decode(&setting)
	if err != nil {
		return -1, err
	}
	return setting["last_updated"].(int64), nil
}

func getNodes(lastUpdated int64, nodeUrl string) (*Node, error) {
	if nodeUrl == "" {
		nodeUrl = os.Getenv("NODE_URL") + "/nodes"
		if lastUpdated != 0 {
			nodeUrl += "?last_updated=" + strconv.Itoa(int(lastUpdated))
		}
	}
	client := http.Client{
		Timeout: time.Second * 5,
	}

	res, err := client.Get(nodeUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data *Node
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func saveOneProfile(client *mongo.Client, profile map[string]interface{}) (bool, error) {
	// find duplicate: if profile_url is the same, print the message skip the data
	coll := client.Database(mongoDefaultDB).Collection("profiles")
	if profile["profile_url"] != nil {
		filter := bson.D{{"profile_url", profile["profile_url"]}}
		count, err := coll.CountDocuments(context.TODO(), filter)
		if err != nil {
			return false, err
		}
		if count > 0 {
			fmt.Println("Profile Duplicated:")
			fmt.Println(profile)
			return true, nil
		}
	}

	// insert the profile
	_, err := coll.InsertOne(context.TODO(), profile)
	if err != nil {
		return false, err
	}
	return false, nil
}

func saveTimestamp(client *mongo.Client, timestamp int64) error {
	coll := client.Database(mongoDefaultDB).Collection("settings")
	filter := bson.D{{"name", "current"}}
	update := bson.D{{"$set", bson.D{{"last_updated", int32(timestamp)}}}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}
