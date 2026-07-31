# recall shell hook

_recall_session_id="${_recall_session_id:-$(date +%s%N)-$$}"

_recall_preexec() {
    _recall_cmd="$1"
    _recall_start=$(date +%s%3N)
}

# trap DEBUG fires before each command; only capture the first call per
# prompt cycle to avoid re-firing for PROMPT_COMMAND itself.
trap '_recall_preexec "$BASH_COMMAND"' DEBUG

_recall_precmd() {
    local exit_code=$?
    # Skip if no command was actually run (e.g. just pressing enter)
    if [ -n "$_recall_cmd" ]; then
        local end
        end=$(date +%s%3N)
        local duration=$(( end - _recall_start ))
        recall log \
            --cmd "$_recall_cmd" \
            --exit "$exit_code" \
            --duration "$duration" \
            --shell "bash" \
            --session "$_recall_session_id" \
            --cwd "$PWD" \
            >/dev/null 2>&1 &
        disown
    fi
    _recall_cmd=""
}

PROMPT_COMMAND="_recall_precmd${PROMPT_COMMAND:+; $PROMPT_COMMAND}"
