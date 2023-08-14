#!/bin/bash
USER=$(echo "$CI_COMMIT_AUTHOR" | cut -d ' ' -f 1)
DIR=/home/gitlab/artifacts/$USER/$(date +%Y-%m-%d-%H-%M-%S)-MR-$CI_MERGE_REQUEST_IID

check_ret() {
    ret=$1
    name=$2
    if [ "$ret" -ne 0 ]; then
        echo "[${name}] failed, ret=${ret}"
        exit "$ret"
    fi
    return
}

dump() {
    mkdir -p "$DIR"
    check_ret $? "make dir"
    cp -fr build code_check.txt "$DIR"
    check_ret $? "copy artifacts"
}

show() {
    echo http://"$CI_RUNNER":8888/"$USER"/
}

case "$1" in
dump)
  dump
  ;;
show)
  show
  ;;
*)
  echo "Usage: $0 {dump|show}"
  exit 2
  ;;
esac

exit $?