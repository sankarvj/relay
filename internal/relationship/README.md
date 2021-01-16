NOTE: Right now, we are going with the psql way for finding the sub-items using the connections table.
But in future, please use redis graph to directly fetch the child items

Because, in this model
1. we need to sync the field update on relationship update
2. connections update on each field update
Simply, we are duplicating the values in multiple places which might not scale.

Beyond simply defining how records can be related to other records, 1:N entity relationships also provide data to address the following questions:

1. When I delete a record should any records related to that record also be deleted?
2. When I assign a record, do I also need to assign all records related to that record to the new owner?
3. How can I streamline the data entry process when I create a new related record in the context of an existing record?
4. How should people viewing a record be able to view the associated records?

Entities can also participate in a N:N (many-to-many) relationship where any number of records for two entities can be associated with each other.