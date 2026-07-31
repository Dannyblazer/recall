# recall shell hook

set -g _recall_session_id (date +%s%N)-(status current-pid)

function _recall_preexec --on-event fish_preexec
    set -g _recall_cmd $argv[1]
    set -g _recall_start (date +%s%3N)
end

function _recall_postexec --on-event fish_postexec
    set -l exit_code $status
    if test -n "$_recall_cmd"
        set -l end (date +%s%3N)
        set -l duration (math $end - $_recall_start)
        recall log \
            --cmd "$_recall_cmd" \
            --exit $exit_code \
            --duration $duration \
            --shell "fish" \
            --session "$_recall_session_id" \
            --cwd "$PWD" \
            >/dev/null 2>&1 &
        disown
    end
    set -g _recall_cmd ""
end
