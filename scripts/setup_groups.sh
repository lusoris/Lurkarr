#!/bin/bash

# Configuration API script to set up instance groups

BASE_URL="http://localhost:9705"
ADMIN_USER="admin"
ADMIN_PASS="admin123"

# Login and get session
echo "Logging in..."
LOGIN_RESPONSE=$(curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" \
  "$BASE_URL/api/auth/login")

# Get CSRF token
echo "Getting CSRF token..."
CSRF=$(curl -s -b /tmp/cookies.txt -I "$BASE_URL/api/instance-groups/sonarr" | grep -i "X-CSRF-Token" | sed 's/.*: //' | tr -d '\r')

echo "CSRF Token: $CSRF"

# Create Sonarr group
echo "Creating Sonarr group..."
curl -s -b /tmp/cookies.txt \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"4K to 1080p"}' \
  "$BASE_URL/api/instance-groups/sonarr" | jq .

# Create Radarr group  
echo "Creating Radarr group..."
curl -s -b /tmp/cookies.txt \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"4K to 1080p"}' \
  "$BASE_URL/api/instance-groups/radarr" | jq .

echo "Groups created!"
