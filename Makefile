.PHONY: latex clean

latex: clean
	cd latex; mkdir -p out &&  latexmk -pdf -jobname=out/tcc main.tex; cd ..

clean:
	rm -r latex/out
