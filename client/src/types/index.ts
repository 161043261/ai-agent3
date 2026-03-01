export interface Message {
  role: "user" | "ai";
  content: string;
  status?: "thinking" | "streaming" | "done" | "error";
}

export interface LoginResponse {
  code: number;
  message?: string;
  token?: string;
  username?: string;
}

export interface Session {
  id: string;
  name: string;
  messages: Message[];
}

export interface BaseResponse {
  code: number;
  message?: string;
}

export interface SessionsResponse extends BaseResponse {
  sessions?: {
    id: string;
    title: string;
  }[];
}

export interface HistoryResponse extends BaseResponse {
  history: {
    is_user: boolean;
    content: string;
  }[];
}

export interface ChatResponse extends BaseResponse {
  session_id?: string;
  answer?: string;
}

export type ModelType = "openai" | "ollama" | "openai-rag" | "ollama-rag";

export const MODELS: Record<string, ModelType> = {
  OPENAI_MODEL: "openai",
  OLLAMA_MODEL: "ollama",
  OPENAI_RAG_MODEL: "openai-rag",
  OLLAMA_RAG_MODEL: "ollama-rag",
};
