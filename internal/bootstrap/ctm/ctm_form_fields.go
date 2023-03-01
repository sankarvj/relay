package ctm

import (
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func TicketVals(ticketEntity entity.Entity, name, contactID, statusID string) map[string]interface{} {
	ticketVals := map[string]interface{}{
		"name":                 name,
		"associated_contacts":  []interface{}{contactID},
		"associated_companies": []interface{}{},
		"status":               []interface{}{statusID},
	}
	return forms.KeyMap(ticketEntity.NameKeyMapWrapper(), ticketVals)
}
