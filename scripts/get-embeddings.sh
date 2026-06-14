#!/bin/sh

if [ ! -f data/dolma_300_2024_1.2M.100_combined.txt ]; then
	if [ ! -f glove.2024.dolma.300d.zip ]; then
		echo "Downloading 2024 Dolma GloVe vectors"
		wget https://nlp.stanford.edu/data/wordvecs/glove.2024.dolma.300d.zip -q --show-progress
	else
		echo "Skipping Download (zip present)"
	fi

	echo "Extracting Vectors"
	unzip glove.2024.dolma.300d.zip

	echo "Cleaning up"
	mkdir -p data
	mv dolma_300_2024_1.2M.100_combined.txt data/
	rm glove.2024.dolma.300d.zip
else
	echo "Skipping Download (txt present)"
fi
