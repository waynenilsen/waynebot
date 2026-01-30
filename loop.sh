#!/usr/bin/env bash
#
# loop.sh - Run crumbler workflow in a continuous loop
#
# Usage:
#   ./loop.sh          # Run indefinitely
#   ./loop.sh 5        # Run 5 iterations
#
# Each iteration:
#   1. Gets prompt from crumbler
#   2. Executes work based on prompt
#   3. Repeats
#


# Enable alias expansion in non-interactive shell
shopt -s expand_aliases

# Source no-guard-bashrc.sh to give node bun bla bla all tools to claude as well as to get the claude alias
[ -f ~/.no-guard-bashrc.sh ] && source ~/.no-guard-bashrc.sh

MAX_ITERATIONS="${1:-0}"  # 0 = unlimited
ITERATION=0

# Get absolute path to this script's directory
pushd "$(dirname "$0")" >/dev/null
SCRIPT_DIR="$(pwd)"
popd >/dev/null

# Colors
CYAN='\033[0;36m'
DIM='\033[2m'
RESET='\033[0m'

main() {
  echo -e "${CYAN}crumbler loop starting${RESET}"
  [ "$MAX_ITERATIONS" -gt 0 ] && echo -e "${DIM}max iterations: ${MAX_ITERATIONS}${RESET}"
  echo ""

  # Source bashrc to ensure agent has access to environment
  [ -f ~/.bashrc ] && source ~/.bashrc || [ -f "$HOME/.bashrc" ] && source "$HOME/.bashrc"

  while true; do
    ITERATION=$((ITERATION + 1))

    # Check for stop signal
    if [ -f "$SCRIPT_DIR/.stop" ]; then
      rm -f "$SCRIPT_DIR/.stop"
      echo -e "${CYAN}Stop signal received. Exiting gracefully.${RESET}"
      exit 0
    fi

    echo -e "${CYAN}━━━ iteration ${ITERATION} ━━━${RESET}"
    echo ""

    # Get prompt from crumbler and execute
    PROMPT=$("$SCRIPT_DIR/crumbler" prompt 2>&1)
    if [ $? -eq 0 ] && [ -n "$PROMPT" ]; then
      # Check if crumbler is done
      if echo "$PROMPT" | grep -q "^# DONE"; then
        echo -e "${CYAN}All crumbs completed. Exiting.${RESET}"
        break
      fi
      "$SCRIPT_DIR/claude-wrapper.sh" "$PROMPT"
    else
      echo -e "${CYAN}No work available or crumbler error${RESET}"
      break
    fi

    # Check iteration limit
    if [ "$MAX_ITERATIONS" -gt 0 ] && [ "$ITERATION" -ge "$MAX_ITERATIONS" ]; then
      echo ""
      echo -e "${CYAN}completed ${ITERATION} iterations${RESET}"
      break
    fi

    echo ""
  done
}

main
