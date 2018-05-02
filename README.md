
# Overview #
This example illustrates how to use an aggregation pipeline in MongoDB to fetch targeted data from deeply nested sub-arrays in collections. 
The goal is to have MongoDB return only the data that is needed, and nothing more. With this technique, developers can avoid fetching large documents from the database,
and then traverse them on the client side only to extract the data of interest, and discard the rest. 

# Testing the Aggregation Query #
**Note**: Make sure you are using MongoDB 3.6 or higher.
1. Create a database called "subdoctest" in your local MongoDB instance. Within this database, create a collection called "catalog"
2. Add the data from data.json into this collection.
3. Execute the following query on this collection in a MongoDB shell. 
```
﻿db.getCollection('catalog').aggregate([
  { "$match": {
    "brands": {
      "$elemMatch": { 
        "items.origin": "Italy",
        "items.price": { "$gte": 500 }
      }
    }
  }},
  { "$project": { "_id":0, "brands":1 } },
  { "$addFields": {
    "brands": {
      "$filter": {
        "input": {
          "$map": {
            "input": "$brands",
            "as": "b",
            "in": {
              "items": {
                "$filter": {
                  "input": "$$b.items",
                  "as": "i",
                  "cond": {
                    "$and": [
                      { "$eq": [ "$$i.origin", "Italy" ] },
                      { "$gte": [ "$$i.price", 500 ] }
                    ]
                  }
                }
              }
            }
          }
        },
        "as": "b",
        "cond": 
            { "$gt": [ { "$size": "$$b.items" }, 0 ] }
      }
    }
  }}
])

```
The output of this query should be as follows:
```
{
    "brands" : [ 
        {
            "items" : [ 
                {
                    "name" : "Olphia",
                    "origin" : "Italy",
                    "price" : 1200.0
                }, 
                {
                    "name" : "Mormont",
                    "origin" : "Italy",
                    "price" : 1300.0
                }
            ]
        }
    ]
}

{
    "brands" : [ 
        {
            "items" : [ 
                {
                    "name" : "Racer",
                    "origin" : "Italy",
                    "price" : 600.0
                }
            ]
        }
    ]
}
```

# Testing the Golang Implementation #
1. Ensure local MongoDB server is running, and has the database, collection and data set up from the previous steps.
2. Execute "go get -d" to pull Go dependencies (such as mgo).
3. Execute "go run nestedarrays.go". The program should produce the following output:
```
Items matching criteria: [{Name:Olphia Origin:Italy Price:1200} {Name:Mormont Origin:Italy Price:1300} {Name:Racer Origin:Italy Price:600}]
```
