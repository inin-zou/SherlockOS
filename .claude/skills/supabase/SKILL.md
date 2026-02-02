---
name: supabase
description: Manage Supabase database - create migrations, push schema changes, generate types, check status. Use when user mentions database, migrations, schema, or supabase.
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
---

# Supabase Database Management for SherlockOS

You are managing the Supabase database for SherlockOS, a detective assistance platform.

## Project Configuration

| Setting | Value |
|---------|-------|
| Project Ref | `hdfaugwofzqqdjuzcsin` |
| Project URL | https://hdfaugwofzqqdjuzcsin.supabase.co |
| Backend Path | `/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend` |
| Migrations | `backend/supabase/migrations/` |
| Password | `Cool-inin1214` |

## Available Actions

### 1. Push Migrations to Remote Database
```bash
cd "/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend" && supabase db push
```

### 2. Create New Migration
```bash
cd "/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend" && supabase migration new <migration_name>
```
Then edit the created file in `supabase/migrations/`.

### 3. List All Migrations
```bash
cd "/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend" && supabase migration list
```

### 4. Check Schema Diff (Local vs Remote)
```bash
cd "/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend" && supabase db diff
```

### 5. Generate TypeScript Types
```bash
cd "/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend" && supabase gen types typescript --project-id hdfaugwofzqqdjuzcsin
```

### 6. Pull Remote Schema to Local
```bash
cd "/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend" && supabase db pull
```

### 7. Reset Local Database (Development)
```bash
cd "/Users/yongkangzou/Desktop/Hackathons/Gemini Hackathon/SherlockOS/backend" && supabase db reset
```

## Database Schema

### Tables
| Table | Description | Key Fields |
|-------|-------------|------------|
| `cases` | Investigation cases | id, title, description, created_at |
| `commits` | Timeline entries (append-only) | id, case_id, type, summary, payload |
| `branches` | Hypothesis branches | id, case_id, name, base_commit_id |
| `scene_snapshots` | Current SceneGraph state | case_id, commit_id, scenegraph |
| `suspect_profiles` | Suspect attributes | case_id, attributes, portrait_asset_key |
| `jobs` | Async processing jobs | id, type, status, progress, input, output |
| `assets` | Storage file references | id, case_id, kind, storage_key |

### Enum Types
- **commit_type**: upload_scan, witness_statement, manual_edit, reconstruction_update, profile_update, reasoning_result, export_report
- **job_type**: reconstruction, imagegen, reasoning, profile, export
- **job_status**: queued, running, done, failed, canceled
- **asset_kind**: scan_image, generated_image, mesh, pointcloud, portrait, report

### Realtime-Enabled Tables
- `commits` - Timeline updates
- `jobs` - Progress tracking

## Migration Workflow

When making schema changes:

1. **Create migration file**:
   ```bash
   supabase migration new descriptive_name
   ```

2. **Edit the SQL file** in `backend/supabase/migrations/`

3. **Push to remote**:
   ```bash
   supabase db push
   ```

4. **Update Go models** in `backend/internal/models/` if needed

5. **Run tests**:
   ```bash
   CGO_ENABLED=0 /usr/local/go/bin/go test ./...
   ```

## Common Migration Patterns

### Add Column
```sql
ALTER TABLE table_name ADD COLUMN column_name data_type;
```

### Add Index
```sql
CREATE INDEX idx_name ON table_name(column_name);
```

### Add Foreign Key
```sql
ALTER TABLE child_table
ADD CONSTRAINT fk_name
FOREIGN KEY (column) REFERENCES parent_table(id);
```

### Enable Realtime
```sql
ALTER PUBLICATION supabase_realtime ADD TABLE table_name;
```

## Troubleshooting

If `supabase db push` fails:
1. Check if linked: `supabase link --project-ref hdfaugwofzqqdjuzcsin`
2. Verify credentials in Supabase Dashboard
3. Check migration syntax with `supabase db diff`
