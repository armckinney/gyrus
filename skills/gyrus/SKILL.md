---
name: gyrus
description: Gyrus Unified Context & Memory Engine agent skill wrapper. Use to search, create, update, link, and suggest context from local OKF documents.
---

# Gyrus Agent Skill

This skill allows AI agents to interact with Gyrus memory and context storage via the compiled `gyrus` CLI binary.

## Available Actions

### 1. Suggest Context for a Prompt
```bash
gyrus suggest-context --prompt "<task or question description>"
```

### 2. Search Memory Documents
```bash
gyrus search --query "<keyword>" --category "<category>" --type "<type>"
```

### 3. Fetch Document Payload
```bash
gyrus get <document-id>
```

### 4. Create Contract Document
```bash
gyrus create --id "<id>" --title "<title>" --category "<category>" --type "<type>" --owner-group "<owner>" --content "<markdown body>"
```

### 5. Update Contract Document
```bash
gyrus update <id> --status "<new-status>" --content "<updated markdown body>"
```

### 6. Link Document Dependency
```bash
gyrus link <from-id> <to-id> --rel-type "depends_on"
```

### 7. Re-index Storage Directory
```bash
gyrus sync
```
