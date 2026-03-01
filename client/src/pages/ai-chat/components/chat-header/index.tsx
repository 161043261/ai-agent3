import SettingsBar from "@/components/settings-bar";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MODELS, type ModelType } from "@/types";
import { ArrowLeft, Paperclip, RefreshCw } from "lucide-react";
import type { ChangeEvent, RefObject } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

interface Props {
  currentSessionId: string | null;
  tempSession: boolean;
  selectedModel: string;
  isStreaming: boolean;
  uploading: boolean;
  fileInputRef: RefObject<HTMLInputElement | null>;
  onSyncHistory: () => void;
  onModelChange: (value: ModelType) => void;
  onStreamingChange: (value: boolean) => void;
  onTriggerUpload: () => void;
  onFileUpload: (e: ChangeEvent<HTMLInputElement>) => void;
}

function ChatHeader({
  currentSessionId,
  tempSession,
  selectedModel,
  isStreaming,
  uploading,
  fileInputRef,
  onSyncHistory,
  onModelChange,
  onStreamingChange,
  onTriggerUpload,
  onFileUpload,
}: Props) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <header className="bg-card border-border flex items-center gap-4 border-b px-6 py-3">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => navigate("/menu")}
        className="text-muted-foreground hover:bg-muted gap-2 rounded-full"
      >
        <ArrowLeft className="h-4 w-4" />
        {t("common.back")}
      </Button>

      <div className="bg-border h-6 w-px" />

      <Button
        variant="ghost"
        size="sm"
        onClick={onSyncHistory}
        disabled={
          currentSessionId === null ||
          currentSessionId.length == 0 ||
          tempSession
        }
        className="text-muted-foreground hover:bg-muted gap-2 rounded-full"
      >
        <RefreshCw className="h-4 w-4" />
        {t("chat.sync_history")}
      </Button>

      <div className="ml-4 flex items-center gap-2">
        <span className="text-muted-foreground text-sm">
          {t("chat.model")}:
        </span>
        <Select value={selectedModel} onValueChange={onModelChange}>
          <SelectTrigger className="border-input h-9 w-36 rounded-md text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={MODELS.OPENAI_MODEL}>OpenAI</SelectItem>
            <SelectItem value={MODELS.OLLAMA_MODEL}>Ollama</SelectItem>
            <SelectItem value={MODELS.OPENAI_RAG_MODEL}>
              OpenAI with RAG
            </SelectItem>
            <SelectItem value={MODELS.OLLAMA_RAG_MODEL}>
              Ollama with RAG
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="ml-4 flex items-center gap-2">
        <Checkbox
          id="streaming"
          checked={isStreaming}
          onCheckedChange={(checked) => onStreamingChange(!!checked)}
          className="border-input data-[state=checked]:bg-primary data-[state=checked]:border-primary"
        />
        <label
          htmlFor="streaming"
          className="text-muted-foreground cursor-pointer text-sm"
        >
          {t("chat.streaming")}
        </label>
      </div>

      <Button
        variant="ghost"
        size="sm"
        onClick={onTriggerUpload}
        disabled={uploading}
        className="text-muted-foreground hover:bg-muted ml-auto gap-2 rounded-full"
      >
        <Paperclip className="h-4 w-4" />
        {t("chat.upload_doc")}
      </Button>
      <input
        ref={fileInputRef}
        type="file"
        accept=".md,.txt"
        className="hidden"
        onChange={onFileUpload}
      />

      <div className="bg-border h-6 w-px" />
      <SettingsBar />
    </header>
  );
}

export default ChatHeader;
