import SettingsBar from "@/components/settings-bar";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { setTokenAtom } from "@/stores/auth";
import { useSetAtom } from "jotai";
import { LogOut, MessageSquare, Sparkles } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

function Menu() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const setToken = useSetAtom(setTokenAtom);

  const handleLogout = () => {
    setToken(null);
    toast.success(t("auth.logout_success"));
    navigate("/login");
  };

  return (
    <div className="bg-background flex min-h-screen flex-col">
      {/* Header */}
      <header className="bg-card border-border border-b">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
          <div className="flex items-center gap-3">
            <div className="bg-accent flex h-10 w-10 items-center justify-center rounded-full">
              <Sparkles className="text-primary h-5 w-5" />
            </div>
            <h1 className="text-foreground text-xl font-normal">
              {t("menu.title")}
            </h1>
          </div>
          <div className="flex items-center gap-2">
            <SettingsBar />
            <div className="bg-border mx-2 h-6 w-px" />
            <Button
              variant="ghost"
              className="text-muted-foreground hover:bg-muted gap-2"
              onClick={handleLogout}
            >
              <LogOut className="h-4 w-4" />
              {t("auth.logout")}
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex flex-1 items-center justify-center p-8">
        <div className="w-full max-w-4xl">
          {/* Welcome Section */}
          <div className="mb-12 text-center">
            <h2 className="text-foreground mb-4 text-4xl font-normal">
              {t("menu.welcome")}
            </h2>
            <p className="text-muted-foreground text-lg">
              {t("menu.select_app")}
            </p>
          </div>

          {/* Feature Cards */}
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
            {/* AI Chat Card */}
            <Card
              className="bg-card border-border hover:border-primary group cursor-pointer border transition-all duration-200 hover:shadow-lg"
              onClick={() => navigate("/ai-chat")}
            >
              <CardContent className="p-8">
                <div className="flex items-start gap-5">
                  <div className="bg-accent group-hover:bg-primary flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl transition-colors">
                    <MessageSquare className="text-primary group-hover:text-primary-foreground h-7 w-7 transition-colors" />
                  </div>
                  <div>
                    <h3 className="text-foreground group-hover:text-primary mb-2 text-lg font-medium transition-colors">
                      {t("menu.ai_chat")}
                    </h3>
                    <p className="text-muted-foreground text-sm leading-relaxed">
                      {t("menu.ai_chat_desc")}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Coming Soon Card */}
            <Card className="bg-card border-border border opacity-60">
              <CardContent className="p-8">
                <div className="flex items-start gap-5">
                  <div className="bg-muted flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl">
                    <span className="text-muted-foreground text-2xl">+</span>
                  </div>
                  <div>
                    <h3 className="text-muted-foreground mb-2 text-lg font-medium">
                      {t("menu.more_features")}
                    </h3>
                    <p className="text-muted-foreground text-sm leading-relaxed">
                      {t("menu.more_features_desc")}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-border bg-card border-t py-4">
        <div className="text-muted-foreground mx-auto flex max-w-7xl items-center justify-between px-6 text-sm">
          <span>
            © {new Date().getFullYear()} {t("menu.title")}
          </span>
          <div className="flex gap-6">
            <span className="hover:text-foreground cursor-pointer">
              {t("menu.privacy_policy")}
            </span>
            <span className="hover:text-foreground cursor-pointer">
              {t("menu.terms_of_service")}
            </span>
            <span className="hover:text-foreground cursor-pointer">
              {t("menu.help")}
            </span>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default Menu;
