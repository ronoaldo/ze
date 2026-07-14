# Novas Features

Plano de funcionalidades desejáveis para o Zé! Esse plano deve ser implementado com TDD.

## Feature 1: Executar comandos no shell com !

A primeira feature desejada é executar comandos com o !, como "!ls",
e ver o stdout/stderr deste comando na sessão do zé.

O agente deve então acumular os comandos executados até o próximo prompt ser enviado para a IA.
Ao enviar, os comandos executados devem ir para o modelo LLM, como mostrado no exemplo abaixo.

Supondo esta sessão de interação com o zé:

```
ze > !ls
file1.txt

ze > Analise o conteúdo do arquivo listado
```

A ia recebe a combinação do prompt do usuário e o comando que ele rodou:

```
User executed command:
$ ls
file1.txt

Analise o conteúdo do arquivo listado.
```

## Feature 2: Tool WebFetch para baixar da web

Essa tool é o mínimo até termos WebSearch via MCPs.

Essa será a tool que pode ser usada pelo modelo quando o usuário passa
links contendo páginas de documentação, etc.

Ela deve, por segurança, só baixar incluir no contexto formato texto:
HTML, JSON, Markdown, TXT, CSV serão aceitos.

## Feature 3: Renderizar Markdown da resposta no terminal

Renderizador básico para suportar:
* Tabelas
* Listas
* Cabeçalhos e Negrito
* Itálico
* Sublinhado

## Feature 4: Tool de "ver a doc completa" go_doc('all')

Melhorar a tool para rodar algo equivalente a:

    go list ./... | while read pkg ; do go doc -all $pkg ; done

Desta forma conseguimos ter um "full api" da implementação atual.
