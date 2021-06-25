# Relay
Project Relay is the sales/customer-success software built on top of the no-code framework. Which means, the end user can build `n` number of entities/modules on top of the base software based on his business needs. But the UI still needs to customized for each software. 

## Getting Started
> make seed
> make crm
> make run

## Running Tests
go test ./...

## Technical Stack
- Core Stack - Golang
- Core DB - Psql
- Segmentation/Workflow - Psql-Pivot-Tables/RedisGraph
- Analytics - Not decided

## Features

### Entities

### Items

### RuleEngine

### WorkFlow

### Pipeline/Playbook

### Relationships

### Search

### Segmentation

### Activity Feeds

### Analytics

### Reports


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

### What is the base/source?
When you create an item with in a parent entity then the parent entity item is called as base. Even when the workflow creates item it chooses the flow entity as its base entity.
### Who will populate the base/source while creating the item from the parent item?
The UI will populate the value of the ref field for the child item with the help of the makeItemProperty function. For the explicit relations the server will handle the connection
### Who will populate the base/source while creating templates?
The UI will populate the value of the ref field for the child item with the help of the makeItemProperty function. The only difference is the value is dynamic. {{baseEntity.id}}
### How the server handles the explicit connection?
The base/source of the item passed by the UI during create API will be used to connect explict coneections during the event creat job.


---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------
### Current shots
E-mail Integ
E-Mail entity
Calendar Integ
Meeting entity
Inbox
Comments
Notes with @mention

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
4. Meetings

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
7. fix the email view? How to connect the inbox
8. fix the task view with remainders and all

### Half Done Shots
1. Add aggregation funcs in the rGraph segmentation.
2. Reminder - notification
3. e-mail integration watch needs to be called every day.

### Ad-Hoc Shots
- receive email and associate - more info on email readme
- define the types in the relationship to answer the following question. (How should people viewing a record be able to view the associated records) Can I introduce new status called related in the relationship table.
- implement BulkCreate/BulkUpdate/BulkDelete in relationship
- all the query must have the accountID (MUST, MUST) / teamID(can leave in some places)


---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------



# Activities
# Runtime setter of blue-prints (Date, owners)