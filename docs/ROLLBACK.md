# Rollback Procedure — Blog Engine
# Last updated: 2026-06-08 | DevOps Agent (i18n-backend sprint)

---

## When to Roll Back

Roll back if any of the following occur after a deploy:
- Frontend: blank page, JS bundle 404, or nginx returning 5xx for more than 2 minutes
- API: `/health` endpoint returning non-200 for more than 2 minutes
- Database migrations fail on startup (container won't start — old container stays live)
- Coverage CI gate fails on main (block merge at PR stage — never reaches deploy)

---

## Frontend Rollback (nginx container)

**Time to rollback: ~2 minutes**

### Step 1 — Identify the last good tag

```bash
gh api /orgs/your-org/packages/container/blog-engine%2Fblog-engine-frontend/versions \
  --jq '.[].metadata.container.tags[]' | head -10
```

### Step 2 — Re-deploy the previous tag with Terraform

```bash
cd iac/
terraform apply \
  -var="frontend_image_tag=<PREVIOUS_TAG>" \
  -var="db_url=$DATABASE_URL" \
  -var="image_tag=$CURRENT_API_TAG" \
  -auto-approve
```

### Step 3 — Verify

```bash
curl -f https://blog-engine.example.com | grep -q "Blog Engine"
echo "Frontend rollback verified"
```

---

## API Rollback (Go container)

**Time to rollback: ~2 minutes**

### Step 1 — Identify the last good tag

```bash
gh api /orgs/your-org/packages/container/blog-engine%2Fblog-engine-api/versions \
  --jq '.[].metadata.container.tags[]' | head -10
```

### Step 2 — Re-deploy the previous tag

```bash
cd iac/
terraform apply \
  -var="image_tag=<PREVIOUS_TAG>" \
  -var="frontend_image_tag=$CURRENT_FRONTEND_TAG" \
  -var="db_url=$DATABASE_URL" \
  -auto-approve
```

### Step 3 — Verify

```bash
curl -f https://api.blog-engine.example.com/health
echo "API rollback verified"
```

---

## Database Migration Rollback

Migrations are **additive only** (no destructive ALTER/DROP in any sprint). If a migration fails:

1. Container startup fails before the old container is removed — no data loss.
2. Fix the migration SQL, push a new commit, let CI re-deploy.

If data corruption is suspected, restore from the last automated backup using PostgreSQL point-in-time recovery to before the deploy timestamp.

### i18n Migration (`002_i18n.sql`) — Rollback Notes

`002_i18n.sql` adds four nullable columns and one index — it is fully backward-compatible:
- `blogs.title_en` (TEXT, nullable)
- `blogs.body_en` (TEXT, nullable)
- `blogs.translation_status` (VARCHAR 20, NOT NULL DEFAULT 'none')
- `users.language_preference` (VARCHAR 5, NOT NULL DEFAULT 'vi')

If rollback to the pre-i18n API image is required while keeping the migration applied, the old binary will ignore unknown columns — no issue. If the migration must be reversed (unlikely):

```sql
-- ONLY if a full DB reset is needed — data in these columns will be lost
ALTER TABLE blogs DROP COLUMN IF EXISTS title_en;
ALTER TABLE blogs DROP COLUMN IF EXISTS body_en;
ALTER TABLE blogs DROP COLUMN IF EXISTS translation_status;
DROP INDEX IF EXISTS idx_blogs_translation_status;
ALTER TABLE users DROP COLUMN IF EXISTS language_preference;
```

**Important:** removing `translation_status` from a live DB will break the i18n-capable API binary. Only run the above after confirming the deployed image is pre-i18n.

---

## Emergency: Full Stack Rollback

```bash
cd iac/
terraform apply \
  -var="image_tag=<LAST_GOOD_API_TAG>" \
  -var="frontend_image_tag=<LAST_GOOD_FRONTEND_TAG>" \
  -var="db_url=$DATABASE_URL" \
  -auto-approve
```

---

## Deploy Tag History

| Sprint | API Tag | Frontend Tag | Date |
|--------|---------|--------------|------|
| Sprint 1 (API) | wireup-initial | — | 2026-05-30 |
| Frontend sprint | wireup-initial | frontend-initial | 2026-05-31 |
| i18n-backend delta | i18n-backend-initial | frontend-initial (unchanged) | 2026-06-08 |
| i18n-frontend delta | i18n-backend-initial (unchanged) | i18n-frontend-initial | 2026-06-08 |
