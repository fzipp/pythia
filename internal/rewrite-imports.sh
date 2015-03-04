#!/bin/sh

if [ "$1" = "-l" ]
then
  find . -type f -name '*.go' -exec grep -E -l '"golang.org/x/' {} \;
elif [ "$1" = "-w" ]
then
  find . -type f -name '*.go' -exec perl -pi -e 's!"golang.org/x/!"github.com/fzipp/pythia/internal/!' {} \;
else
  cat << heredoc
usage: rewrite-imports.sh -l|-w"
        -l: list files where a rewrite is needed"
        -w: actually performs the rewrite"
heredoc
fi
