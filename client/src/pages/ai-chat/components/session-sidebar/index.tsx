import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus } from "lucide-react";
import type { Session } from "@/types";

interface Props {
  sessions: Session[];
  currentSessionId: string | null;
  onCreateNewSession: () => void;
  onSwitch: (sessionId: string) => void;
}

function SessionSidebar({
  sessions,
  currentSessionId,
  onCreateNewSession,
  onSwitch,
}: Props) {
  const { t } = useTranslation();

  return (
    <aside className="bg-card border-border flex w-72 flex-col border-r">
      <div className="border-border border-b p-4">
        <Button
          onClick={onCreateNewSession}
          className="bg-card hover:bg-muted text-foreground border-border h-10 w-full gap-2 rounded-full border shadow-sm"
        >
          <Plus className="text-primary h-5 w-5" />
          {t("chat.new_chat")}
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-2">
          {sessions.map((session) => (
            <button
              key={session.id}
              onClick={() => onSwitch(session.id)}
              className={`mb-1 w-full rounded-full px-4 py-3 text-left text-sm transition-colors ${
                currentSessionId === session.id
                  ? "bg-accent text-accent-foreground font-medium"
                  : "hover:bg-muted text-foreground"
              }`}
            >
              {session.name}
            </button>
          ))}
        </div>
      </ScrollArea>
    </aside>
  );
}

export default SessionSidebar;
