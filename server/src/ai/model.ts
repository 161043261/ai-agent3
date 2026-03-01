import { type BaseMessage, HumanMessage } from "@langchain/core/messages";
import { ChatOllama } from "@langchain/ollama";
import { ChatOpenAI } from "@langchain/openai";
import { getConfig, logger } from "../config/config";
import {
  buildRagPrompt,
  type DocumentRetriever,
  newDocumentRetriever,
} from "../rag/rag";

export enum ModelType {
  OPENAI_MODEL = "openai",
  OLLAMA_MODEL = "ollama",
  OPENAI_RAG_MODEL = "openai-rag",
  OLLAMA_RAG_MODEL = "ollama-rag",
}

export type StreamCallback = (chunk: string) => void;

export interface AiModel {
  response(messages: BaseMessage[]): Promise<string>;
  responseStream(messages: BaseMessage[], cb: StreamCallback): Promise<string>;
  getModelType(): string;
}

function createOllamaLlm(): ChatOllama {
  const cfg = getConfig().ai;
  return new ChatOllama({
    baseUrl: cfg.base_url,
    model: cfg.mode_name,
  });
}

function createOpenaiLlm(): ChatOpenAI {
  const cfg = getConfig().ai;
  return new ChatOpenAI({
    configuration: { baseURL: cfg.base_url + "/v1" },
    modelName: cfg.mode_name,
    openAIApiKey: cfg.api_key || "dummy",
  });
}

// --- OpenAI Model ---
export class OpenaiModel implements AiModel {
  private llm: ChatOpenAI;

  constructor() {
    this.llm = createOpenaiLlm();
  }

  getModelType(): ModelType {
    return ModelType.OPENAI_MODEL;
  }

  async response(messages: BaseMessage[]): Promise<string> {
    const res = await this.llm.invoke(messages);
    return typeof res.content === "string"
      ? res.content
      : JSON.stringify(res.content);
  }

  async responseStream(
    messages: BaseMessage[],
    cb: StreamCallback,
  ): Promise<string> {
    const stream = await this.llm.stream(messages);
    let fullContent = "";
    for await (const chunk of stream) {
      const text = typeof chunk.content === "string" ? chunk.content : "";
      if (text.length > 0) {
        cb(text);
        fullContent += text;
      }
    }
    return fullContent;
  }
}

// --- Ollama Model ---
export class OllamaModel implements AiModel {
  private llm: ChatOllama;

  constructor() {
    this.llm = createOllamaLlm();
  }

  getModelType(): ModelType {
    return ModelType.OLLAMA_MODEL;
  }

  async response(messages: BaseMessage[]): Promise<string> {
    if (messages.length === 0) throw new Error("messages is empty");
    const res = await this.llm.invoke(messages);
    return typeof res.content === "string"
      ? res.content
      : JSON.stringify(res.content);
  }

  async responseStream(
    messages: BaseMessage[],
    cb: StreamCallback,
  ): Promise<string> {
    if (messages.length === 0) throw new Error("messages is empty");
    const stream = await this.llm.stream(messages);
    let fullContent = "";
    for await (const chunk of stream) {
      const text = typeof chunk.content === "string" ? chunk.content : "";
      if (text.length > 0) {
        cb(text);
        fullContent += text;
      }
    }
    return fullContent;
  }
}

// --- RAG helper: shared logic ---
async function ragEnhanceMessages(
  messages: BaseMessage[],
  username: string,
): Promise<BaseMessage[]> {
  let documentRetriever: DocumentRetriever;
  try {
    documentRetriever = await newDocumentRetriever(username);
  } catch (err) {
    logger.warn(
      { err },
      "Create RAG queries error, user may not have uploaded any files",
    );
    return messages;
  }

  const lastMessage = messages[messages.length - 1];
  const latestContent =
    typeof lastMessage.content === "string" ? lastMessage.content : "";
  try {
    const docs = await documentRetriever.retrieveDocuments(latestContent);
    if (docs.length === 0) return messages;
    const ragPrompt = buildRagPrompt(latestContent, docs);
    const newMessages = [...messages.slice(0, -1), new HumanMessage(ragPrompt)];
    return newMessages;
  } catch (err) {
    logger.warn({ err }, "Retrieve documents error");
    return messages;
  }
}

// --- OpenAI RAG Model ---
export class OpenaiRagModel implements AiModel {
  private llm: ChatOpenAI;
  private username: string;

  constructor(username: string) {
    this.llm = createOpenaiLlm();
    this.username = username;
  }

  getModelType(): ModelType {
    return ModelType.OPENAI_RAG_MODEL;
  }

  async response(messages: BaseMessage[]): Promise<string> {
    if (messages.length === 0) throw new Error("messages is empty");
    const enhanced = await ragEnhanceMessages(messages, this.username);
    const res = await this.llm.invoke(enhanced);
    return typeof res.content === "string"
      ? res.content
      : JSON.stringify(res.content);
  }

  async responseStream(
    messages: BaseMessage[],
    cb: StreamCallback,
  ): Promise<string> {
    if (messages.length === 0) throw new Error("messages is empty");
    const enhanced = await ragEnhanceMessages(messages, this.username);
    return this.doResponseStream(enhanced, cb);
  }

  private async doResponseStream(
    messages: BaseMessage[],
    cb: StreamCallback,
  ): Promise<string> {
    const stream = await this.llm.stream(messages);
    let fullContent = "";
    for await (const chunk of stream) {
      const text = typeof chunk.content === "string" ? chunk.content : "";
      if (text.length > 0) {
        cb(text);
        fullContent += text;
      }
    }
    return fullContent;
  }
}

// --- Ollama RAG Model ---
export class OllamaRagModel implements AiModel {
  private llm: ChatOllama;
  private username: string;

  constructor(username: string) {
    this.llm = createOllamaLlm();
    this.username = username;
  }

  getModelType(): ModelType {
    return ModelType.OLLAMA_RAG_MODEL;
  }

  async response(messages: BaseMessage[]): Promise<string> {
    if (messages.length === 0) throw new Error("messages is empty");
    const enhanced = await ragEnhanceMessages(messages, this.username);
    const res = await this.llm.invoke(enhanced);
    return typeof res.content === "string"
      ? res.content
      : JSON.stringify(res.content);
  }

  async responseStream(
    messages: BaseMessage[],
    cb: StreamCallback,
  ): Promise<string> {
    if (messages.length === 0) throw new Error("messages is empty");
    const enhanced = await ragEnhanceMessages(messages, this.username);
    return this.doResponseStream(enhanced, cb);
  }

  private async doResponseStream(
    messages: BaseMessage[],
    cb: StreamCallback,
  ): Promise<string> {
    const stream = await this.llm.stream(messages);
    let fullContent = "";
    for await (const chunk of stream) {
      const text = typeof chunk.content === "string" ? chunk.content : "";
      if (text.length > 0) {
        cb(text);
        fullContent += text;
      }
    }
    return fullContent;
  }
}
