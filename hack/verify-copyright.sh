#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/..
patch_files=$(python3 $ROOT/hack/boilerplate/boilerplate.py)

boilerplate_path=$ROOT/hack/boilerplate/boilerplate.go.txt
year=$(date +%Y)
cat $boilerplate_path | sed "s/YEAR/$year/" >tmp.txt

echo $ROOT
updated=0
for line in $patch_files; do
	cat tmp.txt >tmp.go
	cat $line >>tmp.go
	cat tmp.go >$line
	rm -rf tmp.go
	updated=1
done

rm -rf tmp.txt
if [ $updated -eq 0 ]; then
	echo "already update files copyright"
fi
echo "copyright verified"
