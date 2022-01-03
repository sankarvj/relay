# Relay
Project Relay is the sales/customer-success/service-desk/support-desk software built on top of the no-code framework. Which means, the end user can build `n` number of entities/modules on top of the base software based on his business needs. But the UI still needs to customized for each software for some extend. 

## Prerequisite
> ~Install GO
> ~Install PSQL
> `psql -U postgres` //to enter psql cli
> `\c relaydb` //go to relaydb
> ~Install Redis with graph
> `docker run -p 6379:6379 -it --rm redislabs/redisgraph`

## Getting Started
> make seed
> make crm/csm/ctm
> make run

## Running Tests
go test ./...

## Technical Stack
- Core Stack - Golang
- Core DB - Psql
- Segmentation/Workflow - RedisGraph
- Analytics - Not decided (Analyse RedisGraph or other redis services)

## Features

### Entities/Fields
[entities](internal/entity/README.md)
### Items
[items](internal/item/README.md)
[choices](internal/reference/README.md)
### RuleEngine
[rule-engine](internal/platform/ruleengine/README.md)
### WorkFlow
[workflow](internal/rule/flow/README.md)
### Pipeline/Playbook
[pipeline](internal/rule/flow/README.md)
### Search
[search](internal/item/README.md)
### Segmentation *
[segmentation](internal/platform/graphdb/README.md)
### Activity Feeds *
[feeds](TBD.md)
### Analytics
[analytics](TBD.md)
### Integrations
[integrations](internal/platform/integration/README.md)

## Functions

### Jobs
[jobs](internal/job/README.md)
### References
[references](internal/reference/README.md)
### Authentication/Roles
[auth](internal/platform/auth/README.md)
### Relationships
[relationships](internal/relationship/README.md)
### Bootstrap
[bootstrap](internal/bootstrap/README.md)
### Notes
[notes with @mention](TBD.md)
### Notification
[notification](TBD.md)


# FAQs

### What is the base/source?
When you create an item with in a parent entity then the parent entity item is called as base/source. Even when the workflow creates item it chooses the flow entity as its base entity.
### Who will populate the base/source while creating the item from the parent item?
The UI will populate the value of the ref field for the child item with the help of the makeItemProperty function. For the explicit relations the server will handle the connection
### Who will populate the base/source while creating templates?
The UI will populate the value of the ref field for the child item with the help of the makeItemProperty function. The only difference is the value is dynamic. {{baseEntity.id}}
### How the server handles the explicit connection?
The base/source of the item passed by the UI during create API will be used to connect explict coneections during the event creat job.

### How Dependent field works?
If the field is marked with the dependent parent then the choices for that dependent field will be populated by applying the filterID. The filterID is nothing but the value of the parent it is depent upon. If the `states` field depend on `country` then the states would populate `tamil nadu`, `kerala` if the country selected is `India`. Inorder to fetches the indian states the filterID `india` will be passed when fetching the items from the states. 
[more info](internal/reference/README.md)

### How choices updated in the list/create view?
If the type of the field is reference, the choices will be populated based on the category of the entity referenced. 
### child units:
If the referenced entity is child unit then the items are fetched and passed as the choices
### nodes:
The nodes are handled specifically
### flows:
The flows are handled specifically
[more info](internal/reference/README.md)

### How the segmentation for the reverse usecase works?
Filter the contacts which have deal.amount>1000 where the deal has contacts and not the other way around.
    - Create a conditional field with empty key
    - Assign the entityID of the deal to the ref field value & leave the value as blank
    - Assign the where condition properties in the sub-field, just like other fields
    - Set IsReverse field as `true`
    - If you do like that, the SegmentBaseGNode in pivot.go will create the node with the reverse property enabled.
    - Enabling reverse property by setting the field key as empty and setting the reverse key is the one and only active change which we are making here. The rest of the code is just works like the STRAIGHT case and produce the result. 
    * reference : item_test.go,segment_test.go
[more info](internal/platform/graphdb/README.md)


### Clarity Needed
- Dependent field 
[more info](internal/reference/README.md)
- Workflow
1. The trigger with references. like, deal > 100 for a contact
2. Events as trigger in the workflow
3. Adding templates

### Brainstrom
- How to attach a company to multiple playbooks at a time. (Ans: Create multi projects for each company)
- Playbook properties(status, date on-boarded) for the specific company.
- Workflow - send internal notification when the user activity declined.
- Is the reference handled in the rule engine


### Integrations
1. Phone 
2. Chat
3. Mail
4. Meetings


### Pending work
- ******* Write README in the pipeline/playbooks/workflow ******
- AND/OR in segmentation/workflow
- Add aggregation <,> in "IN" of list rule engine
- Stop cyclic looping of references - pivot.go
- comments for notes/meetings - half completed
- Add aggregation funcs in the rGraph segmentation.
- Task Reminder - notification
- e-mail integration watch needs to be called every day.
- receive email and associate - more info on email readme
- implement BulkCreate/BulkUpdate/BulkDelete in relationship
- all the query must have the accountID (MUST, MUST) 

- email from app to users for signup, updates such as - (@mention/assigned/status change/new ticket)
- email with in entity - user has to integrate and read/write
- email as inbox - not needed now.
- calendar reminder



---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------xxxxxxxxx---------------
