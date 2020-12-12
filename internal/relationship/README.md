Right now, we are going with the psql way for finding the sub-items using the connections table.
But in future, please use redis graph to directly fetch the child items

Because, in this model
1. we need to sync the field update vs relationship update
2. field data update with connections update
Simply, we are duplicating the values in multiple places which might not scale.