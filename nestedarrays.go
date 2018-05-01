package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	items, err := fetchExpensiveItems("Handbags", "Gucci", "Italy", 200)
	if err != nil {
		fmt.Printf("Failed with error: %v", err)
	}

	fmt.Printf("Items matching criteria: %v", items)
}

func fetchExpensiveItems(catalog string, brand string, origin string, minPrice float64) ([]Item, error) {
	session, err := mgo.Dial("mongodb://localhost:27017")
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("subdoctest").C("catalog")
	defer session.Close()

	//create the aggregator pipeline that will fetch just the needed data from MongoDB, and nothing more
	pipe := c.Pipe([]bson.M{
		{"$match": bson.M{
			"catalogName": catalog,
			"brands": bson.M{
				"$elemMatch": bson.M{
					"brandName": brand,
				},
			},
		}},
		{"$addFields": bson.M{
			"brands": bson.M{
				"$filter": bson.M{
					"input": bson.M{
						"$map": bson.M{
							"input": "$brands",
							"as":    "b",
							"in": bson.M{
								"brandName": "$$b.brandName",
								"items": bson.M{
									"$filter": bson.M{
										"input": "$$b.items",
										"as":    "i",
										"cond": bson.M{
											"$and": []interface{}{
												bson.M{"$eq": []interface{}{"$$i.origin", origin}},
												bson.M{"$gte": []interface{}{"$$i.price", minPrice}},
											},
										},
									},
								},
							},
						},
					},
					"as": "b",
					"cond": bson.M{
						"$and": []interface{}{
							bson.M{"$eq": []interface{}{"$$b.brandName", brand}},
							bson.M{"$gt": []interface{}{bson.M{"$size": "$$b.items"}, 0}}},
					},
				},
			},
		},
		}})

	//perform the aggregation
	var resp []bson.M
	err = pipe.All(&resp)
	if err != nil {
		return nil, err
	}

	//traverse the bson Map returned by the aggregation and extract the items
	var itemsFound []Item
	for _, catalogMap := range resp {
		brands := catalogMap["brands"].([]interface{})
		for _, b := range brands {
			brandsMap := b.(bson.M)
			items := brandsMap["items"].([]interface{})
			for _, b := range items {
				itemsMap := b.(bson.M)
				data, _ := json.Marshal(itemsMap)
				var item Item
				if err := json.Unmarshal(data, &item); err != nil {
					return nil, err
				}
				itemsFound = append(itemsFound, item)
			}
		}
	}

	return itemsFound, err
}

type Catalog struct {
	ID          string  `bson:"_id" json:"_id"`
	CatalogName string  `bson:"catalogName" json:"catalogName"`
	Brands      []Brand `bson:"brands" json:"brands"`
}

type Brand struct {
	BrandName string `bson:"brandName" json:"brandName"`
	Items     []Item `bson:"items" json:"items"`
}

type Item struct {
	Name   string  `bson:"name" json:"name"`
	Origin string  `bson:"brandName" json:"origin"`
	Price  float64 `bson:"price" json:"price"`
}
