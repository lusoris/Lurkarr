#!/usr/bin/env bash
# deploy/coder/push-template.sh — Push the Lurkarr Coder workspace template
# to the configured Coder instance.
#
# Prerequisites:
#   - coder CLI installed (https://coder.com/docs/install)
#   - Authenticated: coder login <url>
#
# Usage:
#   ./deploy/coder/push-template.sh [template-name]
#
# Environment:
#   CODER_URL           Coder instance URL (default: https://code.dev.cauda.dev)
#   CODER_TEMPLATE_NAME Template name (default: lurkarr-dev)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_DIR="$SCRIPT_DIR"

CODER_URL="${CODER_URL:-https://coder.ancilla.lol}"
TEMPLATE_NAME="${1:-${CODER_TEMPLATE_NAME:-lurkarr-dev}}"

# Verify coder CLI is available
if ! command -v coder &>/dev/null; then
  echo "Error: coder CLI not found. Install from https://coder.com/docs/install"
  exit 1
fi

# Ensure we're logged in
if ! coder whoami &>/dev/null 2>&1; then
  echo "Not logged in. Logging in to $CODER_URL ..."
  coder login "$CODER_URL"
fi

echo "Pushing template '$TEMPLATE_NAME' from $TEMPLATE_DIR ..."

# Check if template already exists
if coder templates list --output json 2>/dev/null | grep -q "\"$TEMPLATE_NAME\""; then
  echo "Template exists — updating..."
  coder templates push "$TEMPLATE_NAME" \
    --directory "$TEMPLATE_DIR" \
    --yes
else
  echo "Template does not exist — creating..."
  coder templates create "$TEMPLATE_NAME" \
    --directory "$TEMPLATE_DIR" \
    --yes
fi

echo "Done. Template '$TEMPLATE_NAME' is ready at $CODER_URL."
echo ""
echo "Create a workspace:"
echo "  coder create my-lurkarr --template $TEMPLATE_NAME"
