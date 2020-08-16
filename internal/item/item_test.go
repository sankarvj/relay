package item_test

import (
	"testing"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestItemRedis(t *testing.T) {
	residPool, teardown := tests.NewRedisUnit(t)
	defer teardown()
	t.Log(" Given the need to add the pivot for the newly created item")
	{
		t.Log("\twhen adding the item to the graph")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
			accountID := "4d247443-b257-4b06-ba99-493cf9d83ce7"
			entityID := "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
			it := item.NewItem{
				Fields: map[string]interface{}{
					"id":                                   "12345",
					"name":                                 "john",
					"age":                                  32,
					"male":                                 true,
					"4d247443-b257-4b06-ba99-493cf9d83ce7": "cypher",
				},
			}
			entityFields := []entity.Field{
				entity.Field{
					Key:      "id",
					DataType: entity.TypeString,
				},
				entity.Field{
					Key:      "name",
					DataType: entity.TypeString,
				},
				entity.Field{
					Key:      "age",
					DataType: entity.TypeNumber,
				},
				entity.Field{
					Key:      "4d247443-b257-4b06-ba99-493cf9d83ce7",
					DataType: entity.TypeString,
				},
			}
			fields := entity.FillFieldValues(entityFields, it.Fields)

			err := item.AddItemNode(ctx, residPool, accountID, entityID, fields, now)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}
	}

}

func TestMerge(t *testing.T) {
	residPool, teardown := tests.NewRedisUnit(t)
	defer teardown()
	accountID := "4d247443-b257-4b06-ba99-493cf9d83ce7"
	entityID := "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
	colors := []string{"blue", "yellow"}
	it := item.NewItem{
		Fields: map[string]interface{}{
			"id":                                   "12345",
			"name":                                 "john",
			"age":                                  32,
			"male":                                 true,
			"4d247443-b257-4b06-ba99-493cf9d83ce7": colors,
		},
	}
	entityFields := []entity.Field{
		entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
		entity.Field{
			Key:      "name",
			DataType: entity.TypeString,
		},
		entity.Field{
			Key:      "age",
			DataType: entity.TypeNumber,
		},
		entity.Field{
			Key:      "4d247443-b257-4b06-ba99-493cf9d83ce7",
			DataType: entity.TypeString,
			List:     true,
		},
	}
	t.Log(" Given the need create nodes and edges")
	{
		t.Log("\twhen adding the new item to the graph")
		{
			fields := entity.FillFieldValues(entityFields, it.Fields)
			_, err := item.Upsert(residPool, accountID, entityID, "12345", fields)
			if err != nil {
				t.Fatalf("\t%s should create the node(item) to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should create the item node(item) to the graph", tests.Success)
		}

		it := item.NewItem{
			Fields: map[string]interface{}{
				"id":   "12345",
				"name": "vijay",
			},
		}

		t.Log("\twhen updating the existing item to the graph")
		{
			// fields := entity.FillFieldValues(entityFields, it.Fields)
			// _, err := item.Upsert(residPool, accountID, entityID, "12345", fields)
			_, err := item.Update(residPool, accountID, entityID, "12345", it.Fields)
			if err != nil {
				t.Fatalf("\t%s should update the exisiting node(item) with name vijay to the graph - %s", tests.Failed, err)
			}
			t.Logf("\t%s should update the exisiting node(item) with name vijay to the graph", tests.Success)
		}

		t.Log("\twhen fetching the updated item to the graph")
		{
			// fields := entity.FillFieldValues(entityFields, it.Fields)
			// _, err := item.Upsert(residPool, accountID, entityID, "12345", fields)
			_, err := item.Temp(residPool, accountID, entityID, "4d247443-b257-4b06-ba99-493cf9d83ce7", "12345")
			if err != nil {
				t.Fatalf("\t%s fetched node should have the name set to vijay - %s", tests.Failed, err)
			}
			t.Logf("\t%s fetched node should have the name set to vijay", tests.Success)
		}

		t.Log("\twhen adding a relation to the updated item to the graph")
		{
			// fields := entity.FillFieldValues(entityFields, it.Fields)
			// _, err := item.Upsert(residPool, accountID, entityID, "12345", fields)
			_, err := item.UpdateRelation(residPool, accountID, entityID, "12345")
			if err != nil {
				t.Fatalf("\t%s should make a relation - %s", tests.Failed, err)
			}
			t.Logf("\t%s should make a relation", tests.Success)
		}

		t.Log("\twhen fetching the updated item with relation to the graph")
		{
			// fields := entity.FillFieldValues(entityFields, it.Fields)
			// _, err := item.Upsert(residPool, accountID, entityID, "12345", fields)
			_, err := item.Temp2(residPool, accountID, entityID, "4d247443-b257-4b06-ba99-493cf9d83ce7", "12345")
			if err != nil {
				t.Fatalf("\t%s fetched node should have the name set to vijay associated with honda - %s", tests.Failed, err)
			}
			t.Logf("\t%s fetched node should have the name set to vijay associated with honda", tests.Success)
		}
	}

}
