import { useState, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Send } from "lucide-react";

interface Props {
  loading: boolean;
  onSend: (message: string) => void;
}

function ChatInput({ loading, onSend }: Props) {
  const { t } = useTranslation();
  const [inputMessage, setInputMessage] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleSend = () => {
    if (!inputMessage.trim() || loading) return;
    onSend(inputMessage);
    setInputMessage("");
  };

  return (
    <div className="bg-card border-t p-4">
      <div className="mx-auto max-w-3xl">
        <div className="bg-muted focus-within:border-primary focus-within:bg-background relative flex items-center rounded-full border border-transparent transition-all">
          <textarea
            ref={textareaRef}
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t("chat.input_placeholder")}
            disabled={loading}
            rows={1}
            className="text-foreground placeholder-muted-foreground max-h-30 flex-1 resize-none bg-transparent px-6 py-3 focus:outline-none disabled:opacity-50"
          />
          <Button
            onClick={handleSend}
            disabled={inputMessage.trim().length === 0 || loading}
            size="sm"
            className="bg-primary hover:bg-primary/90 disabled:bg-muted mr-2 h-10 w-10 rounded-full p-0 disabled:opacity-100"
          >
            <Send className="text-primary-foreground h-5 w-5" />
          </Button>
        </div>
        <p className="text-muted-foreground mt-2 text-center text-xs">
          {t("chat.send_hint")}
        </p>
      </div>
    </div>
  );
}

export default ChatInput;
