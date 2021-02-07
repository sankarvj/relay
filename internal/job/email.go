package job

import (
	"log"

	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func sendMail(e entity.Entity, itemID string, vals map[string]interface{}) error {
	fields, err := e.Fields()
	if err != nil {
		return err
	}

	var emailEntityItem entity.EmailEntity
	err = entity.ParseFixedEntity(entity.ValueAddFields(fields, vals), &emailEntityItem)
	if err != nil {
		return err
	}

	log.Printf("TODO: SEND MAIL PART IS PENDING %+v", emailEntityItem)

	return nil
}
