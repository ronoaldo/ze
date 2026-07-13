# Novas Features

Plano de funcionalidades desejáveis para o Ze.

## Executar comandos no shell com !

A primeira feature desejada é executar comandos com o !, como "!ls", e ver o stdout/stderr deste comando na sessão do zé.
O output destes comandos será enviado para o modelo de AI, o que permite que ela veja o que fizemos.
Isso também permite "embutir" prompts salvos em arquivos com "!cat docs/todo.md"

## Tool WebFetch para baixar da web

Essa tool é o mínimo até termos WebSearch via MCPs.

Essa será a tool que pode ser usada pelo modelo quando o usuário passa
links contendo páginas de documentação, etc.

Ela deve, por segurança, só baixar incluir no contexto formato texto:
HTML, JSON, Markdown, TXT, CSV serão aceitos.

## Renderizar Markdown da resposta no terminal

Renderizador básico para suportar:
* Tabelas
* Listas
* Cabeçalhos e Negrito
* Itálico
* Sublinhado

## Tool de "ver a doc completa" go_doc('all')

Melhorar a tool para rodar algo equivalente a:

    go list ./... | while read pkg ; do go doc -all $pkg ; done

Desta forma conseguimos ter um "full api" da implementação atual.
