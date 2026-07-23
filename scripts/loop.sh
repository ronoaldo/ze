#!/usr/bin/env bash

TASK_ID="task-$(date +%Y%m%d-%H%M%S).${RANDOM}"

echo "$TASK_ID"

while true ; do
	ze < tmp/prompt 2>&1 | tee -a "tmp/${TASK_ID}"

	if grep "DONE" <( tail -n 3 "tmp/${TASK_ID}" ) ; then
		echo "Detectamos que completou a task"
		break
	else
		echo "Retomando ..."
		sleep 5
	fi
done

