package handler

import (
	"bytes"
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
	"time"
)

const mongoDefaultDB = "mapdata"

type Node struct {
	Profiles []map[string]interface{} `json:"data,omitempty"`
	Meta     map[string]interface{}   `json:"meta,omitempty"`
	Errors   map[string]interface{}   `json:"errors,omitempty"`
}

type Req struct {
	SearchAfter interface{} `json:"search_after"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("authorization")
	apiSecretKey := os.Getenv("API_SECRET_KEY")
	if authorization != "Bearer "+apiSecretKey {
		fmt.Fprintf(w, "<h1>Unauthorized Operation!</h1>")
		return
	}

	// connect to MongoDB
	mongoClient, err := connectMongo()
	if err != nil {
		log.Fatal(err)
		return
	}

	// get current sort from mongo
	sort, err := getSort(mongoClient)

	// get data from MurmurationsServices
	nodes, err := getNodes(sort)
	if err != nil {
		log.Fatal(err)
		return
	}

	if nodes.Profiles != nil {
		// save profiles
		for _, profile := range nodes.Profiles {
			err := saveOneProfile(mongoClient, profile)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		// save current sort to Mongo
		err = saveSort(mongoClient, nodes.Meta["sort"])
		if err != nil {
			log.Fatal(err)
			return
		}
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

func getSort(client *mongo.Client) (interface{}, error) {
	coll := client.Database(mongoDefaultDB).Collection("settings")
	filter := bson.D{{"name", "current"}}
	var setting map[string]interface{}
	err := coll.FindOne(context.TODO(), filter).Decode(&setting)
	if err != nil {
		return nil, err
	}
	return setting["sort"], nil
}

func saveSort(client *mongo.Client, sort interface{}) error {
	coll := client.Database(mongoDefaultDB).Collection("settings")
	filter := bson.D{{"name", "current"}}
	update := bson.D{{"$set", bson.D{{"sort", sort}}}}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func getNodes(sort interface{}) (*Node, error) {
	nodeUrl := os.Getenv("NODE_URL") + "/export"
	client := http.Client{
		Timeout: time.Second * 5,
	}

	reqBody := Req{
		SearchAfter: sort,
	}
	// make request to buffer
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(reqBody)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	res, err := client.Post(nodeUrl, "application/json", &b)
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

func saveOneProfile(client *mongo.Client, profile map[string]interface{}) error {
	coll := client.Database(mongoDefaultDB).Collection("profiles")
	filter := bson.D{{"profile_url", profile["profile_url"]}}
	update := bson.D{{"$set", profile}}
	opts := options.Update().SetUpsert(true)
	// insert the profile
	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}
