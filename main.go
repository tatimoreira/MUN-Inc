package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"fmt"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//Account is the representation of the graphql data model for the account
type Customer struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Lastname string `json:"lastName"`
	email    string `json:"email"`
}

func main() {
	fmt.Println("Starting application...")

	//cluster DB connection
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://dbUser:32YkYR7rCnwV3U8r@cluster0-2hzfn.mongodb.net/test"))
	if err != nil {
		log.Fatal(err)
	}

	//context cancells request-scoped
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//connect to Mongo
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	//ping the connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	//list the db
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)

	munDatabase := client.Database("mun")
	customerCollection := munDatabase.Collection("customer")
	fmt.Println("connected to mongodb")

	customerType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Customer",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"lastName": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
	/*
		blogType := graphql.NewObject(graphql.ObjectConfig{
			Name: "Blog",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.String,
				},
				"account": &graphql.Field{
					Type: graphql.String,
				},
				"title": &graphql.Field{
					Type: graphql.String,
				},
				"content": &graphql.Field{
					Type: graphql.String,
				},
			},
		})*/
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"customers": &graphql.Field{
				Type: graphql.NewList(customerType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					findOptions := options.Find()
					//Define an array in which you can store the decoded documents
					var results []Customer
					cursor, err := customerCollection.Find(context.TODO(), bson.D{{}}, findOptions)
					if err != nil {
						fmt.Println(err) // prints 'document is nil'
					}

					for cursor.Next(context.TODO()) {
						//Create a value into which the single document can be decoded
						var elem Customer
						err := cursor.Decode(&elem)
						if err != nil {
							log.Fatal(err)
						}

						results = append(results, elem)

					}

					if err := cursor.Err(); err != nil {
						log.Fatal(err)
					}

					//Close the cursor once finished
					cursor.Close(context.TODO())

					fmt.Printf("Found multiple documents: %+v\n", results)

					return results, nil

				},
			},
		},
	})
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootMutation",
		Fields: graphql.Fields{},
	})
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: r.URL.Query().Get("query"),
		})
		json.NewEncoder(w).Encode(result)
	})
	http.ListenAndServe(":8080", nil)
}
