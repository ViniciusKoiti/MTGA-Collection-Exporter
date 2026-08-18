# Guia do usuário — MTGA Companion

Task 7.6 do OpenSpec `introduce-agentic-go-companion`. Este guia cobre
confiança das fontes, privacidade, permissões, recuperação e
compatibilidade. Cada afirmação aqui é imposta por código citado na
[matriz de política](policy-matrix.md).

## O companion NÃO automatiza o jogo

O companion **nunca joga por você, nunca envia cliques ou teclas ao MTG
Arena, nunca compra nada e nunca lê a memória do processo do jogo**.
Essas capacidades são impossíveis por construção: não podem sequer ser
registradas como ferramentas (`toolreg`, quinze capacidades provadas
negadas em teste). O único uso do scanner de memória legado é o comando
explícito de compatibilidade (abaixo), executado apenas quando você pede.

## De onde vêm os dados (confiança das fontes)

Em ordem de preferência:

1. **Import de JSON** — você aponta o export existente
   (`mtga_collection.json`); nada roda em segundo plano.
2. **Detailed Logs** *(futuro)* — leitura dos logs oficiais do cliente,
   somente após validação de versão do payload; desabilitado por padrão.
3. **Ponte legada (explícita)** — roda o scanner Python antigo sob prazo
   limitado, importa o resultado e o marca como `legacy_bridge`. Nunca é
   acionável pelo assistente.

Cartas que o catálogo não reconhece permanecem visíveis como "não
resolvidas" — nada é descartado silenciosamente.

## Privacidade

- Sua coleção, decks e logs ficam **no seu computador** (SQLite local).
- Telemetria é **opt-in explícito**; sem consentimento, nada sai do
  dispositivo (`done_local_only` é sucesso). Eventos enviados passam por
  allowlist e nunca incluem caminhos, credenciais ou a lista de cartas.
- O assistente (desligado por padrão) recebe apenas contexto redigido:
  códigos de status, o deck em discussão e totais — caminhos e segredos
  são removidos antes de qualquer provedor, com teste golden de redação.
- Logs locais são estruturados e auto-redigidos; o pacote de diagnóstico
  só inclui as seções que você marcar (opt-in por campo).

## Permissões e aprovações

Leituras (buscar cartas, conferir deck) rodam direto. **Todo efeito
local** — salvar deck, sincronizar, exportar arquivo, copiar para a área
de transferência — pedido pelo assistente exige sua aprovação sobre um
preview exato; o token de aprovação vale uma única vez, expira e cobre
apenas aquele conteúdo (se a coleção mudar, a aprovação antiga é negada).
Toda interação fica auditada na tela de Atividade, só com códigos e
hashes.

## Recuperação

- Snapshots são imutáveis; sincronizações interrompidas retomam do último
  ponto confirmado.
- Se o banco local for perdido, ele é **reconstruído do último export de
  compatibilidade** com marcação de proveniência (`recovered_from_export`).
- Backup frio: copie os arquivos `.db` com o app fechado; a restauração
  reaplica migrações automaticamente.

## Compatibilidade

Os exports (`mtga_collection.json/.csv/.txt`) mantêm o formato do
exportador Python original — o MCP legado continua lendo-os sem mudanças
(paridade verificada por testes que comparam os dois lados sobre a mesma
coleção). O contrato do JSON é congelado: qualquer mudança de formato
quebra testes em vez de corromper dados.
