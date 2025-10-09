#!/bin/bash
# Grant App Engine access to all secrets

set -e

PROJECT_ID="github-copy-code-examples"
PROJECT_NUMBER="1054147886816"
SERVICE_ACCOUNT="${PROJECT_NUMBER}@appspot.gserviceaccount.com"

echo "Granting App Engine service account access to secrets..."
echo "Service Account: ${SERVICE_ACCOUNT}"
echo ""

# Array of secrets to grant access to
SECRETS=(
  "CODE_COPIER_PEM"
  "webhook-secret"
  "mongo-uri"
)

for SECRET in "${SECRETS[@]}"; do
  echo "Granting access to: ${SECRET}"
  gcloud secrets add-iam-policy-binding "${SECRET}" \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/secretmanager.secretAccessor" \
    --project="${PROJECT_ID}" 2>&1 | grep -E "Updated|bindings" || echo "  Already has access"
  echo ""
done

echo "✅ Done! Verifying permissions..."
echo ""

for SECRET in "${SECRETS[@]}"; do
  echo "Permissions for ${SECRET}:"
  gcloud secrets get-iam-policy "${SECRET}" \
    --project="${PROJECT_ID}" \
    --format="table(bindings.members)" 2>&1 | grep -A 5 "serviceAccount:${SERVICE_ACCOUNT}" || echo "  Not found"
  echo ""
done

echo "✅ All secrets are now accessible by App Engine!"

