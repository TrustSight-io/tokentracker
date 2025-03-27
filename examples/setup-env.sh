#!/bin/bash
# This script helps set up the local development environment for using TokenTracker

# Check if GitHub username is provided
if [ -z "$1" ]; then
  echo "Error: GitHub username not provided"
  echo "Usage: $0 <github-username> <personal-access-token>"
  exit 1
fi

# Check if personal access token is provided
if [ -z "$2" ]; then
  echo "Error: Personal access token not provided"
  echo "Usage: $0 <github-username> <personal-access-token>"
  exit 1
fi

GITHUB_USERNAME=$1
GITHUB_TOKEN=$2

# Setup .netrc file
echo "Setting up .netrc file..."
cat > ~/.netrc << EOF
machine github.com
login $GITHUB_USERNAME
password $GITHUB_TOKEN
EOF

# Make sure .netrc has the right permissions
chmod 600 ~/.netrc

# Set GOPRIVATE environment variable
echo "Setting up GOPRIVATE environment variable..."
go env -w GOPRIVATE=github.com/TrustSight-io/*

# Configure git to use https instead of ssh for GitHub
echo "Configuring git to use HTTPS with authentication..."
git config --global url."https://$GITHUB_USERNAME:$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"

echo "Setup complete. You should now be able to use TokenTracker in your Go projects."
echo "Add the following to your go.mod file:"
echo "require ("
echo "    github.com/TrustSight-io/tokentracker v1.2.3  # Use the appropriate version"
echo ")"
