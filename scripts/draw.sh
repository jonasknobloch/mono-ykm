#!/usr/bin/env bash

for entry in *
do
  [[ $entry != *.dot ]] && continue
  echo "Drawing $entry..."
  dot -Tsvg -O "$entry"
done