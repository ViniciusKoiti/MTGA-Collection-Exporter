# Qualidade e metricas

Este projeto tem tres superficies diferentes: leitura de memoria do MTGA,
transformacao/exportacao da colecao, e interfaces GUI/MCP. As metricas abaixo
foram escolhidas para proteger esses pontos sem exigir o MTG Arena real nos
testes automatizados.

## Limites iniciais

| Area | Metrica | Limite inicial |
| --- | --- | --- |
| Testes | Cobertura total do pacote `mtga` | `>= 70%` |
| Testes | Core sem GUI (`collection`, `database`, `scanner`, `pipeline`) | buscar `>= 85%` |
| Complexidade | Complexidade ciclomática por funcao | maximo `B` no `xenon` |
| Manutencao | Maintainability index | acompanhar com `radon mi` |
| Lint | Ruff | zero erros |
| Seguranca | Bandit | zero achados medios/altos |
| Dependencias | pip-audit | zero vulnerabilidades conhecidas |
| Memoria | Crescimento em loop offline | menor que `1 MB` no teste atual |
| Exportacao | JSON/CSV/TXT | mesma quantidade de cartas exportadas |
| Validacao | Status do scan | `OK`, `Atencao` ou `Erro` em `mtga_scan_validation.json` |

Os limites devem ser apertados depois que houver mais testes. Uma boa proxima
meta e subir a cobertura total para `80%`.

## Comandos

Instale as dependencias de desenvolvimento:

```bash
pip install -r requirements-dev.txt
```

Rode os testes:

```bash
pytest
```

Rode com cobertura:

```bash
pytest --cov=mtga --cov-report=term-missing
```

Rode lint:

```bash
ruff check mtga tests
```

Rode complexidade:

```bash
radon cc mtga -s
radon mi mtga
xenon --max-absolute B --max-modules A --max-average A mtga
```

Rode seguranca:

```bash
bandit -q -r mtga -c pyproject.toml -ll
pip-audit -r requirements.txt
```

## Estrategia de testes

- `tests/unit`: testes rapidos e offline.
- `tests/performance`: budgets simples de memoria/performance usando dados fake.
- Testes do scanner devem usar `FakePymem`; nao dependem do MTG Arena aberto.
- Testes de rede/Scryfall devem usar mocks de `requests.get`.
- Testes de exportacao devem usar `tmp_path`, nunca arquivos reais do usuario.
- Testes do validador devem usar colecoes pequenas em memoria e snapshots fake.

## Proximos limites recomendados

- Medir tempo do `run_full_scan` com banco em cache e falhar se piorar mais que
  `25%` em relacao ao baseline local.
- Medir taxa de cartas desconhecidas apos `enrich_collection`; idealmente ficar
  abaixo de `1%` quando a base estiver atualizada.
- Adicionar testes do MCP para `collection_stats`, `search_cards` e `check_deck`
  sem subir servidor real.
- Adicionar um smoke test manual documentado para `python run_gui.py` e
  `python run_mcp.py`.
