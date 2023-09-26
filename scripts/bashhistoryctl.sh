#!/bin/bash
if [ a"$HISTFILE" != a"" ]; then
    BASH_HISTORY_FILE="$HISTFILE"
else
    if [ a"$HOME" != a"" ]; then
        BASH_HISTORY_FILE="$HOME/.bash_history"
    else
        BASH_HISTORY_FILE="/home/$(whoami)/.bash_history"
    fi
fi

dump() {
    if [ a"$1" == a"" ]; then
        echo "Usage: $0 dump <output_file>"
        exit 2
    fi
    cat "$BASH_HISTORY_FILE" >"$1"
}

case "$1" in
"dump")
    dump "$2"
    ;;
"show")
    echo "$BASH_HISTORY_FILE"
    ;;
*)
    echo "Usage: $0 [dump|show]"
    exit 2
    ;;
esac

exit $?
