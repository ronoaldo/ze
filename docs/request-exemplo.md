# Open AI Chat Completions Sample

Request:

```bash
curl -X POST http://localhost:8084/v1/chat/completions -H "Content-Type: application/json" -d '{
  "model": "google/gemma-4-26B",
  "messages": [
    {"role": "user", "content": "Explique o que é um buraco negro de forma concisa."}
  ],
  "temperature": 0.7
}' | jq .
```

Response:

```json
{
  "choices": [
    {
      "finish_reason": "stop",
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Um **buraco negro** é uma região do espaço-tempo com uma força gravitacional tão intensa que nada, nem mesmo a luz, consegue escapar dela.\n\nAqui estão os pontos principais para entender o conceito:\n\n*   **Origem:** Geralmente surgem quando estrelas massivas esgotam seu combustível e colapsam sobre si mesmas devido à própria gravidade.\n*   **Horizonte de Eventos:** É a \"fronteira\" do buraco negro. Uma vez que algo (matéria ou luz) ultrapassa essa linha, não há mais volta; a fuga torna-se impossível.\n*   **Singularidade:** É o centro do buraco negro, um ponto de densidade infinita onde as leis conhecidas da física deixam de funcionar.\n\nEm resumo: é um objeto com tanta massa concentrada em um espaço tão pequeno que sua gravidade cria um \"poço\" sem saída no tecido do universo.",
        "reasoning_content": "*   Topic: Black Hole (Buraco negro).\n*   Constraint: Concise (concisa).\n*   Language: Portuguese.\n\n    *   What is it? A region in spacetime.\n    *   Why does it exist? Intense gravity, usually from a collapsed massive star.\n    *   Key feature: Gravity so strong that nothing, not even light, can escape.\n    *   Key parts: Event horizon (point of no return) and Singularity (center of infinite density).\n\n    *   *Draft 1 (Mental):* Um buraco negro é um lugar no espaço com gravidade muito forte. Ele se forma quando estrelas grandes morrem. A gravidade é tão forte que nem a luz escapa. Tem o horizonte de eventos e a singularidade.\n\n    *   *Draft 2 (Refining for flow and precision):* Um buraco negro é uma região do espaço-tempo com uma força gravitacional tão intensa que nada, nem mesmo a luz, consegue escapar dela. Eles geralmente se formam após o colapso de estrelas massivas. Possuem dois componentes principais: o horizonte de eventos (a fronteira de onde nada escapa) e a singularidade (o centro de densidade infinita).\n\n    *   *Definition:* Focus on gravity and light.\n    *   *Origin:* Collapse of massive stars.\n    *   *Mechanism:* Event horizon and singularity.\n\n    \"Um **buraco negro** é uma região do espaço-tempo com uma força gravitacional tão intensa que nada, nem mesmo a luz, consegue escapar dela.\n\n    Ele se forma, geralmente, quando uma estrela massiva esgota seu combustível e colapsa sobre si mesma. Seus dois componentes principais são:\n\n    1. **Horizonte de Eventos:** A \"fronteira\" ou ponto de não retorno; uma vez cruzada, a fuga é impossível.\n    2. **Singularidade:** O centro do buraco negro, onde a matéria está comprimida em um ponto de densidade infinita.\"\n\nDoes this meet the \"concise\" requirement? Yes. Is it accurate? Yes."
      }
    }
  ],
  "created": 1783729017,
  "model": "google/gemma-4-26B",
  "system_fingerprint": "b9952-2ed3c1abb",
  "object": "chat.completion",
  "usage": {
    "completion_tokens": 650,
    "prompt_tokens": 30,
    "total_tokens": 680,
    "prompt_tokens_details": {
      "cached_tokens": 7
    }
  },
  "id": "chatcmpl-GMlpqKW61ogvqxfFjULQmfJCt16ZYzR6",
  "timings": {
    "cache_n": 7,
    "prompt_n": 23,
    "prompt_ms": 258.248,
    "prompt_per_token_ms": 11.228173913043477,
    "prompt_per_second": 89.0616771475481,
    "predicted_n": 650,
    "predicted_ms": 19000.958,
    "predicted_per_token_ms": 29.232243076923076,
    "predicted_per_second": 34.208801472009995
  }
}
```
