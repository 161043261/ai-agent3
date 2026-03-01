import { readFileSync } from "fs";
import merge from "lodash-es/merge.js";
import { resolve } from "path";
import pino from "pino";

export const logger = pino({ level: "info" });

export const PROVIDER_OLLAMA = "ollama";
export const PROVIDER_OPENAI = "openai";

export interface AppConfig {
  port: number;
  name: string;
  host: string;
}

export interface RedisConfig {
  enabled: boolean;
  port: number;
  db: number;
  host: string;
  password: string;
}

export interface MysqlConfig {
  port: number;
  host: string;
  user: string;
  password: string;
  db: string;
  charset: string;
}

export interface JwtConfig {
  expire_duration: number;
  issuer: string;
  subject: string;
  key: string;
}

export interface RagConfig {
  embedding_model: string;
  docs_dir: string;
  dimension: number;
}

export interface AiConfig {
  provider: string;
  api_key: string;
  mode_name: string;
  base_url: string;
}

export interface Config {
  app: AppConfig;
  redis: RedisConfig;
  mysql: MysqlConfig;
  jwt: JwtConfig;
  rag: RagConfig;
  ai: AiConfig;
}

const defaultConfig: Config = {
  app: {
    name: "ai-agent3",
    host: "0.0.0.0",
    port: 8080,
  },
  redis: {
    enabled: true,
    host: "127.0.0.1",
    port: 6379,
    password: "pass",
    db: 0,
  },
  mysql: {
    port: 3306,
    host: "localhost",
    user: "root",
    password: "pass",
    db: "ai_agent3",
    charset: "utf8mb4",
  },
  jwt: {
    expire_duration: 8760,
    issuer: "ai-agent3",
    subject: "ai-agent3",
    key: "ai-agent3",
  },
  rag: {
    embedding_model: "nomic-embed-text",
    docs_dir: "./docs",
    dimension: 1024,
  },
  ai: {
    provider: "ollama",
    api_key: "",
    mode_name: "qwen2.5:7b",
    base_url: "http://localhost:11434",
  },
};

let config: Config = { ...defaultConfig };

function loadConfig() {
  try {
    const configPath = resolve(process.cwd(), "../config.json");
    const data = readFileSync(configPath, "utf-8");
    const fileConfig = JSON.parse(data);
    config = merge<object, Config, Partial<Config>>(
      {},
      defaultConfig,
      fileConfig,
    );
    logger.info("Config loaded");
  } catch (err) {
    logger.error(err, "Read config.json error, using defaults");
  }
}

loadConfig();

export function getConfig(): Config {
  return config;
}
