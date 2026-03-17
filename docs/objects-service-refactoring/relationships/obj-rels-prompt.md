- Now lets look at following task: I need to extend objects-service with functionality to manage relationships between objects.

- Schema should support mechanism to describe relationships between object: whilst classic parent-child relationship can be easily expressed by including parent_id field (making table a tree) - sometimes it is needed to handle more complex cases, which does not fit into simple tree.

- In other words, I'm describing essentially M:N relationship between objects with additional attributes describing type of concrete relation.

- Relationship type should be a separate entity with attributes describing typical relationship properties.

- Regarding planned purpose of relationship system: no need to plan for special cases as, for example, graphs or other rather scientific or very specialized use cases. We need to handle typical business and probably financial / marco econ research domains.

- Pls propose database structure for such extension of objects-service. We need first to design database schema, then, based on schema - create multi-phase plan with todo-style structure for progress tracking, and only after finalizing comprehensive planning and review - start actual implementation.
