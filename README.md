# Relay
Project Relay is the sales/customer-success software built on top of the no-code framework. Which means, the end user can build `n` number of entities/modules on top of the base software based on his business needs. But the UI still needs to customized for each software.

## Getting Started
> make run
> make crm

## Technical Stack
- Core Stack - Golang
- Core DB - Psql
- Segmentation/Workflow - Psql-Pivot-Tables/RedisGraph
- Analytics - Not decided

## Features

### Entities

### Items

### Rule Engine

### Work Flow

### Pipeline

### Relationships
- Add a new table to maintain the "REVERSE" reference of the reference field. Query that relationship table to find the displayable child entities. The "STRAIGHT" reference of the parent entity is the "REVERSE" reference of the child entity.

REVERSE - BelongsTo
STRAIGHT - HasMany

1. one-to-many (TypeBond - STRAIGHT)
If a contact has many assignees. Then the contact entity would have the assigneeID as the reference field with (type = bond). So, the relationship will be like (src - contact) (dest - assignee). In the display panel of the contacts we will show the assignees as the field property. In the users panel we will show the associated contacts based on "type=bond".
2. one-to-many (TypeBond - REVERSE)
If a contact has many tasks. Then the task entity would have the contactID as the reference field with (type = bond). So, the relationship will be like (src - task) (dest - contact). 
3. many-to-many (TypeAssociation - STRAIGHT/REVERSE) (src - deal) (dest - contact) && (src - contact) (dest - deal)
If a contact has many deals and a deal has many contacts. The relationship will be like (src - deal) (dest - contact)with (type = TypeImplicitBond). In the display panel of the contacts we will show the deals as the child associations and allow to create the deal with contact-id prefilled and vice-versa.
4. special-case (TypeAssociation - REVERSE) (src - activity) (dest - contact)
If a contact has many activities. Then the activity entity would have the contactID as the reference field with (type = association). Though the events is not an regular entity the relationships still holds true for them.

### Facts Of Fields
1. Reference - refernce field holds list of map of entityID,itemID.
2. On Segmentation - we have to handle to use cases. 
    a. STRAIGHT - A has B (on segmenting A)
    b. REVERSE  - B has A (on segmenting A)
3. On REVERSE, make the field key of the conditional field as empty
4. On STRAIGHT, make the field key of the conditinal field as not empty
5. Refer below section to understand more about the REVERSE use cases.

### Usecase 1 - REVERSE (making key="" for the reverse)
1. Filter the contacts which have deal.amount>1000. Where the deal has contacts and not the other way around.
    a. Create a conditional field with empty key
    b. Assign the entityID of the deal to the ref field value & leave the value as blank
    c. Assign the where condition properties in the sub-field, just like other fields
    d. If you do like that, the SegmentBaseGNode in pivot.go will create the node with the reverse property enabled.
    e. Enabling reverse property by setting the field key as empty is the one and only active change which we are making here. The rest of the code is just works like the STRAIGHT case and produce the result. 
#### example: item_test.go,segment_test.go

### The Reference Field
check README.md inside the entity

### Workflow/Playbook/Pipeline
1. Sequence of task nodes is called as workflow
2. Sequence of stages/goals with task nodes in a single direction is called as playbook/pipelines. 
3. The flow types are:- Segment(1) & Pipeline(3)
4. The pipeline flows will always hold the node type called stage.
100. More about this in the package internal/rule READ.ME

### Random Assumptions In This Project Are Captured Here. (Please read & understand before start writing the code)
1. The reference field with DOM type `DomSelect` implicitly means that the field unit is associated with the reference. So, the choices for that field unit must be populated in the fields section in the item retrive call.
(check `updateReferenceFields` in the `item.go` for implementation details)

---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------

### User Shots
1. Users/Members/Roles/Permissions.
2. E-mail Integ/Google Calender Intg.

### Workflow Shots
1. The trigger with references. like, deal > 100 for a contact
2. Events as trigger in the workflow
3. Adding templates

### Integrations
1. Phone 
2. Chat
3. Mail

### Features
1. Segmentation
2. Search
3. Sort
4. Events
5. Playbook like hubspot
6. Notification

### Small Shots
0. ******* Write README in the pipeline/playbooks/workflow ******
1. AND/OR in segmentation/workflow
2. Add aggregation <,> in "IN" of list rule engine
3. Stop cyclic looping of references - pivot.go
4. comments for notes/meetings

### UI Shots
1. workflow UI
2. create/edit item
3. create/edit entity
4. pipeline view
5. rich text for Todo/Notes/Meeting Desc

---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------

### Holy Shots
1. make the UI on par with hubspot!

### Done Shots
1. Delete/Update references - pivot.go 
2. Upsert should include only the modified values - pivot.go
3. Deal Stage
4. Workflow with datatype triggers. Like, on numbers (is gt than, is ls than) on date (before, after)
5. Todo
6. Notes, Tickets -  With Association

### Half Done Shots
1. Add aggregation funcs in the rGraph segmentation.
2. Reminder - notification

---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------
