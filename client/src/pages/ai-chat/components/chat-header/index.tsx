import SettingsBar from "@/components/settings-bar";
import { MODELS, type ModelType } from "@/types";
import { ArrowLeft, Paperclip, RefreshCw } from "lucide-react";
import type { ChangeEvent, RefObject } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";

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
    <header className="bg-background/70 flex items-center gap-4 border-b px-6 py-3 backdrop-blur-md">
      <Button
        variant="ghost"
        size="sm"
        className="rounded-full"
        onClick={() => navigate("/menu")}
      >
        <ArrowLeft data-icon="inline-start" />
        {t("common.back")}
      </Button>

      <Separator orientation="vertical" className="h-6" />

      <Button
        variant="ghost"
        size="sm"
        className="rounded-full"
        onClick={onSyncHistory}
        disabled={
          currentSessionId === null ||
          currentSessionId.length === 0 ||
          tempSession
        }
      >
        <RefreshCw data-icon="inline-start" />
        {t("chat.sync_history")}
      </Button>

      <div className="ml-4 flex items-center gap-2">
        <span className="text-muted-foreground text-sm">
          {t("chat.model")}:
        </span>
        <Select
          value={selectedModel}
          onValueChange={(value) => onModelChange(value as ModelType)}
        >
          <SelectTrigger className="w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value={MODELS.OPENAI_MODEL}>OpenAI</SelectItem>
              <SelectItem value={MODELS.OPENAI_RAG_MODEL}>
                OpenAI with RAG
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <div className="ml-4 flex items-center gap-2">
        <Checkbox
          id="streaming"
          checked={isStreaming}
          onCheckedChange={(checked) => onStreamingChange(checked === true)}
        />
        <Label htmlFor="streaming" className="cursor-pointer font-normal">
          {t("chat.streaming")}
        </Label>
      </div>

      <Button
        variant="ghost"
        size="sm"
        className="ml-auto rounded-full"
        onClick={onTriggerUpload}
        disabled={uploading}
      >
        <Paperclip data-icon="inline-start" />
        {t("chat.upload_doc")}
      </Button>
      <input
        ref={fileInputRef}
        type="file"
        accept=".md,.txt"
        className="hidden"
        onChange={onFileUpload}
      />

      <Separator orientation="vertical" className="h-6" />
      <SettingsBar />
    </header>
  );
}

export default ChatHeader;
