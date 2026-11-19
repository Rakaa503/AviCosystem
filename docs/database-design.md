# Database Design

## Database Engine

PostgreSQL is the primary relational database.

GORM is used as the ORM layer.

## Data Ownership

Database models should belong to their respective domain modules where practical.

Current domain boundaries include:

- Authentication
- Users
- AI
- Ask
- Chat
- Community
- Learning
- Notifications
- Scheduling
- Stack

## Database Guidelines

1. Use explicit model definitions.
2. Use UUID identifiers where defined by the domain model.
3. Normalize and validate user input before persistence.
4. Never store plaintext passwords.
5. Use database constraints for data integrity.
6. Handle database errors explicitly.
7. Use migrations carefully and review schema changes.

## Testing

SQLite in-memory is used by selected repository tests.

Production database configuration remains PostgreSQL.

Test database behavior should not be assumed to be identical to PostgreSQL in every case.
