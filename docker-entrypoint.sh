#!/bin/sh

# Check configuration file exists
if [ ! -f $CONFIG_FILE ]; then
    echo "No configuration file at $CONFIG_FILE"
    exit 1
fi

# Test configuration file
bin/pg -t -c $CONFIG_FILE
if [ $? -ne 0 ]; then
    exit 1
fi

# Line web to pg for aesthetics
ln -s $(pwd)/bin/pg $(pwd)/bin/web

# Launch the application
bin/$1 -c $CONFIG_FILE
