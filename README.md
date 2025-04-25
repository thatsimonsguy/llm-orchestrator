# LLM-Orchestrator

The `llm-orchestrator` is the central microservice that coordinates the Retrieval-Augmented Generation (RAG) pipeline for **MattBot**, the AI assistant available on [matthewpsimons.com](https://matthewpsimons.com).

It acts as the brain of the system, handling client interactions, orchestrating supporting services, building prompts, and streaming responses back to users.

---

### Step-by-Step Flow of MattBot

1. **[Web Client](https://github.com/thatsimonsguy/career-site) ➔ LLM-Orchestrator**
   - A user submits a query from the web client on [matthewpsimons.com ](https://matthewpsimons.com)
   - The request is sent to the `llm-orchestrator` API.

2. **LLM-Orchestrator ➔ [Embedding Service](https://github.com/thatsimonsguy/embedding-service)**
   - `llm-orchestrator` forwards the raw user query to the Embedding Service.
   - The Embedding Service transforms the query into a vector embedding.

3. **[Embedding Service](https://github.com/thatsimonsguy/embedding-service) ➔ LLM-Orchestrator**
   - The embedding is returned to `llm-orchestrator`.

4. **LLM-Orchestrator ➔ [Qdrant (Vector Database)](https://github.com/thatsimonsguy/homelab-k8s/tree/main/apps/qdrant)**
   - `llm-orchestrator` sends the embedding to Qdrant to perform a similarity search.
   - Qdrant returns relevant content chunks from the knowledge base.

5. **Prompt Building (LLM-Orchestrator)**
   - `llm-orchestrator` builds a final prompt using:
     - Retrieved chunks
     - Canonical facts about Matt Simons (work history, skills, philosophy)
     - The original user query

6. **LLM-Orchestrator ➔ General LLM (OpenAI or [Mistral](https://github.com/thatsimonsguy/homelab-k8s/tree/main/apps/ollama))**
   - The constructed prompt is sent to a large language model (configurable):
     - In production: OpenAI API
     - Locally: Self-hosted Mistral model

7. **LLM-Orchestrator Streams Responses ➔ [Web Client](https://github.com/thatsimonsguy/career-site)**
   - As the general LLM generates tokens, `llm-orchestrator` streams the response back to the web client in real time.

---

## Key Responsibilities

- Handle incoming user requests
- Request embeddings from the embedding microservice
- Perform chunk retrieval from Qdrant
- Build context-rich prompts
- Dispatch prompts to language models
- Stream generated responses to the user interface

## Related Services

- **[Web Client](https://github.com/thatsimonsguy/career-site)**: Frontend interface where users interact with MattBot
- **[Embedding Service](https://github.com/thatsimonsguy/embedding-service)**: Lightweight service that generates embeddings from text inputs
- **[Qdrant (Vector Database)](https://github.com/thatsimonsguy/homelab-k8s/tree/main/apps/qdrant)**: Vector database for storing and retrieving semantically relevant chunks
- **OpenAI API / Local Mistral**: Language models that generate human-like responses

## Development Notes

- Written in Golang
- Designed for minimal latency and high modularity
- Uses standard HTTP APIs for service-to-service communication
- Supports streaming responses via Server-Sent Events (SSE)

---
## License

© 2025 Matthew Simons. All rights reserved.

---

## Contact

Questions? Ideas? Feel free to reach out through [matthewpsimons.com](https://matthewpsimons.com).

---

Built with passion, pragmatism, and a touch of madness ✨.

