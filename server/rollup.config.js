import commonjs from "@rollup/plugin-commonjs";
import json from "@rollup/plugin-json";
import resolve from "@rollup/plugin-node-resolve";
import typescript from "@rollup/plugin-typescript";

export default {
  input: "src/main.ts",
  output: {
    file: "dist/main.js",
    format: "esm",
    sourcemap: true,
  },
  external: [
    /^node:/,
    /^@langchain\//,
    /^@koa\//,
    "langchain",
    "axios",
    "crc-32",
    "dotenv",
    "ioredis",
    "jsonwebtoken",
    "knex",
    "koa",
    "koa-static",
    "lodash-es",
    "lru-cache",
    "mysql2",
    "openai",
    "pino",
    "zod",
  ],
  plugins: [
    resolve({ extensions: [".ts", ".js"] }),
    commonjs(),
    json(),
    typescript({
      tsconfig: "./tsconfig.json",
      declaration: false,
      sourceMap: true,
    }),
  ],
};
