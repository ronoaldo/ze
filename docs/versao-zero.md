# Zé, seu agente de programação

Zé é um agente de programação desenhado para atuar com modelos locais através do `llama.cpp`.

## Objetivos

O Zé tem como principais objetivos oferecer uma experiência simples e pronta para uso em programação:

* Ele deve suportar o download automático de modelos, inicialmente suportando toda a família de modelos Gemma 4 em seus diferentes tamanhos e quantizações

* Ele escolhe o melhor modelo para o hardware atualmente detectado (recursos de CPU e GPU e quantidade de VRAM)

* Ele possui um conjunto de ferramentas desejáveis para os agentes de codificação:
  
  * Ferramentas que permitem de forma simples criar e editar arquivos de código no diretório do projeto
  
  * Ferramentas de busca na web como Exa AI, para permitir identificar corretamente documentação online
  
  * Ferramenta git para interagir com o controle de versão, suportando inicialmente realizar commit, pull, push e visualizar o histórico com log
  
  * Ferramenta de documentação, suportando inicialmente inspecionar a documentação de bibliotecas Go com `go doc`
  
  * Suporte a configurar ferramentas adicionais via MCP

* Ele possui um system prompt otimizado para o caso de uso de programação
  
  * Este system prompt deve ser otimizado para os modelos Gemma 4 (instruction tunned), de modo a garantir o correto funcionamento na geração de código *lean* e na invocação de ferramentas.

## Stack

Zé é desenhado para ser eficiente e simples, com o mínimo de dependências externas:

* Ele deve ser implementado em Go oferecendo uma TUI básica

* Ele suporta a produção de múltiplos executáveis para diferentes plataformas via Go Release

## Critérios de Aceitação

O primeiro teste de verificação end-to-end é que ele deve ser capaz de orquestar a criação de um código em Go que implementa uma API web capaz de receber por parâmetros de query string o endereço de um servidor Minecraft, e responder essa API com um esquema JSON informando se o servidor está disponível ou não, e quais são os primeiros 18 jogadores conectados nele.


