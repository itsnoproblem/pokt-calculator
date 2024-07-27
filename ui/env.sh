#!/bin/bash

OUTFILE=./src/configuration.ts
ENVFILE=$1

if [ "$ENVFILE" == "" ]
then
	ENVFILE=.env.prod
fi

# Recreate config file
rm -f $OUTFILE
touch $OUTFILE

# Read each line in .env file
# Each line represents key=value pairs
while read -r line || [[ -n "$line" ]];
do
  # Split env variables by character `=`
  if printf '%s\n' "$line" | grep -q -e '='; then
    varname=$(printf '%s\n' "$line" | sed -e 's/=.*//')
    varvalue=$(printf '%s\n' "$line" | sed -e 's/^[^=]*=//')
  fi

  # Read value of current variable if exists as Environment variable
  value=$(printf '%s\n' "${!varname}")
  # Otherwise use value from .env file
  [[ -z $value ]] && value=${varvalue}

  # Append configuration property to JS file
  echo "export const $varname = \"$value\";
" | tee -a $OUTFILE




done < $ENVFILE


