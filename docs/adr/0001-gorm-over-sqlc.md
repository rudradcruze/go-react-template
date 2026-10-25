# Use GORM in place of sqlc

The first plan used sqlc and pgx for database access, and also an ORM. We use only GORM, with pgx as its driver, and we do not use sqlc. One method of database access gives one set of rules and one type of model. GORM models stay in the `infra` layer, so the domain does not see them. Schema changes use goose SQL migrations, not GORM `AutoMigrate`, so each schema change is clear and can be reversed.
