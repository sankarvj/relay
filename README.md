# Relay
Project Relay is the sales/customer-success software built on top of the no-code framework. Which means, the end user can build `n` number of entities/modules on top of the base software based on his business needs of his own. 

## Getting Started
> make run

## Technical Stack
- Core Stack - Golang
- Core DB - Psql
- Segmentation/Workflow - Psql-Pivot-Tables/RedisGraph
- Analytics - Not decided

## Features

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

### Usecase 1 - REVERSE 
1. Filter the contacts which have deal.amount>1000. Where the deal has contacts and not the other way around.
    a. Create a conditional field with empty key
    b. Assign the entityID of the deal to the conditional field value & leave the itemID as blank
    c. Assign the where condition properties in the sub-field, just like other fields
    d. If you do like that, the SegmentBaseGNode in pivot.go will create the node with the reverse property enabled.
    e. Enabling reverse property by setting the field key as empty is the one and only active change which we are making here. The rest of the code is just works like the STRAIGHT case and produce the result. 
#### example: item_test.go

### Workflow/Playbook/Pipeline
1. Sequence of task nodes is called as workflow
2. Sequence of stages/goals with task nodes in a single direction is called as playbook/pipelines. 
3. The flow types are:- Segment(1) & Pipeline(3)
4. The pipeline flows will always hold the node type called stage.
100. More about this in the package internal/rule READ.ME


### TODO
1. Stop cyclic looping of references - pivot.go
2. Delete/Update references - pivot.go
3. Upsert should include only the modified values - pivot.go
4. Add aggregation funcs in the rGraph segmentation.
5. Add aggregation <,> in "IN" of list rule engine
6. MathAny/MatchAll in segmentation/workflow
7. ******* Add upsertEdge inside the upsert itself ********
8. ******* Write READ.ME in the pipeline/playbooks/workflow ******



