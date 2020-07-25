To migrate go to migrate.go and make changes


TODO/KEEP THIS IN MIND

1) Use the relationship between two items in a single row.... 
    
   | item_id | relationship jsonb |
        1      { entity_id1: [item_id1,item_id2 ....], entity_id2: [item_id1,item_id2 ....]}

2) Use mysql for filtering.

3) Use CH for events and analytics.

