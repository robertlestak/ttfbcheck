#!/bin/bash
set -e

echo "Checking the ttfb..."

URL=$1
TOTAL=${2:-1000}
IFS=',' read -r -a CONCURRENTS <<< "$3"

if [[ ${#CONCURRENTS[@]} -eq 0 ]]; then
	CONCURRENTS=(1 10 50 100)
fi
cp plot.gnu.template plot.gnu.tmp
for i in "${CONCURRENTS[@]}"; do
	echo "Running $TOTAL total with concurrency: $i"
	docker run \
		ttfbcheck -url $URL \
		-concurrent $i \
		-total $TOTAL \
		| jq -r '.TTFB' > plot-$i.tsv
	
	# adding series number to the file
	cat -n plot-$i.tsv > plot.tsv.1 && mv plot.tsv.1 plot-$i.tsv
	echo "'plot-$i.tsv' using 1:2 smooth bezier title '$URL $i concurrent' with lines, \\" >> plot.gnu.tmp
done

# plot the data
gnuplot plot.gnu.tmp

# clean up
rm *.tsv
rm plot.gnu.tmp
