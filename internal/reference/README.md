# Reference 

## What is this choice expression?

The expressions in the choices will be evaluated and if the evalution returns true then that corresponding choiceID will be
set for that field.
example: status field, the status will set to overdue if the due_by field value is lesser than the current time. Here the 
expression would be like: if due_by_time < current_time --> due_by_choice_id

## How to decide whether to populate the choices for the reference or not?

If the Dom type is select then fetch the choices, before that just check that the referenced entity is of type field units.
Because field units by itself will not allow to add more than 100 items per entity.
If the Dom type is autocomplete then don't pre-compute the choices. The choices should be populated on-demand.

## How to show the values for the selected choice in the view. In otherwords, only the DomSelect choices are pre-populated what about the reference data type with DomAutoComplete/DomText/DomNumber?

Those will be populated by the bulk retrive strategy. Only the item selected come. The choices will not populated in the list view atleast.

## How to handle flows/stages in the fields?

To handle flows. Create a new type called TypeFlow & TypeNodes

## What is dependent Field?

Lets say you have a company entity. In that you can have the field called deals. The Company also has the field called contacts. If I set the deals as dependent to contact field. The deals populated in the company view will be filtered by the selected contact ID. But one catch,  for this to work the deal entity should have the contact entity referenced in one of its field.

