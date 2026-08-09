# OpsFlow — Data Schema

## 1. Database Strategy

PostgreSQL is the transactional source of truth.

Each service owns its tables logically. For a single-developer MVP, a shared PostgreSQL instance is acceptable, but service boundaries must remain explicit.

Do not allow arbitrary cross-service table writes.

## 2. Core Entities

```text
users
roles
permissions
user_roles

services
service_environments
service_dependencies
health_checks
health_check_results

incidents
incident_events
incident_comments
incident_evidence
incident_rcas

deployments
deployment_validations

notifications

outbox_events

ai_conversations
ai_messages
ai_tool_calls
ai_usage
ai_approvals

knowledge_documents
knowledge_chunks
```

## 3. Users

```text
users
-----
id UUID PK
email VARCHAR UNIQUE
password_hash TEXT
display_name VARCHAR
status VARCHAR
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 4. Roles

```text
roles
-----
id UUID PK
name VARCHAR UNIQUE
description TEXT
```

## 5. Permissions

```text
permissions
-----------
id UUID PK
code VARCHAR UNIQUE
description TEXT
```

## 6. User Roles

```text
user_roles
----------
user_id UUID FK users.id
role_id UUID FK roles.id

PK(user_id, role_id)
```

## 6a. Role Permissions

```text
role_permissions
----------------
role_id UUID FK roles.id
permission_id UUID FK permissions.id

PK(role_id, permission_id)
```

## 7. Services

```text
services
--------
id UUID PK
name VARCHAR UNIQUE
description TEXT
owner_team VARCHAR
criticality VARCHAR
repository_url TEXT
status VARCHAR
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 8. Service Environments

```text
service_environments
--------------------
id UUID PK
service_id UUID FK services.id
environment VARCHAR
base_url TEXT
health_endpoint TEXT
created_at TIMESTAMP
updated_at TIMESTAMP

UNIQUE(service_id, environment)
```

## 9. Service Dependencies

```text
service_dependencies
---------------------
id UUID PK
service_id UUID FK services.id
depends_on_service_id UUID FK services.id
dependency_type VARCHAR
critical BOOLEAN
created_at TIMESTAMP
```

## 10. Health Checks

```text
health_checks
-------------
id UUID PK
service_environment_id UUID FK service_environments.id
name VARCHAR
method VARCHAR
path TEXT
expected_status INTEGER
timeout_ms INTEGER
interval_seconds INTEGER
enabled BOOLEAN
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 11. Health Check Results

```text
health_check_results
--------------------
id UUID PK
health_check_id UUID FK health_checks.id
status VARCHAR
status_code INTEGER NULL
latency_ms INTEGER NULL
error_code VARCHAR NULL
checked_at TIMESTAMP
```

Use retention policies for high-volume results.

## 12. Incidents

```text
incidents
---------
id UUID PK
incident_key VARCHAR UNIQUE
service_id UUID FK services.id
title VARCHAR
description TEXT
severity VARCHAR
status VARCHAR
assignee_id UUID FK users.id NULL
started_at TIMESTAMP
resolved_at TIMESTAMP NULL
created_by UUID FK users.id
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 13. Incident Events

```text
incident_events
---------------
id UUID PK
incident_id UUID FK incidents.id
event_type VARCHAR
message TEXT
actor_id UUID FK users.id NULL
metadata JSONB
created_at TIMESTAMP
```

## 14. Incident Comments

```text
incident_comments
-----------------
id UUID PK
incident_id UUID FK incidents.id
author_id UUID FK users.id
content TEXT
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 15. Incident Evidence

```text
incident_evidence
-----------------
id UUID PK
incident_id UUID FK incidents.id
object_key TEXT
file_name VARCHAR
content_type VARCHAR
size_bytes BIGINT
checksum VARCHAR
uploaded_by UUID FK users.id
created_at TIMESTAMP
```

Actual binary data belongs in MinIO, not PostgreSQL.

## 16. Incident RCA

```text
incident_rcas
-------------
id UUID PK
incident_id UUID FK incidents.id UNIQUE
summary TEXT
impact TEXT
root_cause TEXT
mitigation TEXT
resolution TEXT
preventive_actions TEXT
status VARCHAR
created_by UUID FK users.id
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 17. Deployments

```text
deployments
-----------
id UUID PK
service_id UUID FK services.id
environment VARCHAR
version VARCHAR
commit_sha VARCHAR
image_reference TEXT
status VARCHAR
deployed_by UUID FK users.id NULL
started_at TIMESTAMP
completed_at TIMESTAMP NULL
created_at TIMESTAMP
```

## 18. Deployment Validations

```text
deployment_validations
----------------------
id UUID PK
deployment_id UUID FK deployments.id
check_name VARCHAR
status VARCHAR
latency_ms INTEGER NULL
message TEXT
executed_at TIMESTAMP
```

## 19. Notifications

```text
notifications
-------------
id UUID PK
user_id UUID FK users.id
channel VARCHAR
event_type VARCHAR
payload JSONB
status VARCHAR
attempts INTEGER
sent_at TIMESTAMP NULL
created_at TIMESTAMP
```

## 20. Outbox

```text
outbox_events
-------------
id UUID PK
aggregate_type VARCHAR
aggregate_id UUID
event_type VARCHAR
payload JSONB
status VARCHAR
attempts INTEGER
published_at TIMESTAMP NULL
created_at TIMESTAMP
```

Index unpublished events by status and created_at.

## 21. AI Conversations

```text
ai_conversations
----------------
id UUID PK
user_id UUID FK users.id
title VARCHAR
model_provider VARCHAR
model_name VARCHAR
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 22. AI Messages

```text
ai_messages
-----------
id UUID PK
conversation_id UUID FK ai_conversations.id
role VARCHAR
content TEXT
tool_call_id VARCHAR NULL
created_at TIMESTAMP
```

## 23. AI Tool Calls

```text
ai_tool_calls
-------------
id UUID PK
conversation_id UUID FK ai_conversations.id
message_id UUID FK ai_messages.id
tool_name VARCHAR
arguments JSONB
result JSONB NULL
status VARCHAR
approved_by UUID FK users.id NULL
executed_at TIMESTAMP NULL
created_at TIMESTAMP
```

## 24. AI Usage

```text
ai_usage
--------
id UUID PK
conversation_id UUID FK ai_conversations.id
provider VARCHAR
model VARCHAR
input_tokens INTEGER
output_tokens INTEGER
latency_ms INTEGER
estimated_cost NUMERIC NULL
created_at TIMESTAMP
```

## 25. AI Approvals

```text
ai_approvals
------------
id UUID PK
tool_call_id UUID FK ai_tool_calls.id
requested_by UUID FK users.id
approved_by UUID FK users.id NULL
status VARCHAR
expires_at TIMESTAMP
created_at TIMESTAMP
approved_at TIMESTAMP NULL
```

## 26. Knowledge Documents

```text
knowledge_documents
-------------------
id UUID PK
title VARCHAR
document_type VARCHAR
source_uri TEXT
content_hash VARCHAR
status VARCHAR
created_at TIMESTAMP
updated_at TIMESTAMP
```

## 27. Knowledge Chunks

```text
knowledge_chunks
----------------
id UUID PK
document_id UUID FK knowledge_documents.id
chunk_index INTEGER
content TEXT
metadata JSONB
embedding_reference TEXT NULL
created_at TIMESTAMP
```

The exact vector storage implementation may use Elasticsearch or another vector-capable store. Keep this behind a repository abstraction.

## 28. Important Indexes

Create indexes for:

- users.email
- incidents.incident_key
- incidents.service_id
- incidents.status
- incidents.severity
- incident_events.incident_id + created_at
- deployments.service_id + created_at
- health_check_results.health_check_id + checked_at
- outbox_events.status + created_at
- ai_messages.conversation_id + created_at

## 29. Data Rules

1. UUIDs are preferred for externally visible IDs.
2. Timestamps are stored in UTC.
3. Database constraints must enforce basic integrity.
4. Domain validation belongs in application code.
5. JSONB is for flexible metadata, not core relational fields.
6. Never store secrets in this schema.
7. Soft delete only where audit/history requires it.
