package job

import (
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func sendMail(e entity.Entity, itemID string, vals map[string]interface{}) error {
	fields, err := e.AllFields()
	if err != nil {
		return err
	}

	var emailEntityItem entity.EmailEntity
	err = entity.ParseFixedEntity(entity.ValueAddFields(fields, vals), &emailEntityItem)
	if err != nil {
		return err
	}

	return nil
}
