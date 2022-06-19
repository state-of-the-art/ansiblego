#/bin/sh
# Script to simplify the style check process

root_dir=$(realpath "$(dirname "$0")")
errors=0

echo
echo '---------------------- Custom Checks ----------------------'
echo
for f in `git ls-files`; do
    # Check text files
    if grep -qI . "$f"; then
        # Ends with newline as POSIX requires
        if [ -n "$(tail -c 1 "$f")" ]; then
            echo "Not ends with newline: $f"
            errors=$((${errors}+1))
        fi
    fi
done


echo
echo '---------------------- GoFmt verify ----------------------'
echo
reformat=$(gofmt -l .)
if [ "${reformat}" ]; then
    echo "Please run 'gofmt -w .':"
    echo "${reformat}"
    errors=$((${errors}+$(echo "${reformat}" | wc -l)))
fi


echo
echo '---------------------- GoModTidy verify ----------------------'
echo
cp -af go.mod go.sum /tmp/
tidy=$(go mod tidy -v)
if [ "${tidy}" -o "x$(date -r /tmp/go.mod ; date -r /tmp/go.sum)" != "x$(date -r go.mod ; date -r go.sum)" ]; then
    echo "Please run 'go mod tidy -v':"
    echo "${tidy}"
    errors=$((${errors}+$(echo "${tidy}" | wc -l)))
fi
mv /tmp/go.mod /tmp/go.sum ./


exit ${errors}
