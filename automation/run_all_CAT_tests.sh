#!/bin/bash

#USAGE="Usage:  ${0}"
#echo -e "$USAGE"

echo -e "Reminder: you may set and export environment variables to specify network configure parameters."
echo -e "Refer to '../automation/go_record.sh' for details and for the default values."

# USE THIS _sigs() signal catcher/forwarder to pass signal to the child process.
trap 'echo $0 Received termination signal.; kill $! 2>/dev/null; exit' SIGHUP SIGINT SIGQUIT SIGTERM SIGABRT

#cd ../CAT
#../automation/go_record.sh CAT*.go

./go_record.sh ../CAT/CAT*.go

