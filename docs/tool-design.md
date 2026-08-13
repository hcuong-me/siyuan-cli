## Tool Design and ACI Practices

### Three Stages of Tool Evolution

```yaml
Stage 1: API Wrapping
Problem: Too fine-grained; Agent must coordinate multiple tools
Example:
  - get_file_content()
  - write_file()
  - commit_changes()
  
Stage 2: ACI (Agent-Computer Interface)
Core: Tools map to Agent goals rather than low-level operations
Example:
  - update_api_endpoint()      # High-level goal
  - create_user_story()        # Business intent
  - deploy_to_staging()        # Deployment action
  
Stage 3: Advanced Tool Use
Features:
  - Dynamic tool discovery
  - Code-orchestrated tool calls
  - Example-driven execution
```

### Four Principles of Tool Design

#### ① Granularity: Map to Agent Goals

```python
# WRONG: Low-level API operations
@tool
def get_file_content(path: str) -> str:
    """Reads file content"""
    return read_file(path)

@tool
def write_file(path: str, content: str) -> bool:
    """Writes to file"""
    return save_file(path, content)

# CORRECT: High-level goal
@tool
def update_api_endpoint(
    endpoint: str,
    changes: dict,
    reason: str
) -> UpdateResult:
    """
    Updates the implementation of an API endpoint.
    
    Args:
        endpoint: API path (e.g., /api/users)
        changes: Change details
        reason: Reason for change (automatically written to commit message)
    
    Returns:
        UpdateResult containing:
        - List of modified files
        - Affected test cases
        - Suggested validation steps
    """
    # Internal automated handling: Read → Modify → Format → Lint → Commit
    pass
```

#### ② Returns: Decision-Relevant Information

```python
# WRONG: Returning full raw data
def get_user(user_id: str) -> dict:
    return database.query(f"SELECT * FROM users WHERE id = {user_id}")

# CORRECT: Returning decision-relevant information
@tool
def get_user_context(user_id: str) -> UserContext:
    """
    Gets the user context (containing only information required for Agent decision-making)
    
    Returns:
        UserContext:
            - exists: bool (whether user exists)
            - can_modify: bool (has permissions to modify)
            - related_entities: List[str] (associated entities for cascade actions)
            - suggested_actions: List[str] (recommended next steps)
    """
    pass
```

#### ③ Error Handling: Structured Information

```python
# WRONG: Simple string errors
raise Exception("File not found")

# CORRECT: Structured error info
@dataclass
class ToolError:
    error_code: str          # "FILE_NOT_FOUND"
    message: str             # Human-readable description
    fix_suggestion: str      # Actionable fix suggestion for the Agent
    retry_possible: bool     # Whether retry is possible

@tool
def deploy_service(config: dict) -> DeployResult:
    try:
        # Deployment logic
        pass
    except ValidationError as e:
        return ToolError(
            error_code="INVALID_CONFIG",
            message=f"Configuration validation failed: {e}",
            fix_suggestion="Check the config.yaml schema to ensure all required fields exist",
            retry_possible=True,
        )
```

#### ④ Description: Clear Usage Boundaries

```python
@tool
def create_database_migration(
    changes: dict,
    auto_apply: bool = False
) -> MigrationResult:
    """
    Creates a database migration.
    
    Applicable Scenarios:
    ✅ Adding new tables or columns
    ✅ Modifying column types or constraints
    ✅ Adding indexes
    
    Non-Applicable Scenarios:
    ❌ Dropping tables or columns (requires manual confirmation)
    ❌ Modifying core business data (risk is too high)
    ❌ Cross-service data migrations (requires coordination)
    
    Parameter Example:
        changes = {
            "add_table": {
                "name": "orders",
                "columns": [
                    {"name": "id", "type": "uuid", "primary": true},
                    {"name": "user_id", "type": "uuid", "reference": "users.id"}
                ]
            }
        }
    
    Notes:
    - When auto_apply=True, migration executes immediately.
    - Production environment must have auto_apply=False.
    - Migration files are automatically added to version control.
    """
    pass
```

### betaZodTool Practice Example

```typescript
import { betaZodTool } from '@agent/core';
import { z } from 'zod';

const UpdateYuquePostSchema = z.object({
  post_id: z.string().describe("Yuque post ID"),
  title: z.string().optional().describe("New title"),
  content: z.string().optional().describe("New content (Markdown format)"),
  update_reason: z.string().describe("Reason for update, used to generate release notes")
});

export const updateYuquePost = betaZodTool({
  name: "update_yuque_post",
  description: "Updates a Yuque document; suitable for updating existing document content",
  parameters: UpdateYuquePostSchema,
  execute: async (params, context) => {
    // 1. Parameter validation (completed automatically)
    // 2. Permission check
    if (!context.user.canEdit(params.post_id)) {
      return {
        success: false,
        error_code: "PERMISSION_DENIED",
        fix_suggestion: "Please contact the document owner to obtain edit permissions"
      };
    }
    
    // 3. Execute update
    const result = await yuqueApi.updatePost(params);
    
    // 4. Return decision-relevant info
    return {
      success: true,
      updated_fields: Object.keys(params).filter(k => k !== 'update_reason'),
      version: result.version,
      view_url: result.url,
      suggested_next_steps: [
        "Notify relevant members about the update",
        "Check if document links need to be synchronized"
      ]
    };
  }
});
```

## SiYuan CLI contract notes

The Go tool catalog is the source of truth for JSON field types, examples, and
risk boundaries. Side-effecting operations are confirmation-gated: preview
must resolve concrete remote or local targets, publish deterministic
`irreversible_effects`, and issue a token only when state can be inspected.

Errors remain one response envelope. Selector ambiguity uses
`error.candidates[]` with stable IDs and optional names or paths. Apply
re-resolves every target and returns `CONFIRMATION_STALE` before mutating when
the state changes or disappears.

Document search is deliberately unpaginated. It returns all adapter results,
a count, and stable IDs when the `(notebook, hPath)` pair resolves.
