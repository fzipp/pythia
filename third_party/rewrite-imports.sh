#!/bin/sh

if [ "$1" = "-l" ]
then
  find . -type f -name '*.go' -exec grep -E -l '"code.google.com/' {} \;
elif [ "$1" = "-w" ]
then
  find . -type f -name '*.go' -exec perl -pi -e 's!"code.google.com/p/!"github.com/fzipp/pythia/third_party/!' {} \;
else
  cat << heredoc
usage: rewrite-imports.sh -l|-w"
        -l: list files where a rewrite is needed"
        -w: actually performs the rewrite"
heredoc
fi
