# recall shell hook

_recall_session_id="${_recall_session_id:-$(date +%s%N)-$$}"

_recall_preexec() {
    _recall_cmd="$1"
    _recall_start=$(date +%s%3N)
}

_rec_precmd() {
    local exit_code=$?
    if [ -n "$_recall_cmd" ]; then
        local end duration
        end=$(date +%s%3N)
        duration=$(( end - _recall_start ))
        recall log \
            --cmd "$_recall_cmd" \
            --exit "$exit_code" \
            --duration "$duration" \
            --shell "zsh" \
            --session "$_recall_session_id" \
            --cwd "$PWD" \
            >/dev/null 2>&1 &!
    fi
    _recall_cmd=""
}

autoload -Uz add-zsh-hook
add-zsh-hook preexec _recall_preexec
add-zsh-hook precmd _recall_precmd
