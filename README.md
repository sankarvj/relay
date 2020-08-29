To migrate go to migrate.go and make changes


TODO/KEEP THIS IN MIND

1) Use the relationship between two items in a single row.... 
    
   | item_id | relationship jsonb |
        1      { entity_id1: [item_id1,item_id2 ....], entity_id2: [item_id1,item_id2 ....]}

2) Use redisgraph/mysql pivot for filtering.

3) Use CH for events and analytics.

**Relationships
- Add a new table to maintain the reverse relationship of the reference field. Query that table to find the displayable child entities of the parent entity.
1. one-to-many 
If a contact has many tasks. Then the task entity would have the contactID as the reference field with (type = one). So, the relationship will be like (src - task) (dest - contact). In the display panel of the contacts we will show the tasks as the child entity and allow to create the task with contact-id prefilled.
2. one-to-many 
If a contact has many assignees. Then the contact entity would have the assigneeID as the reference field with (type = one/zero). So, the relationship will be like (src - contact) (dest - assignee). In the display panel of the contacts we will show the assignees as the field property. In the users panel we will show the associated contacts based on "type=one/zero".
3. many-to-many 
If a contact has many deals and a deal has many contacts. The deal will have the multiselect contactID as the reference field with (type = two). So, the relationship will be like (src - deal) (dest - contact). In the display panel of the contacts we will show the deals as the child entity and allow to create the deal with contact-id prefilled. In the similar fashion deals are allowed to associate multiple contacts.

**How the above design helps segmentation
In the redisGraph always relate two reference fields with the bi-directional relationship. (make sure to delete/recreate the relation in the event of update)
1. Filter the contacts which have deal.amount>1000. Match the c, Match the d, c-contains->d with where class d.amount>1000. 






