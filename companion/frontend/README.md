# Frontend do companion

Boundary reservado para o shell desktop Wails + TypeScript (tarefa 4.1 do
OpenSpec `introduce-agentic-go-companion`): navegação Home, Collection,
Decks, Assistant e Settings.

Regras de fronteira (decisão 2 do design):

- o frontend apresenta estado e emite comandos tipados; nunca decide
  transições de workflow nem possui estado de domínio;
- pacotes de domínio e aplicação não podem importar nada daqui (teste de
  arquitetura em `internal/arch`);
- comandos externos entram exclusivamente pelo inventário de
  `internal/activity`.
