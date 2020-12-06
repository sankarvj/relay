## How delete/update works?
There are three types of deletions, 
1) node delete
2) properties delete
3) relationship delete

### Properties Delete:
To delete a property the user has to set the deleted field with value as NULL

### Relationship Delete:
To delete a relationship the user has to set the unlink offset in the field. (e.g., If the unlink offset is 1, then all the element relationships inside the list/reference will be deleted. If the unlink offset is 2, then all the elements relationship except first element will be deleted. No delete operation happens if the offset is 0)

example:
GRAPH.QUERY Skeleton "MERGE (a:actor{name:'siva'}) SET a.age=42"
GRAPH.QUERY Skeleton "MATCH (a:actor) WITH a,SUM(a.age) as sumofage WHERE sumofage < 100 RETURN a"