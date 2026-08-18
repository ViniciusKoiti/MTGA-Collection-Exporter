# Inventário de atividades

Tarefa 1.3 do OpenSpec `add-graph-workflow-harness`: todo comando externo e
todo job agendado do companion mapeia para um grafo registrado ou para uma
isenção de consulta pura documentada. O inventário executável vive em
`internal/activity` (o teste `TestInventarioCompletoDoCompanion` monta esta
tabela em código); comandos que não constam aqui não podem ser expostos —
o teste de arquitetura impede `cmd/*` de importar o runtime diretamente.

## Atividades orquestradas (grafo)

| Atividade externa | Grafo (kind) | Origem hoje (Python) | Status |
| --- | --- | --- | --- |
| `sync-collection` | `collection-sync` v1 | GUI "escanear" + MCP `rescan` | grafo pendente (tarefa 4.1) |
| `recommend-decks` | `meta-deck-recommendation` v1 | — (novo) | grafo pendente (tarefa 4.2) |
| `export-approved` | `approved-export` v1 | GUI exportar JSON/CSV/TXT | grafo pendente (tarefa 4.3) |
| `flush-telemetry` | `telemetry-flush` v1 | — (novo, job agendado) | grafo pendente (tarefa 4.4) |
| `run-dev-scenario` | `development-scenario` v1 | — (dev-mcp, tarefa 7.3) | grafo pendente (tarefa 4.5) |

## Isenções de consulta pura

Consultas somente leitura sobre o snapshot corrente, sem efeito, retry,
aprovação ou estado resumível — por definição da decisão 1 do design não
viram grafo. Cada isenção exige justificativa registrada em código.

| Atividade externa | Origem hoje (Python) | Justificativa |
| --- | --- | --- |
| `query-collection` | MCP `get_collection` | leitura direta do snapshot exportado |
| `query-stats` | MCP `collection_stats` | agregação determinística em memória |
| `search-cards` | MCP `search_cards` | filtro puro sobre o snapshot |
| `check-deck` | MCP `check_deck` | aritmética de posse sem efeito |

## Regras

- Novo comando ou job => nova linha aqui + registro em `internal/activity`
  no mesmo commit; `Registry.Validate` reprova atividade de grafo sem
  definição registrada no runtime.
- O MCP Python atual é consumidor legado do export de compatibilidade; suas
  ferramentas migram para as linhas acima quando o companion Go assumir.
