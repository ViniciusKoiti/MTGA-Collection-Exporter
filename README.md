# MTGA Collection Exporter

Ferramenta que lê sua coleção do **MTG Arena** direto da memória do jogo e a
disponibiliza de duas formas:

- 🖥️ **App desktop** — configure tudo por uma interface visual e veja suas
  cartas numa tabela com busca e filtros.
- 🤖 **Servidor MCP** — deixa uma IA (Claude) acessar sua coleção para ajudar
  a montar decks só com o que você tem.

## Arquitetura

Pacote `mtga/` com módulos pequenos e focados; duas "frentes" (GUI e MCP)
reaproveitam o mesmo núcleo:

```
mtga/
├── config.py        ← constantes, caminhos, dataclass Config (config.json)
├── database.py      ← banco de cartas (SQLite local + Scryfall + cache)
├── scanner.py       ← leitura de memória do MTGA.exe (pymem) + exceções
├── collection.py    ← enriquecer / exportar / reler a coleção
├── pipeline.py      ← orquestração run_full_scan
├── mcp_server.py    ← ferramentas MCP
└── gui/
    ├── theme.py         ← aparência + estilo da tabela
    ├── app.py           ← janela principal + scan
    ├── config_tab.py    ← aba Configuração
    └── collection_tab.py← aba Coleção
run_gui.py           ← atalho: abre o app desktop
run_mcp.py           ← atalho: sobe o servidor MCP
```

A lógica de leitura fica no núcleo (`scanner`/`pipeline`); GUI e MCP só consomem.

## Instalação

```bash
pip install -r requirements.txt
```

> ⚠️ A leitura de memória usa `pymem` e funciona **somente no Windows**, com o
> MTG Arena aberto. Rode como administrador se o scan falhar.

## App desktop

```bash
python run_gui.py
```

1. Aba **Configuração**: defina a pasta `Raw` do MTGA (ou clique em *Detectar*),
   a pasta de saída, a fonte do banco de cartas e as **âncoras de calibração**
   (3–5 cartas raras/míticas que você possui, com a quantidade exata).
2. Clique em **Salvar configuração**.
3. Com o MTG Arena aberto na aba *Decks*, clique em **▶ Escanear Coleção**.
4. Veja as cartas na aba **Coleção** (busca + filtro por raridade). Os arquivos
   `mtga_collection.{txt,json,csv}` são gravados na pasta de saída.
5. Ao final do scan, o app mostra um status de validação (`OK`, `Atencao` ou
   `Erro`) e grava `mtga_scan_validation.json` com os detalhes.

Tudo fica salvo em `config.json`, então nas próximas vezes é só abrir e escanear.

## Servidor MCP (integração com IA)

O MCP expõe a coleção exportada para uma IA. Por padrão lê o
`mtga_collection.json` (não precisa do jogo aberto).

**Ferramentas expostas:**

| Ferramenta | O que faz |
|------------|-----------|
| `get_collection()` | Retorna a coleção inteira |
| `collection_stats()` | Resumo por raridade, cor e set |
| `search_cards(query, colors, rarity, type_contains, min_count)` | Filtra cartas |
| `check_deck(cards)` | Verifica o que falta para montar uma decklist |
| `rescan()` | Refaz a leitura ao vivo (precisa do MTGA aberto) |

**Claude Desktop** — em `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mtga-collection": {
      "command": "python",
      "args": ["C:/caminho/para/run_mcp.py"]
    }
  }
}
```

**Claude Code**:

```bash
claude mcp add mtga-collection -- python C:/caminho/para/run_mcp.py
```

Depois é só pedir, por exemplo: *"monte um deck de aggro vermelho só com as
cartas que eu tenho"* — a IA consulta o MCP e usa sua coleção real.

## Empacotamento (.exe)

```bash
pip install pyinstaller
pyinstaller --onefile --windowed --name "MTGA Exporter" run_gui.py
```

## Observações

- A **fonte do banco** pode ser `local` (arquivos `.mtga` do jogo), `scryfall`
  (baixa metadados ricos: cores, tipo, mana) ou `auto` (tenta local, cai para
  Scryfall). Para o MCP montar decks bem, `scryfall` ou `auto` é o ideal, pois
  traz cores/tipos das cartas.
- O banco é cacheado em `arena_id_lookup.json`. Apague o arquivo para forçar
  uma atualização.
- O validador compara o scan atual com a exportação anterior e avisa sobre
  coleção vazia, poucas âncoras, banco aparentemente pequeno ou queda grande de
  cartas/cópias desde o último scan.
