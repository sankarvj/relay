### The Reference Field

1. The key inside the sub-field is useful for display purpose only. Whatever the key is the actual value is always the list of ids.  
2. The reference field with DOM type `DomSelect` implicitly means that the field unit is associated with the reference. So, the choices for that field unit must be populated in the fields section in the item retrive call.
(check `updateReferenceFields` in the `item.go` for implementation details)


### The Category Field
1. The field units cannot take more than 20 items in it.
2. Each form can have 4 field units at the max.