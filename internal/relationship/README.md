# Relationships

There are two kinds of relationships.
    - Implicit - relationships created by lookup field (ex - a lookupfield (to) in emails references contact)
    When an item gets added/updated the related connections will be updated automatically.
    - Explicit - relationships can be created explicitly (ex - user can create a relationship between contact & company). A new item should be created manually.


NOTE: Right now, we are going with the psql way for finding the sub-items using the connections table.
But in future, please use redis graph to directly fetch the child items

Because, in this model
1. we need to sync the field update on relationship update
2. connections update on each field update
Simply, we are duplicating the values in multiple places which might not scale.

Beyond simply defining how records can be related to other records, 1:N entity relationships also provide data to address the following questions:

1. When I delete a record should any records related to that record also be deleted?
2. When I assign a record, do I also need to assign all records related to that record to the new owner? (for example, assigning a company to deal should also assign that company's contacts to the deal)
3. How can I streamline the data entry process when I create a new related record in the context of an existing record? (It should happen from the UI)
4. How should people viewing a record be able to view the associated records? (Child Entities)

Entities can also participate in a N:N (many-to-many) relationship where any number of records for two entities can be associated with each other.

~ Explicit 
< Implicit by the children's lookup field
> Implicit by the parent's lookup field

Let's see the relationship of contacts
1. Tasks (Implicit) <
2. Companies (Explicit) ~
3. Deals (Implicit) <
4. Notes (Implicit) <
5. Emails (Implicit) <
6. Meetings (Implicit) <
7. Tickets (Implicit) <

Let's see the relationship of deals
1. Tasks (Implicit) <
2. Companies (Implicit) >
3. Contacts (Implicit) >
4. Notes (Implicit) <
5. Emails (Explicit) ~
6. Meetings (Explicit) ~
7. Tickets (Explicit) ~

Let's see the relationship of companies
1. Tasks (Implicit) <
2. Deals (Implicit) <
3. Contacts (Explicit) ~
4. Notes (Implicit) <
5. Emails (Implicit) <
6. Meetings (Implicit) <
7. Tickets (Implicit) <

Let's see the relationship of tasks
1. Companies (Implicit) >
2. Deals (Implicit) >
3. Contacts (Implicit) >
4. Notes (Explicit) ~
5. Emails (Explicit) ~
6. Meetings (Explicit) ~
7. Tickets (Explicit) ~
