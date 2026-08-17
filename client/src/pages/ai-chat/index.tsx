import { Sparkles } from "lucide-react";
import { useMemo, useCallback, useRef, useState } from "react";
import useFileUpload from "@/hooks/use-file-upload";
import SessionSidebar from "./components/session-sidebar";
import ChatHeader from "./components/chat-header";
import ChatInput from "./components/chat-input";
import MessageList from "./components/message-list";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
  useSessions,
  useChatHistory,
  useCreateSessionAndSendMessage,
  useSendMessage2Session,
  useStreamMessage,
} from "@/hooks/queries";
import { MODELS, type ModelType, type Message, type Session } from "@/types";

function AiChat() {
  const { t } = useTranslation();
  const [sessions, setSessions] = useState<{ [sessionId: string]: Session }>(
    {},
  );
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(
    "temp",
  );
  const [tempSession, setTempSession] = useState(true);
  const [currentMessages, setCurrentMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedModel, setSelectedModel] = useState<ModelType>(
    MODELS.OPENAI_MODEL,
  );
  const [isStreaming, setIsStreaming] = useState(false);
  const [fetchHistorySessionId, setFetchHistorySessionId] = useState<
    string | null
  >(null);
  const [historyEnabled, setHistoryEnabled] = useState(false);

  // Query: load sessions
  const sessionsQuery = useSessions();

  const createNewSession = useCallback(() => {
    setCurrentSessionId("temp");
    setTempSession(true);
    setCurrentMessages([]);
  }, []);

  // Sync query data into local state during render (guarded adjust pattern)
  const sessionsData = sessionsQuery.data;
  const [prevSessionsData, setPrevSessionsData] = useState(sessionsData);
  if (sessionsData !== prevSessionsData) {
    setPrevSessionsData(sessionsData);
    if (
      sessionsData?.code === 1000 &&
      sessionsData.sessions &&
      sessionsData.sessions.length
    ) {
      const sessionId2title: { [sessionId: string]: Session } = {};
      for (const { id, title } of sessionsData.sessions) {
        sessionId2title[id] = {
          id,
          name: title || `${t("chat.session")} ${id}`,
          messages: [],
        };
      }
      setSessions(sessionId2title);
    } else {
      createNewSession();
    }
  }

  // Query: load chat history
  const chatHistoryQuery = useChatHistory(
    fetchHistorySessionId,
    historyEnabled,
  );

  const historyData = chatHistoryQuery.data;
  const [prevHistoryData, setPrevHistoryData] = useState(historyData);
  if (historyData !== prevHistoryData) {
    setPrevHistoryData(historyData);
    if (
      historyData?.code === 1000 &&
      historyData.history &&
      fetchHistorySessionId
    ) {
      const messages: Message[] = historyData.history.map(
        ({ is_user, content }) => ({
          role: is_user ? "user" : "ai",
          content,
        }),
      );
      setSessions((prev) => ({
        ...prev,
        [fetchHistorySessionId]: {
          ...prev[fetchHistorySessionId],
          messages,
        },
      }));
      setCurrentMessages(messages);
      setHistoryEnabled(false);
    }
  }

  // Mutation: create session + send message
  const createSessionMutation = useCreateSessionAndSendMessage();

  // Mutation: send message to existing session
  const sendMessageMutation = useSendMessage2Session();

  // Buffer for in-flight streamed text. Chunks land here (zero React renders);
  // TypingText reads it directly, and the final content is committed to state
  // exactly once when the stream settles.
  const streamTextRef = useRef("");

  /** Flush the buffered stream into the trailing AI message in a single setState. */
  const commitStreamedMessage = useCallback((status: "done" | "error") => {
    const content = streamTextRef.current;
    streamTextRef.current = "";
    setCurrentMessages((prev) => {
      const newMessages = [...prev];
      const lastIdx = newMessages.length - 1;
      if (newMessages[lastIdx]?.role === "ai") {
        newMessages[lastIdx] = {
          ...newMessages[lastIdx],
          content,
          status,
        };
      }
      return newMessages;
    });
  }, []);

  // Streaming mutation
  const streamMutation = useStreamMessage({
    onSessionCreated: (sessionId: string) => {
      if (tempSession) {
        setSessions((prev) => ({
          ...prev,
          [sessionId]: {
            id: sessionId,
            name: t("chat.new_session"),
            messages: [],
          },
        }));
        setCurrentSessionId(sessionId);
        setTempSession(false);
      }
    },
    onChunk: (fullContent: string) => {
      // Hot path: ref write only — no setState, no re-render per chunk.
      streamTextRef.current = fullContent;
    },
    onDone: () => {
      setLoading(false);
      commitStreamedMessage("done");
    },
    onError: () => {
      setLoading(false);
      // Preserve the partial answer so the user keeps whatever arrived.
      commitStreamedMessage("error");
      toast.error(t("chat.stream_error"));
    },
  });

  const switchSession = useCallback(
    (sessionId: string) => {
      if (!sessionId) return;
      setCurrentSessionId(sessionId);
      setTempSession(sessionId === "temp");
      if (
        !(sessionId in sessions) ||
        sessions[sessionId].messages.length === 0
      ) {
        setFetchHistorySessionId(sessionId);
        setHistoryEnabled(true);
      } else {
        setCurrentMessages(sessions[sessionId].messages);
      }
    },
    [sessions],
  );

  const syncHistory = useCallback(() => {
    if (!currentSessionId || tempSession) {
      toast.warning(t("chat.select_session"));
      return;
    }
    setFetchHistorySessionId(currentSessionId);
    setHistoryEnabled(true);
  }, [currentSessionId, tempSession, t]);

  const handleStreaming = useCallback(
    async (question: string) => {
      const aiMessage: Message = {
        role: "ai",
        content: "",
        status: "streaming",
      };
      setCurrentMessages((prev) => [...prev, aiMessage]);

      try {
        await streamMutation.mutateAsync({
          question,
          model_type: selectedModel,
          session_id: currentSessionId ?? undefined,
          isNewSession: tempSession,
        });
        // Safety net: commit even if the server closed without a [DONE] sentinel.
        if (streamTextRef.current) {
          setLoading(false);
          commitStreamedMessage("done");
        }
      } catch {
        // The mutation's onError already surfaced the failure and preserved
        // the partial content — nothing more to do here.
      }
    },
    [
      tempSession,
      selectedModel,
      currentSessionId,
      streamMutation,
      commitStreamedMessage,
    ],
  );

  const handleNormal = useCallback(
    async (question: string) => {
      const thinkingMessage: Message = {
        role: "ai",
        content: "",
        status: "thinking",
      };
      setCurrentMessages((prev) => [...prev, thinkingMessage]);

      if (tempSession) {
        createSessionMutation.mutate(
          { question, model_type: selectedModel },
          {
            onSuccess: ({
              code,
              session_id: sessionId = "",
              answer,
              message,
            }) => {
              if (code === 1000) {
                const aiMessage: Message = {
                  role: "ai",
                  content: answer || "",
                };
                setSessions((prev) => ({
                  ...prev,
                  [sessionId]: {
                    id: sessionId,
                    name: t("chat.new_session"),
                    messages: [{ role: "user", content: question }, aiMessage],
                  },
                }));
                setCurrentSessionId(sessionId);
                setTempSession(false);
                setCurrentMessages([
                  { role: "user", content: question },
                  aiMessage,
                ]);
              } else {
                toast.error(message || t("chat.send_failed"));
                setCurrentMessages((prev) => prev.slice(0, -1));
              }
              setLoading(false);
            },
            onError: () => {
              toast.error(t("chat.send_failed"));
              setCurrentMessages((prev) => prev.slice(0, -1));
              setLoading(false);
            },
          },
        );
      } else {
        sendMessageMutation.mutate(
          {
            question,
            model_type: selectedModel,
            session_id: currentSessionId!,
          },
          {
            onSuccess: ({ code, answer, message }) => {
              if (code === 1000) {
                const aiMessage: Message = {
                  role: "ai",
                  content: answer || "",
                };
                setCurrentMessages((prev) => {
                  const updated = [...prev];
                  updated[updated.length - 1] = aiMessage;
                  return updated;
                });
                if (currentSessionId) {
                  setSessions((prev) => ({
                    ...prev,
                    [currentSessionId]: {
                      ...prev[currentSessionId],
                      messages: [
                        ...prev[currentSessionId].messages,
                        { role: "user", content: question },
                        aiMessage,
                      ],
                    },
                  }));
                }
              } else {
                toast.error(message || t("chat.send_failed"));
                setCurrentMessages((prev) => prev.slice(0, -1));
              }
              setLoading(false);
            },
            onError: () => {
              toast.error(t("chat.send_failed"));
              setCurrentMessages((prev) => prev.slice(0, -1));
              setLoading(false);
            },
          },
        );
      }
    },
    [
      tempSession,
      selectedModel,
      currentSessionId,
      t,
      createSessionMutation,
      sendMessageMutation,
    ],
  );

  const sendMessage = useCallback(
    async (inputMessage: string) => {
      if (!inputMessage.trim()) {
        toast.warning(t("chat.message_required"));
        return;
      }

      const userMessage: Message = {
        role: "user",
        content: inputMessage,
      };

      setCurrentMessages((prev) => [...prev, userMessage]);

      try {
        setLoading(true);
        if (isStreaming) {
          await handleStreaming(inputMessage);
        } else {
          await handleNormal(inputMessage);
        }
      } catch (err) {
        if (import.meta.env.DEV) console.error(err);
        toast.error(t("chat.send_failed"));
        setCurrentMessages((prev) => prev.slice(0, -1));
      } finally {
        if (!isStreaming) {
          setLoading(false);
        }
      }
    },
    [t, isStreaming, handleStreaming, handleNormal],
  );

  const sessionList = useMemo(() => Object.values(sessions), [sessions]);

  const { uploading, fileInputRef, triggerFileUpload, handleFileUpload } =
    useFileUpload();

  return (
    <div className="bg-background flex h-screen overflow-hidden">
      {/* Session Sidebar */}
      <SessionSidebar
        sessions={sessionList}
        currentSessionId={currentSessionId}
        onCreateNewSession={createNewSession}
        onSwitch={switchSession}
      />

      {/* Main Chat Area */}
      <main className="flex flex-1 flex-col overflow-hidden">
        {/* Header */}
        <ChatHeader
          currentSessionId={currentSessionId}
          tempSession={tempSession}
          selectedModel={selectedModel}
          isStreaming={isStreaming}
          uploading={uploading}
          fileInputRef={fileInputRef}
          onSyncHistory={syncHistory}
          onModelChange={setSelectedModel}
          onStreamingChange={setIsStreaming}
          onTriggerUpload={triggerFileUpload}
          onFileUpload={handleFileUpload}
        />

        {/* Messages Area */}
        <div className="flex-1 overflow-hidden">
          {currentMessages.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center gap-4">
              <div className="bg-primary/10 flex size-16 items-center justify-center rounded-2xl">
                <Sparkles className="text-primary size-8" />
              </div>
              <h2 className="text-2xl font-normal">{t("chat.how_can_help")}</h2>
              <p className="text-muted-foreground">
                {t("chat.start_conversation")}
              </p>
            </div>
          ) : (
            <MessageList messages={currentMessages} streamRef={streamTextRef} />
          )}
        </div>

        {/* Input Area */}
        <ChatInput loading={loading} onSend={sendMessage} />
      </main>
    </div>
  );
}

export default AiChat;
