### The Reference Field

1. The key inside the sub-field is useful for display purpose only. Whatever the key is the actual value is always the list of ids.  
2. The reference field with DOM type `DomSelect` implicitly means that the field unit is associated with the reference. So, the choices for that field unit must be populated in the fields section in the item retrive call.
(check `updateReferenceFields` in the `item.go` for implementation details)
3. For reminder field type, it must/automatically conjugated with the assigned_to field.
4. Similarly the due date field is conjugated with the status field.



### The Category Fields

#### The Field Units
1. The field units cannot take more than 20 items in it.
2. Each form can have 4 field units at the max.

#### The Task
1. Has Reminder 
2. Has Assinged To
3. Has Due Date
4. Has Status
