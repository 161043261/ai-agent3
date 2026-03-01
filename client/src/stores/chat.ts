import type { Message, Session } from "@/types";
import { atom } from "jotai";

// Sessions map
export const sessionsAtom = atom<{
  [sessionId: string]: Session;
}>({});

// Current session ID
export const currentSessionIdAtom = atom<string | null>(null);

// Is temp session (not saved yet)
export const tempSessionAtom = atom<boolean>(false);

// Current messages (derived from current session)
export const currentMessagesAtom = atom<Message[]>([]);

// Selected model
export const selectedModelAtom = atom<string>("1");

// Streaming mode
export const isStreamingAtom = atom<boolean>(false);

// Loading state
export const loadingAtom = atom<boolean>(false);
