#!/usr/bin/env bash
#
# claude-wrapper.sh - Wrapper for claude command with standard flags
#
# Usage:
#   ./claude-wrapper.sh "<any prompt string>"


# Enable alias expansion in non-interactive shell
shopt -s expand_aliases

# Source no-guard-bashrc.sh to give node bun bla bla all tools to claude as well as to get the claude alias
[ -f ~/.no-guard-bashrc.sh ] && source ~/.no-guard-bashrc.sh

if [ -z "${1:-}" ]; then
  echo "Usage: claude-wrapper.sh <prompt>"
  echo "  e.g., claude-wrapper.sh \"run the promptgram @promptgrams/ralph.md\""
  exit 1
fi

PROMPT="$1"

PREAMBLE="REMEMBER: You are on your own. No human is coming to help. If you need a tool, install it (no sudo). Put env vars in ~/.no-guard-bashrc.sh. If you encounter issues or ambiguity in the prompts/scripts, YOU MUST UPDATE THEM to improve the system for future runs."

POSTAMBLE="If you have created or modified any files, commit and push using conventional commits. Update .gitignore if needed. This is a unit of work for you."

claude -p "$PREAMBLE $PROMPT $POSTAMBLE" \
  --dangerously-skip-permissions \
  --output-format stream-json \
  --verbose | cclean
