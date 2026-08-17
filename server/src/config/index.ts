import { resolve } from "node:path";
import dotenv from "dotenv";
import pino from "pino";

dotenv.config({ path: resolve(process.cwd(), ".env") });

export const logger = pino({ level: "info" });

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
  password?: string;
}

export interface MysqlConfig {
  port: number;
  host: string;
  user: string;
  password?: string;
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
  base_url: string;
}

export interface OpenaiConfig {
  mode_name: string;
}

export interface Config {
  app: AppConfig;
  redis: RedisConfig;
  mysql: MysqlConfig;
  jwt: JwtConfig;
  rag: RagConfig;
  openai: OpenaiConfig;
}

const config: Config = {
  app: {
    name: process.env.APP_NAME || "swifty-chatbot",
    host: process.env.APP_HOST || "0.0.0.0",
    port: Number.parseInt(process.env.APP_PORT || "8088", 10),
  },
  redis: {
    enabled: process.env.REDIS_ENABLED === "true",
    host: process.env.REDIS_HOST || "127.0.0.1",
    port: Number.parseInt(process.env.REDIS_PORT || "6379", 10),
    password: process.env.REDIS_PASSWORD || "",
    db: Number.parseInt(process.env.REDIS_DB || "0", 10),
  },
  mysql: {
    port: Number.parseInt(process.env.MYSQL_PORT || "3306", 10),
    host: process.env.MYSQL_HOST || "localhost",
    user: process.env.MYSQL_USER || "root",
    password: process.env.MYSQL_PASSWORD || "pass",
    db: process.env.MYSQL_DB || "swifty_chatbot",
    charset: process.env.MYSQL_CHARSET || "utf8mb4",
  },
  jwt: {
    expire_duration: Number.parseInt(
      process.env.JWT_EXPIRE_DURATION || "8760",
      10,
    ),
    issuer: process.env.JWT_ISSUER || "swifty-chatbot",
    subject: process.env.JWT_SUBJECT || "swifty-chatbot",
    key: process.env.JWT_KEY || "swifty-chatbot",
  },
  rag: {
    embedding_model: process.env.EMBEDDING_MODEL || "nomic-embed-text",
    docs_dir: process.env.DOCS_DIR || "./docs",
    dimension: Number.parseInt(process.env.EMBEDDING_DIMENSION || "1024", 10),
    base_url: process.env.EMBEDDING_MODEL_BASE_URL || "",
  },
  openai: {
    mode_name: process.env.OPENAI_MODE_NAME || "qwen3",
  },
};

logger.info("Config loaded from environment variables");

export function getConfig(): Config {
  return config;
}
