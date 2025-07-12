#!/bin/bash

# Set Go proxy and sum database exclusions for private repositories
export GOPRIVATE='github.com/sakibcoolz/*'
export GONOPROXY='github.com/sakibcoolz/*'
export GONOSUMDB='github.com/sakibcoolz/*'

echo "Go proxy configuration set:"
echo "GOPRIVATE=$GOPRIVATE"
echo "GONOPROXY=$GONOPROXY"
echo "GONOSUMDB=$GONOSUMDB"

# Execute the command passed as arguments
if [ $# -gt 0 ]; then
    exec "$@"
else
    # If no command provided, start a new shell with the environment
    exec "$SHELL"
fi
