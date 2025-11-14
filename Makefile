.PHONY: python latex clean venv

VENV := src/core_python/.venv
PYTHON := $(VENV)/bin/python
SCRIPT := src/core_python/main.py

venv:
	@if [ ! -d "$(VENV)" ]; then \
		echo ">> Criando venv em $(VENV)..."; \
		python3.11 -m venv $(VENV); \
	fi

python: venv
	@echo ">> Instalando dependências e executando script..."
	@. $(VENV)/bin/activate; \
	python3.11 -m ensurepip --upgrade; \
	pip install --upgrade pip; \
	pip install -r src/core_python/requirements.txt; \
	clear && echo "executando python..." && python3.11 $(SCRIPT)

latex:
	cd latex; mkdir -p out && latexmk -pdf -jobname=out/tcc main.tex; cd ..

latex-live:
	cd latex; mkdir -p out && latexmk -pdf -pvc -jobname=out/tcc main.tex; cd ..

clean:
	-@rm -rf latex/out $(VENV) __pycache__ *.pyc
