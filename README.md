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
1. one-to-many 
If a contact has many tasks. Then the task entity would have the contactID as the reference field with (type = one). So, the relationship will be like (src - task) (dest - contact). In the display panel of the contacts we will show the tasks as the child entity and allow to create the task with contact-id prefilled.
2. one-to-many 
If a contact has many assignees. Then the contact entity would have the assigneeID as the reference field with (type = one/zero). So, the relationship will be like (src - contact) (dest - assignee). In the display panel of the contacts we will show the assignees as the field property. In the users panel we will show the associated contacts based on "type=one/zero".
3. many-to-many 
If a contact has many deals and a deal has many contacts. The deal will have the multiselect contactID as the reference field with (type = two). So, the relationship will be like (src - deal) (dest - contact). In the display panel of the contacts we will show the deals as the child entity and allow to create the deal with contact-id prefilled (REVERSE). In the similar fashion deals are allowed to associate multiple contacts (STRAIGHT).
4. special-case
Though the events is not an regular entity the relationships still holds true for them.


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

###


### Big Shots
0. The trigger with references. like, deal > 100 for a contact
1. E-mail Integ
2. Phone Integ
3. Chat
4. Search
5. Segmentation
6. Events
7. Events as trigger in the workflow
8. Playbook like hubspot
9. Use graphDB to fetch records with sorting order

### Small Shots
0. ******* Write READ.ME in the pipeline/playbooks/workflow ******
1. AND/OR in segmentation/workflow
2. Add aggregation <,> in "IN" of list rule engine
3. Stop cyclic looping of references - pivot.go

### Half Done Shots
1. Add aggregation funcs in the rGraph segmentation.

### Current Shots
2. Meetings - Google Calender Intg.
3. Todo - With Due Date
4. Activity Feed
5. Notes, Tickets -  With Association
6. The entity templates during the Workflow 
7. Workflow with datatype triggers. Like, on numbers (is gt than, is ls than) on date (before, after)
8. users/members

### UI Shots
1. Workflow UI
2. create/edit item
3. create/edit entity
4. pipeline view

### Holy Shots
1. make the UI on par with hubspot!

### Done Shots
1. Delete/Update references - pivot.go 
2. Upsert should include only the modified values - pivot.go
3. Deal Stage


