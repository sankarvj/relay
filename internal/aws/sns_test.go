package aws

import (
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestToken(t *testing.T) {
	tst := tests.NewIntegration(t)
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6W10sImV4cCI6MjI4NDY2OTYyOSwiaWF0IjoxNjc5ODY5NjI5LCJzdWIiOiIzY2YxNzI2Ni0zNDczLTQwMDYtOTg0Zi05MzI1MTIyNjc4YjcifQ.G-ATM9VAUt9T_Iv0GZedghhV3m4Jr7riOr-vNtLcee6vZZSWo-dEF__vr9zQI0CG7emAhQ_D6z3Egbqv3kabWfC3_OxkaRBHf5R5EIrpT96QCv9CbWp0cdb2R2s02FIDVyW7oqDUxASb7no6QYMYMSxKA57Wj4OiXZs3cN0SM4PLtLCHvSxUgeyF-aMOtvqQ_YQqVyD9Xasr-5RU47x4D9ZXIww5HQCinaqX7XqwjcEjXUMuBrxk6x6ZQBItaW47CfRpNJw8Ecv2D7fZq_MmvD8r6brqwiLuacuP4VFggbdmXKQQZelz-OJ8um64C6LolXPawqu7b-ovjFqNOZJlQA"
	t.Log("\twhen adding the aws message")
	{

		claims, err := tst.Authenticator.ParseClaims(token)
		if err != nil {
			t.Fatalf("\tShould be able to parse token - %s", err)
		}
		log.Println("claims ", claims)
	}
}

func TestAWSIncidents(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)

	// sdb, teardown2 := tests.NewSecDbUnit(t)
	// defer teardown2()

	incidentEntityID := "69d2af4e-61a5-4239-988c-d5e1bd7bc4e7"

	t.Log("Add incident from aws message")
	{
		t.Log("\twhen adding the aws message")
		{

			fields := map[string]interface{}{
				"block":         incidentEntityID,
				"incident_name": "EC2 lost",
				"tags":          []interface{}{"tag1"},
				"unique":        "1",
			}

			err := SaveAlert(tests.Context(), schema.SeedAccountID, fields, db, nil, "")
			if err != nil {
				t.Fatalf("\tShould be able to create an aws incident entity - %s", err)
			}
		}
	}
}
