import { AiAgent } from "./agent";
import {
  type AiModel,
  ModelType,
  OllamaModel,
  OllamaRagModel,
  OpenaiModel,
  OpenaiRagModel,
} from "./model";

export type AiModelCreator = (config: Record<string, unknown>) => AiModel;

class AiModelFactory {
  private creators: Map<string, AiModelCreator> = new Map();

  constructor() {
    this.registerCreators();
  }

  private registerCreators(): void {
    this.creators.set(ModelType.OPENAI_MODEL, () => new OpenaiModel());

    this.creators.set(ModelType.OLLAMA_MODEL, () => new OllamaModel());

    this.creators.set(ModelType.OPENAI_RAG_MODEL, (cfg) => {
      const username = String(cfg.username);
      if (!username) throw new Error("RAG model requires username");
      return new OpenaiRagModel(username);
    });

    this.creators.set(ModelType.OLLAMA_RAG_MODEL, (cfg) => {
      const username = String(cfg.username);
      if (!username) throw new Error("RAG model requires username");
      return new OllamaRagModel(username);
    });
  }

  createAiModel(
    modelType: ModelType,
    config: Record<string, unknown>,
  ): AiModel {
    const creator = this.creators.get(modelType);
    if (!creator) {
      throw new Error(`Unsupported model type: ${modelType}`);
    }
    return creator(config);
  }

  createAiAgent(
    modelType: ModelType,
    sessionId: string,
    config: Record<string, unknown>,
  ): AiAgent {
    const model = this.createAiModel(modelType, config);
    return new AiAgent(model, sessionId);
  }

  registerModel(modelType: ModelType, creator: AiModelCreator) {
    this.creators.set(modelType, creator);
  }
}

let factoryInstance: AiModelFactory | null = null;

export function getAiModelFactory(): AiModelFactory {
  if (!factoryInstance) {
    factoryInstance = new AiModelFactory();
  }
  return factoryInstance;
}
