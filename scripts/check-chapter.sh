#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "使い方: scripts/check-chapter.sh <章番号>" >&2
  echo "例: scripts/check-chapter.sh 04" >&2
  exit 2
fi

chapter="$1"

case "$chapter" in
  01) tag="chapter01" ;;
  02) tag="chapter02" ;;
  03) tag="chapter03" ;;
  04) tag="chapter04" ;;
  05) tag="chapter05" ;;
  06) tag="chapter06" ;;
  07) tag="chapter07" ;;
  08) tag="chapter08" ;;
  09) tag="chapter09" ;;
  *) echo "未知の章番号です: $chapter" >&2; exit 2 ;;
esac

echo "==> go test -tags $tag ./..."
go test -tags "$tag" ./...
